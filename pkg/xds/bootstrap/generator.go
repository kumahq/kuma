package bootstrap

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	xds_config "github.com/kumahq/kuma/pkg/config/xds"
	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, KumaDpBootstrap, error)
}

func NewDefaultBootstrapGenerator(
	resManager core_manager.ResourceManager,
	serverConfig *bootstrap_config.BootstrapServerConfig,
	proxyConfig xds_config.Proxy,
	dpServerCertFile string,
	authEnabledForProxyType map[string]bool,
	enableReloadableTokens bool,
	hdsEnabled bool,
	defaultAdminPort uint32,
	deltaXdsEnabled bool,
) (BootstrapGenerator, error) {
	hostsAndIps, err := hostsAndIPsFromCertFile(dpServerCertFile)
	if err != nil {
		return nil, err
	}
	if serverConfig.Params.XdsHost != "" && !hostsAndIps[serverConfig.Params.XdsHost] {
		return nil, errors.Errorf("hostname: %s set by KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST is not available in the DP Server certificate. Available hostnames: %q. Change the hostname or generate certificate with proper hostname.", serverConfig.Params.XdsHost, hostsAndIps.slice())
	}
	return &bootstrapGenerator{
		resManager:              resManager,
		config:                  serverConfig,
		proxyConfig:             proxyConfig,
		xdsCertFile:             dpServerCertFile,
		authEnabledForProxyType: authEnabledForProxyType,
		enableReloadableTokens:  enableReloadableTokens,
		hostsAndIps:             hostsAndIps,
		hdsEnabled:              hdsEnabled,
		defaultAdminPort:        defaultAdminPort,
		deltaXdsEnabled:         deltaXdsEnabled,
	}, nil
}

