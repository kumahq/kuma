package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

// RouteConfigurationGenerator generates Kuma gateway listeners.
type RouteConfigurationGenerator struct{}

func (*RouteConfigurationGenerator) SupportsProtocol(p mesh_proto.MeshGateway_Listener_Protocol) bool {
	switch p {
	case mesh_proto.MeshGateway_Listener_UDP,
		mesh_proto.MeshGateway_Listener_TCP,
		mesh_proto.MeshGateway_Listener_TLS,
		mesh_proto.MeshGateway_Listener_HTTP,
		mesh_proto.MeshGateway_Listener_HTTPS:
		return true
	default:
		return false
	}
}

func (*RouteConfigurationGenerator) Generate(ctx xds_context.Context, info *GatewayListenerInfo) *envoy_routes.RouteConfigurationBuilder {
	return envoy_routes.NewRouteConfigurationBuilder(info.Proxy.APIVersion).
		Configure(
			envoy_routes.CommonRouteConfiguration(info.Listener.ResourceName),
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
