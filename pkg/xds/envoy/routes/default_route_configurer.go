package routes

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v2"
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
	routeAction := envoy_route.RouteAction{
		// This disable the timeout of the response. As Envoy docs suggest
		// disabling this solves problems with long lived and streaming requests.
		Timeout: &duration.Duration{Seconds: 0, Nanos: 0},
	}
	if len(c.subsets) == 1 {
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: c.subsets[0].ClusterName,
		}
		routeAction.MetadataMatch = envoy_metadata.LbMetadata(c.subsets[0].Tags)
	} else {
		var weightedClusters []*envoy_route.WeightedCluster_ClusterWeight
		var totalWeight uint32
		for _, subset := range c.subsets {
			weightedClusters = append(weightedClusters, &envoy_route.WeightedCluster_ClusterWeight{
				Name:          subset.ClusterName,
				Weight:        &wrappers.UInt32Value{Value: subset.Weight},
				MetadataMatch: envoy_metadata.LbMetadata(subset.Tags),
			})
			totalWeight += subset.Weight
		}
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_WeightedClusters{
			WeightedClusters: &envoy_route.WeightedCluster{
				Clusters:    weightedClusters,
				TotalWeight: &wrappers.UInt32Value{Value: totalWeight},
			},
		}
	}
	return &routeAction
}
