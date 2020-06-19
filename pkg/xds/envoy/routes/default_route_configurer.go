package routes

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

func DefaultRoute(subsets ...envoy_common.ClusterSubset) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.Add(&DefaultRouteConfigurer{
			RouteConfigurer: RouteConfigurer{
				subsets: subsets,
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
	// Subsets to forward traffic to.
	subsets []envoy_common.ClusterSubset
}

func (c RouteConfigurer) routeAction() *envoy_route.RouteAction {
	routeAction := envoy_route.RouteAction{}
	if len(c.subsets) == 1 && envoy_common.Metadata(c.subsets[0].Tags) == nil {
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: c.subsets[0].ClusterName,
		}
	} else {
		var weightedClusters []*envoy_route.WeightedCluster_ClusterWeight
		for _, subset := range c.subsets {
			weightedClusters = append(weightedClusters, &envoy_route.WeightedCluster_ClusterWeight{
				Name:          subset.ClusterName,
				Weight:        &wrappers.UInt32Value{Value: subset.Weight},
				MetadataMatch: envoy_common.Metadata(subset.Tags),
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
