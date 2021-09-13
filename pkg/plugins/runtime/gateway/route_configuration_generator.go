package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

// RouteConfigurationGenerator generates Kuma gateway listeners.
type RouteConfigurationGenerator struct{}

func (*RouteConfigurationGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	switch p {
	case mesh_proto.Gateway_Listener_UDP,
		mesh_proto.Gateway_Listener_TCP,
		mesh_proto.Gateway_Listener_TLS,
		mesh_proto.Gateway_Listener_HTTP,
		mesh_proto.Gateway_Listener_HTTPS:
		return true
	default:
		return false
	}
}

func (*RouteConfigurationGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	if info.Resources.RouteConfiguration != nil {
		return nil, nil
	}

	info.Resources.RouteConfiguration = envoy_routes.NewRouteConfigurationBuilder(info.Proxy.APIVersion).
		Configure(
			envoy_routes.CommonRouteConfiguration(info.Listener.ResourceName),
			// TODO(jpeach) propagate merged listener tags.
			// envoy_routes.TagsHeader(...),
			envoy_routes.ResetTagsHeader(),
		)

	// TODO(jpeach) apply additional route configuration configuration.

	return nil, nil
}
