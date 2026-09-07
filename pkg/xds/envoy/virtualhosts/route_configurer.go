package virtualhosts

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type VirtualHostBasicRouteConfigurer struct {
	Cluster string
}

func (c VirtualHostBasicRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
		Match: &envoy_config_route_v3.RouteMatch{
			PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &envoy_config_route_v3.Route_Route{
			Route: &envoy_config_route_v3.RouteAction{
				ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
					Cluster: c.Cluster,
				},
			},
		},
	})
	return nil
}

type VirtualHostDirectResponseRouteConfigurer struct {
	status      uint32
	responseMsg string
}

func (c VirtualHostDirectResponseRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
		Match: &envoy_config_route_v3.RouteMatch{
			PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &envoy_config_route_v3.Route_DirectResponse{
			DirectResponse: &envoy_config_route_v3.DirectResponseAction{
				Status: c.status,
				Body: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{
						InlineString: c.responseMsg,
					},
				},
			},
		},
	})
	return nil
}
