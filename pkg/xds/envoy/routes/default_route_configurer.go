package routes

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

func DefaultRoute(clusters ...envoy_common.ClusterInfo) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.Add(&DefaultRouteConfigurer{
			RouteConfigurer: RouteConfigurer{
				clusters: clusters,
			},
		})
	})
}

type DefaultRouteConfigurer struct {
	RouteConfigurer
}

func (c DefaultRouteConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	route := &envoy_route.Route{
		Match: &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &envoy_route.Route_Route{
			Route: c.routeAction(),
		},
	}
	virtualHost.Routes = append(virtualHost.Routes, route)
	return nil
}

type RouteConfigurer struct {
	// Clusters to forward traffic to.
	clusters []envoy_common.ClusterInfo
}

func (c RouteConfigurer) routeAction() *envoy_route.RouteAction {
	routeAction := envoy_route.RouteAction{}
	if len(c.clusters) == 1 {
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: c.clusters[0].Name,
		}
	} else {
		var weightedClusters []*envoy_route.WeightedCluster_ClusterWeight
		for _, cluster := range c.clusters {
			weightedClusters = append(weightedClusters, &envoy_route.WeightedCluster_ClusterWeight{
				Name:   cluster.Name,
				Weight: &wrappers.UInt32Value{Value: cluster.Weight},
			})
		}
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_WeightedClusters{
			WeightedClusters: &envoy_route.WeightedCluster{
				Clusters: weightedClusters,
			},
		}
	}
	return &routeAction
}
