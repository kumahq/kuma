package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

func GenerateRouteConfig(proxy *xds.Proxy, protocol mesh_proto.MeshGateway_Listener_Protocol, name string) *envoy_routes.RouteConfigurationBuilder {
	switch protocol {
	case mesh_proto.MeshGateway_Listener_HTTPS,
		mesh_proto.MeshGateway_Listener_HTTP:
	default:
		return nil
	}

	return envoy_routes.NewRouteConfigurationBuilder(proxy.APIVersion, name).
		Configure(
			envoy_routes.CommonRouteConfiguration(),
			envoy_routes.IgnorePortInHostMatching(),
			// TODO(jpeach) propagate merged listener tags.
			// Ideally we would propagate the tags header
			// to mesh services but not to external services,
			// but in the route configuration, we don't know
			// yet where the request will route to.
			// envoy_routes.TagsHeader(...),
			envoy_routes.ResetTagsHeader(),
		)

	// TODO(jpeach) apply additional route configuration configuration.
}
