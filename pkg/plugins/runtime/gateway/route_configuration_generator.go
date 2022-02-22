package gateway

import (
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

func GenerateRouteConfig(ctx xds_context.Context, info GatewayListenerInfo) *envoy_routes.RouteConfigurationBuilder {
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
