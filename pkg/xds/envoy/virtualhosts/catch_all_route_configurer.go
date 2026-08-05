package virtualhosts

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

type CatchAllRouteConfigurer struct {
	Cluster envoy_common.Cluster
}

func (c CatchAllRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
		// Path match is required by Envoy, and this route never narrows the
		// traffic down, so match every path.
		Match: &envoy_config_route_v3.RouteMatch{
			PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Name: envoy_common.AnonymousResource,
		Action: &envoy_config_route_v3.Route_Route{
			Route: &envoy_config_route_v3.RouteAction{
				// Keep the legacy route timeout override of 0s so Envoy request
				// timeouts stay disabled by default.
				Timeout: util_proto.Duration(0),
				ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
					Cluster: c.Cluster.Name(),
				},
			},
		},
	})
	return nil
}
