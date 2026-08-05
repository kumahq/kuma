package bootstrap

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	bootstrap_config "github.com/kumahq/kuma/v3/pkg/config/xds/bootstrap"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/xds/bootstrap/types"
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, KumaDpBootstrap, error)
}

func NewDefaultBootstrapGenerator(
	resManager core_manager.ResourceManager,
	serverConfig *bootstrap_config.BootstrapServerConfig,
	dpServerCertFile string,
	authEnabledForProxyType map[string]bool,
	enableReloadableTokens bool,
	hdsEnabled bool,
	defaultAdminPort uint32,
	envoyAdminUnixSocket bool,
) (BootstrapGenerator, error) {
	cert, err := loadCertFromFile(dpServerCertFile)
	if err != nil {
		return nil, err
	}
	if serverConfig.Params.XdsHost != "" {
		if err := cert.parsed.VerifyHostname(serverConfig.Params.XdsHost); err != nil {
			return nil, errors.Errorf("hostname: %s set by KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST is not available in the DP Server certificate. Available hostnames: %q. Change the hostname or generate certificate with proper hostname.", serverConfig.Params.XdsHost, cert.parsed.DNSNames)
		}
	}
	return &bootstrapGenerator{
		resManager:              resManager,
		config:                  serverConfig,
		xdsCertFile:             dpServerCertFile,
		authEnabledForProxyType: authEnabledForProxyType,
		enableReloadableTokens:  enableReloadableTokens,
		hdsEnabled:              hdsEnabled,
		defaultAdminPort:        defaultAdminPort,
		envoyAdminUnixSocket:    envoyAdminUnixSocket,
	}, nil
}

type bootstrapGenerator struct {
	resManager              core_manager.ResourceManager
	config                  *bootstrap_config.BootstrapServerConfig
	authEnabledForProxyType map[string]bool
	enableReloadableTokens  bool
	xdsCertFile             string
	hdsEnabled              bool
	defaultAdminPort        uint32
	envoyAdminUnixSocket    bool
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, KumaDpBootstrap, error) {
	if request.ProxyType == "" {
		request.ProxyType = string(mesh_proto.DataplaneProxyType)
	}
	kumaDpBootstrap := KumaDpBootstrap{}
	// The DP server reloads its certificate when it is rotated on disk, so it
	// cannot be parsed once at startup. Read it at most once per request
	// instead, both the SAN check and the CA handed to the proxy need it.
	loadCert := sync.OnceValues(b.loadDpServerCert)
	if err := b.validateRequest(request, loadCert); err != nil {
		return nil, kumaDpBootstrap, err
	}
	features := make(xds_types.Features, len(request.Features))
	for _, feature := range request.Features {
		features[feature] = true
	}

	proxyId := core_xds.BuildProxyId(request.Mesh, request.Name)
	params := configParameters{
		Id:                            proxyId.String(),
		AdminAddress:                  b.config.Params.AdminAddress,
		AdminAccessLogPath:            b.adminAccessLogPath(request.OperatingSystem),
		XdsHost:                       b.xdsHost(request),
		XdsPort:                       b.config.Params.XdsPort,
		XdsConnectTimeout:             b.config.Params.XdsConnectTimeout.Duration,
		XdsGrpcMaxReceiveMessageBytes: b.config.Params.XdsGrpcMaxReceiveMessageBytes,
		DataplaneToken:                request.DataplaneToken,
		DataplaneTokenPath:            request.DataplaneTokenPath,
		DataplaneResource:             request.DataplaneResource,
		Version: &mesh_proto.Version{
			KumaDp: &mesh_proto.KumaDpVersion{
				Version:   request.Version.KumaDp.Version,
				GitTag:    request.Version.KumaDp.GitTag,
				GitCommit: request.Version.KumaDp.GitCommit,
				BuildDate: request.Version.KumaDp.BuildDate,
			},
			Envoy: &mesh_proto.EnvoyVersion{
				Version:          request.Version.Envoy.Version,
				Build:            request.Version.Envoy.Build,
				KumaDpCompatible: request.Version.Envoy.KumaDpCompatible,
			},
		},
		DynamicMetadata:      request.DynamicMetadata,
		DNSPort:              request.DNSPort,
		ReadinessPort:        request.ReadinessPort,
		AppProbeProxyEnabled: request.AppProbeProxyEnabled,
		ProxyType:            request.ProxyType,
		Features:             features,
		Resources:            request.Resources,
		Workdir:              request.Workdir,
		MetricsCertPath:      request.MetricsResources.CertPath,
		MetricsKeyPath:       request.MetricsResources.KeyPath,
		SystemCaPath:         request.SystemCaPath,
		TransparentProxy:     request.TransparentProxy,
		IPv6Enabled:          request.IPv6Enabled,
		SpireSocketPath:      request.SpireResources.SocketPath,
		OtelEnvInventory:     request.OtelEnv,
	}

	setAdminPort := func(adminPortFromResource uint32) {
		if adminPortFromResource != 0 {
			params.AdminPort = adminPortFromResource
		} else {
			params.AdminPort = b.defaultAdminPort
		}
	}

	if b.envoyAdminUnixSocket {
		if request.Workdir != "" {
			params.AdminSocketPath = core_xds.AdminSocketName(request.Workdir)
		} else {
			log.Info("[WARNING] admin unix domain socket enabled but workdir is empty, falling back to TCP admin", "proxyId", proxyId.String())
		}
	}

	meshResource := core_mesh.NewMeshResource()
	switch mesh_proto.ProxyType(params.ProxyType) {
	case mesh_proto.DataplaneProxyType, "":
		params.HdsEnabled = b.hdsEnabled
		dataplane, err := b.dataplaneFor(ctx, request, proxyId)
		if err != nil {
			return nil, kumaDpBootstrap, err
		}

		kumaDpBootstrap.NetworkingConfig.Address = dataplane.Spec.GetNetworking().GetAddress()
		params.Service = dataplane.IdentifyingName()
		setAdminPort(dataplane.Spec.GetNetworking().GetAdmin().GetPort())

		err = b.resManager.Get(ctx, meshResource, core_store.GetByKey(dataplane.Meta.GetMesh(), core_model.NoMesh))
		if err != nil {
			return nil, kumaDpBootstrap, err
		}

	default:
		return nil, kumaDpBootstrap, errors.Errorf("unknown proxy type %v", params.ProxyType)
	}
	var err error
	if params.CertBytes, err = b.caCert(request, loadCert); err != nil {
		return nil, kumaDpBootstrap, err
	}

	config, err := genConfig(params, b.enableReloadableTokens, meshResource)
	if err != nil {
		return nil, kumaDpBootstrap, errors.Wrap(err, "failed creating bootstrap conf")
	}
	if err = config.Validate(); err != nil {
		return nil, kumaDpBootstrap, errors.Wrap(err, "Envoy bootstrap config is not valid")
	}
	return config, kumaDpBootstrap, nil
}

