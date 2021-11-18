package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
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
	resourceName := envoy_names.Join(info.Listener.ResourceName, info.Host.Hostname)

	info.Resources.RouteConfiguration = envoy_routes.NewRouteConfigurationBuilder(info.Proxy.APIVersion).
		Configure(
			envoy_routes.CommonRouteConfiguration(resourceName),
			// TODO(jpeach) propagate merged listener tags.
			// Ideally we would propagate the tags header
			// to mesh services but not to external services,
			// but in the route configuration, we don't know
			// yet where the request will route to.
			// envoy_routes.TagsHeader(...),
			envoy_routes.ResetTagsHeader(),
		)

	// TODO(jpeach) apply additional route configuration configuration.

	scoped := route.ScopedRouteBuilder{}
	scoped.Configure(
		route.ScopedCommonConfiguration("scoped-"+resourceName, resourceName),
		route.ScopedRouteOnDemand(false), // Experimentally, enabling on-demand does not work.
		route.ScopedKeyFragment(info.Host.Hostname),
	)

	resources := ResourceAggregator{}
	if err := resources.AddSet(BuildResourceSet(&scoped)); err != nil {
		return nil, err
	}

	return resources.Get(), nil
}
