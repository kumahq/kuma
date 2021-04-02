package bootstrap

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"text/template"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/validators"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"

	envoy_bootstrap_v2 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	envoy_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
	_ "github.com/kumahq/kuma/pkg/xds/envoy" // import Envoy protobuf definitions so (un)marshalling Envoy protobuf works in tests (normally it is imported in root.go)
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, types.BootstrapVersion, error)
}

func NewDefaultBootstrapGenerator(
	resManager core_manager.ResourceManager,
	config *bootstrap_config.BootstrapServerConfig,
	dpServerCertFile string,
	dpAuthEnabled bool,
	hdsEnabled bool,
) (BootstrapGenerator, error) {
	hostsAndIps, err := hostsAndIPsFromCertFile(dpServerCertFile)
	if err != nil {
		return nil, err
	}
	if config.Params.XdsHost != "" && !hostsAndIps[config.Params.XdsHost] {
		return nil, errors.Errorf("hostname: %s set by KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST is not available in the DP Server certificate. Available hostnames: %q. Change the hostname or generate certificate with proper hostname.", config.Params.XdsHost, hostsAndIps.slice())
	}
	return &bootstrapGenerator{
		resManager:    resManager,
		config:        config,
		xdsCertFile:   dpServerCertFile,
		dpAuthEnabled: dpAuthEnabled,
		hostsAndIps:   hostsAndIps,
		hdsEnabled:    hdsEnabled,
	}, nil
}

type bootstrapGenerator struct {
	resManager    core_manager.ResourceManager
	config        *bootstrap_config.BootstrapServerConfig
	dpAuthEnabled bool
	xdsCertFile   string
	hostsAndIps   SANSet
	hdsEnabled    bool
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, types.BootstrapVersion, error) {
	if err := b.validateRequest(request); err != nil {
		return nil, "", err
	}

	proxyId, err := core_xds.BuildProxyId(request.Mesh, request.Name)
	if err != nil {
		return nil, "", err
	}

	dataplane, err := b.dataplaneFor(ctx, request, proxyId)
	if err != nil {
		return nil, "", err
	}

	return b.generateFor(*proxyId, dataplane, request)
}

var DpTokenRequired = errors.New("Dataplane Token is required. Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP")

var InvalidBootstrapVersion = errors.New(`Invalid BootstrapVersion. Available values are: "2", "3"`)

var NotCA = errors.New("A data plane proxy is trying to verify the control plane using the certificate which is not a certificate authority (basic constraint 'CA' is set to 'false').\n" +
	"Provide CA that was used to sign a certificate used in the control plane by using 'kuma-dp run --ca-cert-file=file' or via KUMA_CONTROL_PLANE_CA_CERT_FILE")

func SANMismatchErr(host string, sans []string) error {
	return errors.Errorf("A data plane proxy is trying to connect to the control plane using %q address, but the certificate in the control plane has the following SANs %q. "+
		"Either change the --cp-address in kuma-dp to one of those or execute the following steps:\n"+
		"1) Generate a new certificate with the address you are trying to use. It is recommended to use trusted Certificate Authority, but you can also generate self-signed certificates using 'kumactl generate tls-certificate --type=server --cp-hostname=%s'\n"+
		"2) Set KUMA_GENERAL_TLS_CERT_FILE and KUMA_GENERAL_TLS_KEY_FILE or the equivalent in Kuma CP config file to the new certificate.\n"+
		"3) Restart the control plane to read the new certificate and start kuma-dp.", host, sans, host)
}

func ISSANMismatchErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "A data plane proxy is trying to connect to the control plane using")
}

func (b *bootstrapGenerator) validateRequest(request types.BootstrapRequest) error {
	if b.dpAuthEnabled && request.DataplaneTokenPath == "" && request.DataplaneToken == "" {
		return DpTokenRequired
	}
	if b.config.Params.XdsHost == "" { // XdsHost takes precedence over Host in the request, so validate only when it is not set
		if !b.hostsAndIps[request.Host] {
			return SANMismatchErr(request.Host, b.hostsAndIps.slice())
		}
	}
	if request.DataplaneToken != "" && request.DataplaneTokenPath != "" {
		verr := validators.ValidationError{}
		verr.AddViolation("dataplaneToken", "only one of dataplaneToken and dataplaneTokenField can be defined")
		return verr.OrNil()
	}
	return nil
}

