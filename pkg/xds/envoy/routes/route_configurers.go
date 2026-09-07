package routes

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

// RouteActionDirectResponse sets the direct response for a route
func RouteActionDirectResponse(status uint32, respStr string) RouteConfigurer {
	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		r.Action = &envoy_config_route_v3.Route_DirectResponse{
			DirectResponse: &envoy_config_route_v3.DirectResponseAction{
				Status: status,
				Body: &envoy_config_core_v3.DataSource{
					Specifier: &envoy_config_core_v3.DataSource_InlineString{
						InlineString: respStr,
					},
				},
			},
		}
		return nil
	})
}
