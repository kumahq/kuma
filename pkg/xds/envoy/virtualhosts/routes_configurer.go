package virtualhosts

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

type RoutesConfigurer struct {
	Routes                envoy_common.Routes
	ConfigureRouteTimeout bool
}

func (c RoutesConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	for i := range c.Routes {
		route := c.Routes[i]
		virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
			// Path match is required by Envoy, and these routes never narrow
			// the traffic down, so match every path.
			Match: &envoy_config_route_v3.RouteMatch{
				PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
					Prefix: "/",
				},
			},
			Name: envoy_common.AnonymousResource,
			Action: &envoy_config_route_v3.Route_Route{
				Route: c.routeAction(route.Clusters),
			},
		})
	}
	return nil
}

func (c RoutesConfigurer) hasExternal(clusters []envoy_common.Cluster) bool {
	for _, cluster := range clusters {
		if cluster.IsExternalService() {
			return true
		}
	}
	return false
}

func (c RoutesConfigurer) routeAction(clusters []envoy_common.Cluster) *envoy_config_route_v3.RouteAction {
	routeAction := &envoy_config_route_v3.RouteAction{}
	if c.ConfigureRouteTimeout && len(clusters) != 0 {
		routeAction.Timeout = util_proto.Duration(0)
	}

	switch len(clusters) {
	case 0:
		// Leave the cluster unset when no upstreams survive route generation.
		// Callers currently filter those routes out, but avoiding an empty
		// WeightedClusters config keeps this helper safe on its own.
	case 1:
		routeAction.ClusterSpecifier = &envoy_config_route_v3.RouteAction_Cluster{
			Cluster: clusters[0].Name(),
		}
	default:
		var weightedClusters []*envoy_config_route_v3.WeightedCluster_ClusterWeight
		for _, cluster := range clusters {
			cw := &envoy_config_route_v3.WeightedCluster_ClusterWeight{
				Name:   cluster.Name(),
				Weight: util_proto.UInt32(1),
			}

			if c, ok := cluster.(*envoy_common.ClusterImpl); ok {
				cw.Weight = util_proto.UInt32(c.Weight())
			}

			weightedClusters = append(weightedClusters, cw)
		}
		routeAction.ClusterSpecifier = &envoy_config_route_v3.RouteAction_WeightedClusters{
			WeightedClusters: &envoy_config_route_v3.WeightedCluster{
				Clusters: weightedClusters,
			},
		}
	}
	if c.hasExternal(clusters) {
		routeAction.HostRewriteSpecifier = &envoy_config_route_v3.RouteAction_AutoHostRewrite{
			AutoHostRewrite: util_proto.Bool(true),
		}
	}
	return routeAction
}
