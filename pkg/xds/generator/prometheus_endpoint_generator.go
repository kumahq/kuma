package generator

import (
	"net"

	"github.com/pkg/errors"

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

	// It should be always possible to scrape metrics out of a Dataplane,
	// even when it doesn't have any inbound interfaces (e.g., gateway scenario).
	// Therefore, we always bind Prometheus endpoint to `0.0.0.0`
	// instead of trying to reuse IP address of an inbound listener.
	prometheusEndpointIP := net.IPv4zero // 0.0.0.0
	prometheusEndpointAddress := prometheusEndpointIP.String()

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

	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(prometheusListenerName, prometheusEndpointAddress, prometheusEndpoint.Port)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
			Configure(envoy_listeners.PrometheusEndpoint(prometheusListenerName, prometheusEndpoint.Path, envoyAdminClusterName)))).
		Build()
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