// dataplaneFor returns dataplane for two flows
// 1) Dataplane is passed to kuma-dp run, in this case we just read DP from the BootstrapRequest
// 2) Dataplane is created before kuma-dp run, in this case we access storage to fetch it (ex. Kubernetes)
func (b *bootstrapGenerator) dataplaneFor(ctx context.Context, request types.BootstrapRequest, proxyId *core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
	if request.DataplaneResource != "" {
		res, err := rest.UnmarshallToCore([]byte(request.DataplaneResource))
		if err != nil {
			return nil, err
		}
		dp, ok := res.(*core_mesh.DataplaneResource)
		if !ok {
			return nil, errors.Errorf("invalid resource")
		}
		if err := dp.Validate(); err != nil {
			return nil, err
		}
		// this part of validation works only for Universal scenarios with TransparentProxying
		if dp.Spec.Networking.TransparentProxying != nil && len(dp.Spec.Networking.Outbound) != 0 {
			var err validators.ValidationError
			err.AddViolation("outbound", "should be empty since dataplane is in Transparent Proxying mode")
			return nil, err.OrNil()
		}
		if err := b.validateMeshExist(ctx, dp.Meta.GetMesh()); err != nil {
			return nil, err
		}
		return dp, nil
	} else {
		dataplane := core_mesh.NewDataplaneResource()
		if err := b.resManager.Get(ctx, dataplane, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
			return nil, err
		}
		return dataplane, nil
	}
}

func (b *bootstrapGenerator) validateMeshExist(ctx context.Context, mesh string) error {
	if err := b.resManager.Get(ctx, core_mesh.NewMeshResource(), core_store.GetByKey(mesh, model.NoMesh)); err != nil {
		if core_store.IsResourceNotFound(err) {
			verr := validators.ValidationError{}
			verr.AddViolation("mesh", fmt.Sprintf("mesh %q does not exist", mesh))
			return verr.OrNil()
		}
		return err
	}
	return nil
}

func (b *bootstrapGenerator) generateFor(proxyId core_xds.ProxyId, dataplane *core_mesh.DataplaneResource, request types.BootstrapRequest) (proto.Message, types.BootstrapVersion, error) {
	// if dataplane has no service - fill this with placeholder. Otherwise take the first service
	service := dataplane.Spec.GetIdentifyingService()

	adminPort := b.config.Params.AdminPort
	if request.AdminPort != 0 {
		adminPort = request.AdminPort
	}

	if err := b.verifyAdminPort(adminPort, dataplane); err != nil {
		return nil, "", err
	}

	cert, origin, err := b.caCert(request)
	if err != nil {
		return nil, "", err
	}

	if err := b.validateCaCert(cert, origin, request); err != nil {
		return nil, "", err
	}

	accessLogSocket := envoy_common.AccessLogSocketName(request.Name, request.Mesh)
	params := configParameters{
		Id:                 proxyId.String(),
		Service:            service,
		AdminAddress:       b.config.Params.AdminAddress,
		AdminPort:          adminPort,
		AdminAccessLogPath: b.config.Params.AdminAccessLogPath,
		XdsHost:            b.xdsHost(request),
		XdsPort:            b.config.Params.XdsPort,
		XdsConnectTimeout:  b.config.Params.XdsConnectTimeout,
		AccessLogPipe:      accessLogSocket,
		DataplaneTokenPath: request.DataplaneTokenPath,
		DataplaneToken:     request.DataplaneToken,
		DataplaneResource:  request.DataplaneResource,
		CertBytes:          base64.StdEncoding.EncodeToString(cert),
		KumaDpVersion:      request.Version.KumaDp.Version,
		KumaDpGitTag:       request.Version.KumaDp.GitTag,
		KumaDpGitCommit:    request.Version.KumaDp.GitCommit,
		KumaDpBuildDate:    request.Version.KumaDp.BuildDate,
		EnvoyVersion:       request.Version.Envoy.Version,
		EnvoyBuild:         request.Version.Envoy.Build,
		HdsEnabled:         b.hdsEnabled,
		DynamicMetadata:    request.DynamicMetadata,
	}
	log.WithValues("params", params).Info("Generating bootstrap config")
	return b.configForParameters(params, request.BootstrapVersion)
}

func (b *bootstrapGenerator) validateCaCert(cert []byte, origin string, request types.BootstrapRequest) error {
	if request.DataplaneTokenPath != "" {
		// when using GoogleGRPC it is valid to put non-CA certificate therefore we should only verify this for EnvoyGRPC
		// EnvoyGRPC is used when DataplaneToken is passed via request as inline value, not as a file.
		return nil
	}

	pemCert, _ := pem.Decode(cert)
	if pemCert == nil {
		return errors.New("could not parse certificate from " + origin)
	}
	x509Cert, err := x509.ParseCertificate(pemCert.Bytes)
	if err != nil {
		return errors.Wrap(err, "could not parse certificate from "+origin)
	}
	// checking just x509Cert.IsCA is not enough, because it's valid to generate CA without CA:TRUE basic constraint
	if x509Cert.BasicConstraintsValid && !x509Cert.IsCA {
		return NotCA
	}
	return nil
}

