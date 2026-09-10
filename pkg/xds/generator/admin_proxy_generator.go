package generator

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/dynconf"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/system_names"
)

var staticEndpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/ready",
		RewritePath: "/ready",
	},
}

var staticTlsEndpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/ready",
		RewritePath: "/ready",
	},
	{
		Path:        "/",
		RewritePath: "/",
	},
}

// AdminProxyGenerator generates resources to expose some endpoints of Admin API on public interface.
// By default, Admin API is exposed only on loopback interface because of security reasons.
type AdminProxyGenerator struct{}

var adminAddressAllowedValues = map[string]struct{}{
	"127.0.0.1": {},
	"0.0.0.0":   {},
	"::1":       {},
	"::":        {},
	"":          {},
}

func (g AdminProxyGenerator) Generate(ctx context.Context, _ *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	if err := g.generateIdentityReadiness(resources, proxy); err != nil {
		return nil, err
	}
	if proxy.Metadata.GetAdminPort() == 0 {
		// It's not possible to export Admin endpoints if Envoy Admin API has not been enabled on that dataplane.
		if resources.Empty() {
			return nil, nil
		}
		return resources, nil
	}

	adminPort := proxy.Metadata.GetAdminPort()
	readinessPort := proxy.Metadata.GetReadinessPort()
	if readinessPort == 0 {
		return nil, errors.New("ReadinessPort has to be in (0, 65535] range")
	}
	// We assume that Admin API must be available on a loopback interface (while users
	// can override the default value `127.0.0.1` in the Bootstrap Server section of `kuma-cp` config,
	// the only reasonable alternatives are `::1`, `0.0.0.0` or `::`).
	// In contrast to `AdminPort`, we shouldn't trust `AdminAddress` from the Envoy node metadata
	// since it would allow a malicious user to manipulate that value and use Prometheus endpoint
	// as a gateway to another host.
	envoyAdminClusterName := system_names.SystemResourceNameEnvoyAdmin
	dppReadinessClusterName := system_names.SystemResourceNameReadiness
	adminAddress := proxy.Metadata.GetAdminAddress()
	if _, ok := adminAddressAllowedValues[adminAddress]; !ok {
		var allowedAddresses []string
		for _, address := range util_maps.SortedKeys(adminAddressAllowedValues) {
			allowedAddresses = append(allowedAddresses, fmt.Sprintf(`%q`, address))
		}
		return nil, errors.Errorf("envoy admin cluster is not allowed to have addresses other than %s", strings.Join(allowedAddresses, ", "))
	}
	switch adminAddress {
	case "", "0.0.0.0":
		adminAddress = "127.0.0.1"
	case "::":
		adminAddress = "::1"
	}

	var adminEndpoint core_xds.Endpoint
	if proxy.Metadata.GetAdminSocketPath() != "" {
		adminEndpoint = core_xds.Endpoint{
			UnixDomainPath: proxy.Metadata.GetAdminSocketPath(),
		}
	} else {
		adminEndpoint = core_xds.Endpoint{Target: adminAddress, Port: adminPort}
	}

	envoyAdminCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, envoyAdminClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(
			govalidator.IsIPv6(adminAddress),
			adminEndpoint)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, err
	}

	for _, se := range staticEndpointPaths {
		se.ClusterName = dppReadinessClusterName
	}
	for _, se := range staticTlsEndpointPaths {
		switch se.Path {
		case "/ready":
			se.ClusterName = dppReadinessClusterName
		default:
			se.ClusterName = envoyAdminClusterName
		}
	}

	// We bind admin to 127.0.0.1 by default, creating another listener with same address and port will result in error.
	if g.getAddress(proxy) != adminAddress {
		envoyAdminListenerName := system_names.SystemResourceNameEnvoyAdmin
		filterChains := []envoy_listeners.ListenerBuilderOpt{
			envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
				Configure(envoy_listeners.StaticEndpoints(proxy.Metadata.GetIPv6Enabled(), envoyAdminListenerName, staticEndpointPaths)),
			),
		}
		filterChains = append(filterChains, envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.MatchTransportProtocol("tls")).
			Configure(envoy_listeners.StaticEndpoints(proxy.Metadata.GetIPv6Enabled(), envoyAdminListenerName, staticTlsEndpointPaths)).
			Configure(envoy_listeners.ServerSideStaticMTLS(proxy.EnvoyAdminMTLSCerts)),
		))

		listener, err := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, g.getAddress(proxy), adminPort, core_xds.SocketAddressProtocolTCP, proxy.Metadata.HasFeature(xds_types.FeatureReusePort)).
			WithOverwriteName(envoyAdminListenerName).
			Configure(envoy_listeners.TLSInspector()).
			Configure(filterChains...).
			Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   metadata.OriginAdmin,
			Resource: listener,
		})
	}

	resources.Add(&core_xds.Resource{
		Name:     envoyAdminCluster.GetName(),
		Origin:   metadata.OriginAdmin,
		Resource: envoyAdminCluster,
	})

	readinessCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, dppReadinessClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(govalidator.IsIPv6(adminAddress), core_xds.Endpoint{Target: adminAddress, Port: readinessPort})).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, err
	}

	resources.Add(&core_xds.Resource{
		Name:     readinessCluster.GetName(),
		Origin:   metadata.OriginAdmin,
		Resource: readinessCluster,
	})

	return resources, nil
}

