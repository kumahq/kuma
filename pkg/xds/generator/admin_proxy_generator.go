package generator

import (
	"strings"

	"github.com/Masterminds/semver/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/version"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginAdmin is a marker to indicate by which ProxyGenerator resources were generated.
const OriginAdmin = "admin"

var staticEndpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/ready",
		RewritePath: "/ready",
	},
}

var staticTlsEndpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/",
		RewritePath: "/",
	},
	{
		Path:        "/config_dump",
		RewritePath: "/config_dump",
	},
}

// AdminProxyGenerator generates resources to expose some endpoints of Admin API on public interface.
// By default, Admin API is exposed only on loopback interface because of security reasons.
type AdminProxyGenerator struct {
}

// backwards compatibility with 1.3.x
var HasCPValidationCtxInBootstrap = func(ver *mesh_proto.Version) (bool, error) {
	if ver.GetKumaDp().GetVersion() == "" { // mostly for tests but also for very old version of Kuma
		return false, nil
	}

	if strings.HasPrefix(ver.GetKumaDp().GetVersion(), version.DevVersionPrefix) {
		return true, nil
	}

	semverVer, err := semver.NewVersion(ver.KumaDp.GetVersion())
	if err != nil {
		return false, err
	}
	return !semverVer.LessThan(semver.MustParse("1.4.0")), nil
}

func (g AdminProxyGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	if proxy.Metadata.GetAdminPort() == 0 {
		// It's not possible to export Admin endpoints if Envoy Admin API has not been enabled on that dataplane.
		return nil, nil
	}

	adminPort := proxy.Metadata.GetAdminPort()
	// We assume that Admin API must be available on a loopback interface (while users
	// can override the default value `127.0.0.1` in the Bootstrap Server section of `kuma-cp` config,
	// the only reasonable alternative is `0.0.0.0`).
	// In contrast to `AdminPort`, we shouldn't trust `AdminAddress` from the Envoy node metadata
	// since it would allow a malicious user to manipulate that value and use Prometheus endpoint
	// as a gateway to another host.
	adminAddress := "127.0.0.1"
	envoyAdminClusterName := envoy_names.GetEnvoyAdminClusterName()
	cluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(envoy_clusters.ProvidedEndpointCluster(envoyAdminClusterName, false, core_xds.Endpoint{Target: adminAddress, Port: adminPort})).
		Build()
	if err != nil {
		return nil, err
	}

	resources := core_xds.NewResourceSet()

	for _, se := range staticEndpointPaths {
		se.ClusterName = envoyAdminClusterName
	}

	// We bind admin to 127.0.0.1 by default, creating another listener with same address and port will result in error.
	if g.getAddress(proxy) != "127.0.0.1" {
		filterChains := []envoy_listeners.ListenerBuilderOpt{
			envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
				Configure(envoy_listeners.StaticEndpoints(envoy_names.GetAdminListenerName(), staticEndpointPaths)),
			),
		}
		hasCpValidationCtx, err := HasCPValidationCtxInBootstrap(proxy.Metadata.Version)
		if err != nil {
			return nil, err
		}
		if hasCpValidationCtx {
			for _, se := range staticTlsEndpointPaths {
				se.ClusterName = envoyAdminClusterName
			}
			filterChains = append(filterChains, envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
				Configure(envoy_listeners.MatchTransportProtocol("tls")).
				Configure(envoy_listeners.StaticEndpoints(envoy_names.GetAdminListenerName(), staticTlsEndpointPaths)).
				Configure(envoy_listeners.ServerSideMTLSWithCP(ctx)),
			))
		}
		listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
			Configure(envoy_listeners.InboundListener(envoy_names.GetAdminListenerName(), g.getAddress(proxy), adminPort, core_xds.SocketAddressProtocolTCP)).
			Configure(envoy_listeners.TLSInspector()).
			Configure(filterChains...).
			Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   OriginAdmin,
			Resource: listener,
		})
	}

	resources.Add(&core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   OriginAdmin,
		Resource: cluster,
	})
	return resources, nil
}

func (g AdminProxyGenerator) getAddress(proxy *core_xds.Proxy) string {
	if proxy.Dataplane != nil {
		return proxy.Dataplane.Spec.GetNetworking().Address
	}

	if proxy.ZoneEgressProxy != nil {
		return proxy.ZoneEgressProxy.ZoneEgressResource.Spec.GetNetworking().GetAddress()
	}

	return proxy.ZoneIngress.Spec.GetNetworking().GetAddress()
}