var DpTokenRequired = errors.New("Dataplane Token is required. Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP")

var NotCA = errors.New("A data plane proxy is trying to verify the control plane using the certificate which is not a certificate authority (basic constraint 'CA' is set to 'false').\n" +
	"Provide CA that was used to sign a certificate used in the control plane by using 'kuma-dp run --ca-cert-file=file' or via KUMA_CONTROL_PLANE_CA_CERT_FILE")

func SANMismatchErr(host string, cert *x509.Certificate) error {
	var sans []string
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	sans = append(sans, cert.DNSNames...)
	return errors.Errorf("A data plane proxy is trying to connect to the control plane using %q address, but the certificate in the control plane has the following SANs %q. "+
		"Either change the --cp-address in kuma-dp to one of those or execute the following steps:\n"+
		"1) Generate a new certificate with the address you are trying to use. It is recommended to use trusted Certificate Authority, but you can also generate self-signed certificates using 'kumactl generate tls-certificate --type=server --hostname=%s'\n"+
		"2) Set KUMA_GENERAL_TLS_CERT_FILE and KUMA_GENERAL_TLS_KEY_FILE or the equivalent in Kuma CP config file to the new certificate.\n"+
		"3) Start kuma-dp, the control plane picks up the new certificate on its own.", host, sans, host)
}

func ISSANMismatchErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "A data plane proxy is trying to connect to the control plane using")
}