type bootstrapGenerator struct {
	resManager              core_manager.ResourceManager
	config                  *bootstrap_config.BootstrapServerConfig
	proxyConfig             xds_config.Proxy
	authEnabledForProxyType map[string]bool
	enableReloadableTokens  bool
	xdsCertFile             string
	hostsAndIps             SANSet
	hdsEnabled              bool
	defaultAdminPort        uint32
	deltaXdsEnabled         bool
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, KumaDpBootstrap, error) {
	if request.ProxyType == "" {
		request.ProxyType = string(mesh_proto.DataplaneProxyType)
	}
	kumaDpBootstrap := KumaDpBootstrap{}
	if err := b.validateRequest(request); err != nil {
		return nil, kumaDpBootstrap, err
	}

	proxyId := core_xds.BuildProxyId(request.Mesh, request.Name)
	params := configParameters{
		Id:                 proxyId.String(),
		AdminAddress:       b.config.Params.AdminAddress,
		AdminAccessLogPath: b.adminAccessLogPath(request.OperatingSystem),
		XdsHost:            b.xdsHost(request),
		XdsPort:            b.config.Params.XdsPort,
		XdsConnectTimeout:  b.config.Params.XdsConnectTimeout.Duration,
		DataplaneToken:     request.DataplaneToken,
		DataplaneTokenPath: request.DataplaneTokenPath,
		DataplaneResource:  request.DataplaneResource,
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
		Features:             request.Features,
		Resources:            request.Resources,
		Workdir:              request.Workdir,
		MetricsCertPath:      request.MetricsResources.CertPath,
		MetricsKeyPath:       request.MetricsResources.KeyPath,
		SystemCaPath:         request.SystemCaPath,
	}

	setAdminPort := func(adminPortFromResource uint32) {
		if adminPortFromResource != 0 {
			params.AdminPort = adminPortFromResource
		} else {
			params.AdminPort = b.defaultAdminPort
		}
	}
	setXdsTransportProtocolVariant := func(resourceMode mesh_proto.EnvoyConfiguration_XdsTransportProtocolVariant) {
		switch resourceMode {
		case mesh_proto.EnvoyConfiguration_DEFAULT:
			params.XdsTransportProtocolVariant = types.GRPC
			if b.deltaXdsEnabled {
				params.XdsTransportProtocolVariant = types.DELTA_GRPC
			}
		case mesh_proto.EnvoyConfiguration_DELTA_GRPC:
			params.XdsTransportProtocolVariant = types.DELTA_GRPC
		case mesh_proto.EnvoyConfiguration_GRPC:
			params.XdsTransportProtocolVariant = types.GRPC
		}
	}

	switch mesh_proto.ProxyType(params.ProxyType) {
	case mesh_proto.IngressProxyType:
		zoneIngress, err := b.zoneIngressFor(ctx, request, proxyId)
		if err != nil {
			return nil, kumaDpBootstrap, err
		}

		params.Service = "ingress"
		setAdminPort(zoneIngress.Spec.GetNetworking().GetAdmin().GetPort())
		setXdsTransportProtocolVariant(zoneIngress.Spec.GetEnvoy().GetXdsTransportProtocolVariant())
	case mesh_proto.EgressProxyType:
		zoneEgress, err := b.zoneEgressFor(ctx, request, proxyId)
		if err != nil {
			return nil, kumaDpBootstrap, err
		}
		params.Service = "egress"
		setAdminPort(zoneEgress.Spec.GetNetworking().GetAdmin().GetPort())
		setXdsTransportProtocolVariant(zoneEgress.Spec.GetEnvoy().GetXdsTransportProtocolVariant())
	case mesh_proto.DataplaneProxyType, "":
		params.HdsEnabled = b.hdsEnabled
		dataplane, err := b.dataplaneFor(ctx, request, proxyId)
		if err != nil {
			return nil, kumaDpBootstrap, err
		}

		if dataplane.Spec.IsBuiltinGateway() {
			params.IsGatewayDataplane = true
		}
		kumaDpBootstrap.NetworkingConfig.IsUsingTransparentProxy = dataplane.IsUsingTransparentProxy()
		kumaDpBootstrap.NetworkingConfig.Address = dataplane.Spec.GetNetworking().GetAddress()
		if b.config.Params.CorefileTemplatePath != "" {
			corefileTemplate, err := os.ReadFile(b.config.Params.CorefileTemplatePath)
			if err != nil {
				return nil, kumaDpBootstrap, errors.Wrap(err, "could not read Corefile template")
			}
			kumaDpBootstrap.NetworkingConfig.CorefileTemplate = corefileTemplate
		}
		params.Service = dataplane.Spec.GetIdentifyingService()
		setAdminPort(dataplane.Spec.GetNetworking().GetAdmin().GetPort())
		setXdsTransportProtocolVariant(dataplane.Spec.GetEnvoy().GetXdsTransportProtocolVariant())

		err = b.getMetricsConfig(ctx, dataplane, &kumaDpBootstrap)
		if err != nil {
			return nil, kumaDpBootstrap, err
		}

	default:
		return nil, kumaDpBootstrap, errors.Errorf("unknown proxy type %v", params.ProxyType)
	}
	var err error
	if params.CertBytes, err = b.caCert(request); err != nil {
		return nil, kumaDpBootstrap, err
	}

	config, err := genConfig(params, b.proxyConfig, b.enableReloadableTokens)
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

func SANMismatchErr(host string, sans []string) error {
	return errors.Errorf("A data plane proxy is trying to connect to the control plane using %q address, but the certificate in the control plane has the following SANs %q. "+
		"Either change the --cp-address in kuma-dp to one of those or execute the following steps:\n"+
		"1) Generate a new certificate with the address you are trying to use. It is recommended to use trusted Certificate Authority, but you can also generate self-signed certificates using 'kumactl generate tls-certificate --type=server --hostname=%s'\n"+
		"2) Set KUMA_GENERAL_TLS_CERT_FILE and KUMA_GENERAL_TLS_KEY_FILE or the equivalent in Kuma CP config file to the new certificate.\n"+
		"3) Restart the control plane to read the new certificate and start kuma-dp.", host, sans, host)
}

func ISSANMismatchErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "A data plane proxy is trying to connect to the control plane using")
}

func (b *bootstrapGenerator) getMetricsConfig(
	ctx context.Context,
	dataplane *core_mesh.DataplaneResource,
	kumaDpBootstrap *KumaDpBootstrap,
) error {
	meshResource := core_mesh.NewMeshResource()
	err := b.resManager.Get(ctx, meshResource, core_store.GetByKey(dataplane.Meta.GetMesh(), core_model.NoMesh))
	if err != nil {
		return err
	}
	config, err := dataplane.GetPrometheusConfig(meshResource)
	if err != nil {
		return err
	}
	if config != nil {
		aggregateConfig := []AggregateMetricsConfig{}
		for _, config := range config.GetAggregate() {
			if config.GetEnabled() != nil && !config.GetEnabled().GetValue() {
				continue
			}
			aggregateConfig = append(aggregateConfig, AggregateMetricsConfig{
				Address: b.getMetricsAddress(config, dataplane),
				Name:    config.Name,
				Port:    config.Port,
				Path:    config.Path,
			})
		}
		kumaDpBootstrap.AggregateMetricsConfig = aggregateConfig
	}
	return nil
}

