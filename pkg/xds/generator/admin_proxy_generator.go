package generator

import (
	"fmt"

	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginAdmin is a marker to indicate by which ProxyGenerator resources were generated.
const OriginAdmin = "admin"

var staticEnpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/ready",
		RewritePath: "/ready",
	},
}

var staticTlsEnpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/",
		RewritePath: "/",
	},
}

// AdminProxyGenerator generates resources to expose some endpoints of Admin API on public interface.
// By default, Admin API is exposed only on loopback interface because of security reasons.
type AdminProxyGenerator struct {
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
		Configure(envoy_clusters.StaticCluster(envoyAdminClusterName, adminAddress, adminPort)).
		Build()
	if err != nil {
		return nil, err
	}

	resources := core_xds.NewResourceSet()

	for _, se := range staticEnpointPaths {
		se.ClusterName = envoyAdminClusterName
	}

	for _, se := range staticTlsEnpointPaths {
		se.ClusterName = envoyAdminClusterName

		token, err := ctx.EnvoyAdminClient.GenerateAPIToken(proxy.Dataplane)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to generate the API token")
		}
		se.Header = "Authorization"
		se.HeaderExactMatch = fmt.Sprintf("Bearer %s", token)
	}

	// We bind admin to 127.0.0.1 by default, creating another listener with same address and port will result in error.
	if proxy.Dataplane.Spec.GetNetworking().Address != "127.0.0.1" {
		listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
			Configure(envoy_listeners.InboundListener(envoy_names.GetAdminListenerName(), proxy.Dataplane.Spec.GetNetworking().Address, adminPort, core_xds.SocketAddressProtocolTCP)).
			Configure(envoy_listeners.TLSInspector()).
			Configure(
				envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
					Configure(envoy_listeners.StaticEndpoints(envoy_names.GetAdminListenerName(), staticEnpointPaths)),
				),
				envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
					Configure(envoy_listeners.FilterChainMatch("tls")).
					Configure(envoy_listeners.StaticTlsEndpoints(envoy_names.GetAdminListenerName(), ctx.ControlPlane.AdminProxyKeyPair, staticTlsEnpointPaths)),
				)).
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
