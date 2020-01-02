package generator

import (
	"net"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	"github.com/Kong/kuma/pkg/xds/envoy"
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

func (g PrometheusEndpointGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) ([]*Resource, error) {
	meshLevelSettings := ctx.Mesh.Resource.Spec.Metrics.GetPrometheus()
	if meshLevelSettings == nil {
		// Prometheus metrics must be enabled Mesh-wide for Prometheus endpoint to be generated.
		return nil, nil
	}
	if proxy.Metadata.GetAdminPort() == 0 {
		// It's not possible to export Prometheus metrics if Envoy Admin API has not been enabled on that dataplane.

		// TODO(yskopets): find a way to communicate this to users
		return nil, nil
	}
	dataplaneLevelSettings := proxy.Dataplane.Spec.Metrics.GetPrometheus()
	prometheusEndpoint := g.mergePrometheusSettings(meshLevelSettings, dataplaneLevelSettings)

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
	envoyAdminClusterName := envoyAdminClusterName()
	prometheusListenerName := prometheusListenerName()
	virtual := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPort() != 0
	return []*Resource{
		// CDS resource
		&Resource{
			Name:     envoyAdminClusterName,
			Version:  "",
			Resource: envoy.CreateLocalCluster(envoyAdminClusterName, adminAddress, adminPort),
		},
		// LDS resource
		&Resource{
			Name:     prometheusListenerName,
			Version:  "",
			Resource: envoy.CreatePrometheusListener(ctx, prometheusListenerName, prometheusEndpointAddress, prometheusEndpoint.Port, prometheusEndpoint.Path, envoyAdminClusterName, virtual, proxy.Metadata),
		},
	}, nil
}

func (_ PrometheusEndpointGenerator) mergePrometheusSettings(meshLevel *mesh_proto.Metrics_Prometheus, dataplaneLevel *mesh_proto.Metrics_Prometheus) mesh_proto.Metrics_Prometheus {
	result := mesh_proto.Metrics_Prometheus{}
	proto.Merge(&result, meshLevel)
	proto.Merge(&result, dataplaneLevel)
	return result
}