func (b *bootstrapGenerator) validateRequest(request types.BootstrapRequest, loadCert func() (*dpServerCert, error)) error {
	if b.authEnabledForProxyType[request.ProxyType] && request.DataplaneToken == "" && request.DataplaneTokenPath == "" {
		return DpTokenRequired
	}
	if b.config.Params.XdsHost == "" { // XdsHost takes precedence over Host in the request, so validate only when it is not set
		cert, err := loadCert()
		if err != nil {
			return err
		}
		if err := cert.parsed.VerifyHostname(request.Host); err != nil {
			return SANMismatchErr(request.Host, cert.parsed)
		}
	}
	return nil
}

// dataplaneFor returns dataplane for two flows
// 1) Dataplane is passed to kuma-dp run, in this case we just read DP from the BootstrapRequest
// 2) Dataplane is created before kuma-dp run, in this case we access storage to fetch it (ex. Kubernetes)
func (b *bootstrapGenerator) dataplaneFor(ctx context.Context, request types.BootstrapRequest, proxyId *core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
	if request.DataplaneResource != "" {
		res, err := rest.YAML.UnmarshalCore([]byte(request.DataplaneResource))
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
		if request.TransparentProxy.Enabled() && len(dp.Spec.GetNetworking().GetOutbound()) > 0 {
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
	if err := b.resManager.Get(ctx, core_mesh.NewMeshResource(), core_store.GetByKey(mesh, core_model.NoMesh)); err != nil {
		if core_store.IsNotFound(err) {
			verr := validators.ValidationError{}
			verr.AddViolation("mesh", fmt.Sprintf("mesh %q does not exist", mesh))
			return verr.OrNil()
		}
		return err
	}
	return nil
}

// caCert gets CA cert that was used to signed cert that DP server is protected with.
// Technically result of this function does not have to be a valid CA.
// When user provides custom cert + key and does not provide --ca-cert-file to kuma-dp run, this can return just a regular cert
func (b *bootstrapGenerator) caCert(request types.BootstrapRequest, loadCert func() (*dpServerCert, error)) ([]byte, error) {
	// CaCert from the request takes precedence. It is only visible if user provides --ca-cert-file to kuma-dp run
	var cert *dpServerCert
	switch {
	case request.CaCert != "":
		parsed, err := parseCert([]byte(request.CaCert))
		if err != nil {
			return nil, errors.Wrap(err, "could not parse certificate from request .CaCert")
		}
		cert = &dpServerCert{pem: []byte(request.CaCert), parsed: parsed}
	case b.xdsCertFile != "":
		var err error
		if cert, err = loadCert(); err != nil {
			return nil, errors.Wrapf(err, "failed getting cert from file %s", b.xdsCertFile)
		}
	default:
		return nil, nil
	}
	// checking just IsCA is not enough, because it's valid to generate CA without CA:TRUE basic constraint
	if cert.parsed.BasicConstraintsValid && !cert.parsed.IsCA {
		return nil, NotCA
	}
	return cert.pem, nil
}

func (b *bootstrapGenerator) xdsHost(request types.BootstrapRequest) string {
	if b.config.Params.XdsHost != "" { // XdsHost from config takes precedence over Host from request
		return b.config.Params.XdsHost
	} else {
		return request.Host
	}
}

func (b *bootstrapGenerator) adminAccessLogPath(operatingSystem string) string {
	if operatingSystem == "" { // backwards compatibility
		return b.config.Params.AdminAccessLogPath
	}
	return b.config.Params.AdminAccessLogPath
}

// dpServerCert is the certificate the DP server presents, as read from disk.
// Both representations are kept because the CA is handed to the proxy as PEM
// while the SAN and CA checks need the parsed form.
type dpServerCert struct {
	pem    []byte
	parsed *x509.Certificate
}

func (b *bootstrapGenerator) loadDpServerCert() (*dpServerCert, error) {
	return loadCertFromFile(b.xdsCertFile)
}

func loadCertFromFile(dpServerCertFile string) (*dpServerCert, error) {
	certBytes, err := os.ReadFile(filepath.Clean(dpServerCertFile))
	if err != nil {
		return nil, errors.Wrap(err, "could not read certificate")
	}
	cert, err := parseCert(certBytes)
	if err != nil {
		return nil, err
	}
	return &dpServerCert{pem: certBytes, parsed: cert}, nil
}

func parseCert(certBytes []byte) (*x509.Certificate, error) {
	pemCert, _ := pem.Decode(certBytes)
	if pemCert == nil {
		return nil, errors.New("could not parse certificate")
	}
	cert, err := x509.ParseCertificate(pemCert.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse certificate")
	}
	return cert, nil
}