// caCert gets CA cert that was used to signed cert that DP server is protected with.
// Technically result of this function does not have to be a valid CA.
// When user provides custom cert + key and does not provide --ca-cert-file to kuma-dp run, this can return just a regular cert
func (b *bootstrapGenerator) caCert(request types.BootstrapRequest) ([]byte, string, error) {
	// CaCert from the request takes precedence. It is only visible if user provides --ca-cert-file to kuma-dp run
	if request.CaCert != "" {
		return []byte(request.CaCert), "request .CaCert", nil
	}
	if b.xdsCertFile != "" {
		file, err := ioutil.ReadFile(b.xdsCertFile)
		if err != nil {
			return nil, "file " + b.xdsCertFile, err
		}
		return file, "", nil
	}
	return nil, "", nil
}

func (b *bootstrapGenerator) xdsHost(request types.BootstrapRequest) string {
	if b.config.Params.XdsHost != "" { // XdsHost from config takes precedence over Host from request
		return b.config.Params.XdsHost
	} else {
		return request.Host
	}
}

func (b *bootstrapGenerator) verifyAdminPort(adminPort uint32, dataplane *core_mesh.DataplaneResource) error {
	// The admin port in kuma-dp is always bound to 127.0.0.1
	if dataplane.UsesInboundInterface(core_mesh.IPv4Loopback, adminPort) {
		return errors.Errorf("Resource precondition failed: Port %d requested as both admin and inbound port.", adminPort)
	}

	if dataplane.UsesOutboundInterface(core_mesh.IPv4Loopback, adminPort) {
		return errors.Errorf("Resource precondition failed: Port %d requested as both admin and outbound port.", adminPort)
	}
	return nil
}

func (b *bootstrapGenerator) configForParameters(params configParameters, version types.BootstrapVersion) (proto.Message, types.BootstrapVersion, error) {
	switch {
	// V2
	case version == types.BootstrapV2:
		fallthrough
	case version == "" && b.config.APIVersion == envoy_common.APIV2: // if client did not overridden bootstrap version, provide bootstrap based on Kuma CP config
		cfg, err := b.configForParametersV2(params)
		if err != nil {
			return nil, "", err
		}
		return cfg, types.BootstrapV2, nil
	// V3
	case version == types.BootstrapV3:
		fallthrough
	case version == "" && b.config.APIVersion == envoy_common.APIV3: // if client did not overridden bootstrap version, provide bootstrap based on Kuma CP config
		cfg, err := b.configForParametersV3(params)
		if err != nil {
			return nil, "", err
		}
		return cfg, types.BootstrapV3, nil
	default:
		return nil, "", InvalidBootstrapVersion
	}
}

func (b *bootstrapGenerator) configForParametersV2(params configParameters) (proto.Message, error) {
	tmpl, err := template.New("bootstrap").Parse(configTemplateV2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config template")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, errors.Wrap(err, "failed to render config template")
	}
	config := &envoy_bootstrap_v2.Bootstrap{}
	if err := util_proto.FromYAML(buf.Bytes(), config); err != nil {
		return nil, errors.Wrap(err, "failed to parse bootstrap config")
	}
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "Envoy bootstrap config is not valid")
	}
	return config, nil
}

func (b *bootstrapGenerator) configForParametersV3(params configParameters) (proto.Message, error) {
	tmpl, err := template.New("bootstrap").Parse(configTemplateV3)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config template")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, errors.Wrap(err, "failed to render config template")
	}
	config := &envoy_bootstrap_v3.Bootstrap{}
	if err := util_proto.FromYAML(buf.Bytes(), config); err != nil {
		return nil, errors.Wrap(err, "failed to parse bootstrap config")
	}
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "Envoy bootstrap config is not valid")
	}
	return config, nil
}

type SANSet map[string]bool

func (s SANSet) slice() []string {
	sans := []string{}
	for san := range s {
		sans = append(sans, san)
	}
	sort.Strings(sans)
	return sans
}

func hostsAndIPsFromCertFile(dpServerCertFile string) (SANSet, error) {
	certBytes, err := ioutil.ReadFile(dpServerCertFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not read certificate")
	}
	pemCert, _ := pem.Decode(certBytes)
	if pemCert == nil {
		return nil, errors.New("could not parse certificate")
	}
	cert, err := x509.ParseCertificate(pemCert.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse certificate")
	}

	hostsAndIps := map[string]bool{}
	for _, dnsName := range cert.DNSNames {
		hostsAndIps[dnsName] = true
	}
	for _, ip := range cert.IPAddresses {
		hostsAndIps[ip.String()] = true
	}
	return hostsAndIps, nil
}
