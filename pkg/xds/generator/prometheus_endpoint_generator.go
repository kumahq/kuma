package generator

import (
	"net"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	manager_dataplane "github.com/Kong/kuma/pkg/core/managers/apis/dataplane"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/Kong/kuma/pkg/xds/envoy/names"
)

// PrometheusEndpointGenerator generates an inbound Envoy listener
// that forwards HTTP requests into the `/stats/prometheus`
// endpoint of the Envoy Admin API.
//
// When generating such a listener, it's important not to overshadow
// a port that is already in use by the application or other Envoy listeners.
// In the latter case we prefer not generate Prometheus endpoint at all
// rather than introduce undeterministic behaviour.
type PrometheusEndpointGenerator struct {
}

func (g PrometheusEndpointGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) ([]*core_xds.Resource, error) {
	prometheusEndpoint, err := proxy.Dataplane.GetPrometheusEndpoint(ctx.Mesh.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "could not get prometheus endpoint")
	}
	if prometheusEndpoint == nil {
		// Prometheus metrics must be enabled Mesh-wide for Prometheus endpoint to be generated.
		return nil, nil
	}
	if proxy.Metadata.GetAdminPort() == 0 {
		// It's not possible to export Prometheus metrics if Envoy Admin API has not been enabled on that dataplane.

		// TODO(yskopets): find a way to communicate this to users
		return nil, nil
	}

	prometheusEndpointAddress := proxy.Dataplane.Spec.GetNetworking().Address
	prometheusEndpointIP := net.ParseIP(prometheusEndpointAddress)

	if proxy.Dataplane.UsesInterface(prometheusEndpointIP, prometheusEndpoint.Port) {
		// If the Prometheus endpoint would otherwise overshadow one of interfaces of that Dataplane,
		// we prefer not to do that.

		// TODO(yskopets): find a way to communicate this to users
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
	prometheusListenerName := envoy_names.GetPrometheusListenerName()

	inbound, err := manager_dataplane.PrometheusInbound(proxy.Dataplane, ctx.Mesh.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "could not get prometheus inbound interface")
	}

	iface := proxy.Dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
	var listener *envoy_api.Listener
	if secureMetrics(prometheusEndpoint, ctx.Mesh.Resource) {
		listener, err = envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.InboundListener(prometheusListenerName, prometheusEndpointAddress, prometheusEndpoint.Port)).
			// generate filter chain that does not require mTLS when DP scrapes itself (for example DP next to Prometheus Server)
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
				Configure(envoy_listeners.SourceMatcher(proxy.Dataplane.Spec.GetNetworking().Address)).
				Configure(envoy_listeners.PrometheusEndpoint(prometheusListenerName, prometheusEndpoint.Path, envoyAdminClusterName)),
			)).
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
				Configure(envoy_listeners.PrometheusEndpoint(prometheusListenerName, prometheusEndpoint.Path, envoyAdminClusterName)).
				Configure(envoy_listeners.ServerSideMTLS(ctx, proxy.Metadata)).
				Configure(envoy_listeners.NetworkRBAC(prometheusListenerName, ctx.Mesh.Resource.MTLSEnabled(), proxy.TrafficPermissions[iface])),
			)).
			Build()
	} else {
		listener, err = envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.InboundListener(prometheusListenerName, prometheusEndpointAddress, prometheusEndpoint.Port)).
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
				Configure(envoy_listeners.PrometheusEndpoint(prometheusListenerName, prometheusEndpoint.Path, envoyAdminClusterName)),
			)).
			Build()
	}
	if err != nil {
		return nil, err
	}
	cluster, err := envoy_clusters.NewClusterBuilder().
		Configure(envoy_clusters.StaticCluster(envoyAdminClusterName, adminAddress, adminPort)).
		Build()
	if err != nil {
		return nil, err
	}
	resources := (&core_xds.ResourceSet{}).AddNamed(cluster, listener)
	return resources.List(), nil
}

func secureMetrics(cfg *mesh_proto.PrometheusMetricsBackendConfig, mesh *mesh_core.MeshResource) bool {
	return !cfg.SkipMTLS.GetValue() && mesh.MTLSEnabled()
}