func (g AdminProxyGenerator) generateIdentityReadiness(resources *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	config := core_xds.IdentityReadinessConfig{Required: proxy.WorkloadIdentityRequired}
	if proxy.WorkloadIdentity != nil {
		config.ExpirationTime = proxy.WorkloadIdentity.ExpirationTime
		if proxy.WorkloadIdentity.ManagementMode == core_xds.KumaManagementMode && proxy.WorkloadIdentity.AdditionalResources != nil {
			certificateHash, err := bundledIdentityCertificateHash(proxy.WorkloadIdentity)
			if err != nil {
				return err
			}
			config.CertificateHash = certificateHash
		}
	}
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := dynconf.AddConfigRoute(proxy, resources, "identity-readiness", core_xds.IdentityReadinessPath, bytes); err != nil {
		return err
	}
	if proxy.WorkloadIdentity == nil {
		return nil
	}

	secretSource := proxy.WorkloadIdentity.IdentitySourceConfigurer
	if secretSource == nil {
		return errors.New("workload identity secret source is missing")
	}
	dtls, err := bldrs_tls.NewDownstreamTLSContext().
		Configure(bldrs_tls.DownstreamCommonTlsContext(
			bldrs_tls.NewCommonTlsContext().Configure(
				bldrs_tls.TlsCertificateSdsSecretConfigs([]*bldrs_common.Builder[envoy_tls.SdsSecretConfig]{
					bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(secretSource()),
				}),
			),
		)).Build()
	if err != nil {
		return err
	}

	name := system_names.SystemResourceNameIdentityReadiness
	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.DirectResponse(name, []envoy_listeners_v3.DirectResponseEndpoints{{
			Path:       "/ready",
			StatusCode: 200,
			Response:   "ready",
		}}, core_xds.LocalHostAddresses, proxy.Metadata.GetIPv6Enabled())).
		Configure(envoy_listeners.DownstreamTlsContext(dtls))
	listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion, name).
		Configure(envoy_listeners.PipeListener(core_xds.IdentityReadinessSocketName(proxy.Metadata.WorkDir))).
		Configure(envoy_listeners.FilterChain(filterChain)).
		Build()
	if err != nil {
		return err
	}
	resources.Add(&core_xds.Resource{Name: listener.GetName(), Origin: metadata.OriginAdmin, Resource: listener})
	return nil
}

func bundledIdentityCertificateHash(identity *core_xds.WorkloadIdentity) (string, error) {
	if identity.AdditionalResources == nil {
		return "", errors.New("workload identity resources are missing")
	}
	for _, resource := range identity.AdditionalResources.Resources(envoy_resource.SecretType) {
		secret, ok := resource.Resource.(*envoy_tls.Secret)
		if ok && secret.GetTlsCertificate() != nil {
			dataSource := secret.GetTlsCertificate().GetCertificateChain()
			certificatePEM := dataSource.GetInlineBytes()
			if len(certificatePEM) == 0 {
				certificatePEM = []byte(dataSource.GetInlineString())
			}
			block, _ := pem.Decode(certificatePEM)
			if block == nil || block.Type != "CERTIFICATE" {
				return "", errors.New("workload identity certificate is not valid PEM")
			}
			hash := sha256.Sum256(block.Bytes)
			return fmt.Sprintf("%x", hash), nil
		}
	}
	return "", errors.New("workload identity certificate secret is missing")
}

func (g AdminProxyGenerator) getAddress(proxy *core_xds.Proxy) string {
	if proxy.Dataplane != nil {
		return proxy.Dataplane.Spec.GetNetworking().Address
	}

	return ""
}