func (b *bootstrapGenerator) getMetricsAddress(
	metricsConfig *mesh_proto.PrometheusAggregateMetricsConfig,
	dataplane *core_mesh.DataplaneResource,
) string {
	if metricsConfig.Address != "" {
		return metricsConfig.Address
	}

	return dataplane.Spec.GetNetworking().GetAddress()
}

func (b *bootstrapGenerator) validateRequest(request types.BootstrapRequest) error {
	if b.authEnabledForProxyType[request.ProxyType] && request.DataplaneToken == "" && request.DataplaneTokenPath == "" {
		return DpTokenRequired
	}
	if b.config.Params.XdsHost == "" { // XdsHost takes precedence over Host in the request, so validate only when it is not set
		if !b.hostsAndIps[request.Host] {
			return SANMismatchErr(request.Host, b.hostsAndIps.slice())
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

func (b *bootstrapGenerator) zoneIngressFor(ctx context.Context, request types.BootstrapRequest, proxyId *core_xds.ProxyId) (*core_mesh.ZoneIngressResource, error) {
	if request.DataplaneResource != "" {
		res, err := rest.YAML.UnmarshalCore([]byte(request.DataplaneResource))
		if err != nil {
			return nil, err
		}
		zoneIngress, ok := res.(*core_mesh.ZoneIngressResource)
		if !ok {
			return nil, errors.Errorf("invalid resource")
		}
		if err := zoneIngress.Validate(); err != nil {
			return nil, err
		}
		return zoneIngress, nil
	} else {
		zoneIngress := core_mesh.NewZoneIngressResource()
		if err := b.resManager.Get(ctx, zoneIngress, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
			return nil, err
		}
		return zoneIngress, nil
	}
}

func (b *bootstrapGenerator) zoneEgressFor(ctx context.Context, request types.BootstrapRequest, proxyId *core_xds.ProxyId) (*core_mesh.ZoneEgressResource, error) {
	if request.DataplaneResource != "" {
		res, err := rest.YAML.UnmarshalCore([]byte(request.DataplaneResource))
		if err != nil {
			return nil, err
		}
		zoneEgress, ok := res.(*core_mesh.ZoneEgressResource)
		if !ok {
			return nil, errors.Errorf("invalid resource")
		}
		if err := zoneEgress.Validate(); err != nil {
			return nil, err
		}
		return zoneEgress, nil
	} else {
		zoneEgress := core_mesh.NewZoneEgressResource()
		if err := b.resManager.Get(ctx, zoneEgress, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
			return nil, err
		}
		return zoneEgress, nil
	}
}

func (b *bootstrapGenerator) validateMeshExist(ctx context.Context, mesh string) error {
	if err := b.resManager.Get(ctx, core_mesh.NewMeshResource(), core_store.GetByKey(mesh, core_model.NoMesh)); err != nil {
		if core_store.IsResourceNotFound(err) {
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
func (b *bootstrapGenerator) caCert(request types.BootstrapRequest) ([]byte, error) {
	// CaCert from the request takes precedence. It is only visible if user provides --ca-cert-file to kuma-dp run
	var cert []byte
	var origin string
	switch {
	case request.CaCert != "":
		cert = []byte(request.CaCert)
		origin = "request .CaCert"
	case b.xdsCertFile != "":
		var err error
		cert, err = os.ReadFile(b.xdsCertFile)
		origin = "file " + b.xdsCertFile
		if err != nil {
			return nil, errors.Wrapf(err, "failed getting cert from %s", origin)
		}
	default:
		return nil, nil
	}
	pemCert, _ := pem.Decode(cert)
	if pemCert == nil {
		return nil, errors.Errorf("could not parse certificate from %s", origin)
	}
	x509Cert, err := x509.ParseCertificate(pemCert.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse certificate %s", origin)
	}
	// checking just x509Cert.IsCA is not enough, because it's valid to generate CA without CA:TRUE basic constraint
	if x509Cert.BasicConstraintsValid && !x509Cert.IsCA {
		return nil, NotCA
	}
	return cert, nil
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
	if b.config.Params.AdminAccessLogPath == os.DevNull && operatingSystem == "windows" {
		// when AdminAccessLogPath was not explicitly set and DPP OS is Windows we need to set window specific DevNull.
		// otherwise when CP is on Linux, we would set /dev/null which is not valid on Windows.
		return "NUL"
	}
	return b.config.Params.AdminAccessLogPath
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
	certBytes, err := os.ReadFile(dpServerCertFile)
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
