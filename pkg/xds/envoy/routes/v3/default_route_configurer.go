package v3

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type DefaultRouteConfigurer struct {
	// Clusters to forward traffic to.
	Clusters []envoy_common.Cluster
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

func (c DefaultRouteConfigurer) hasExternal() bool {
	for _, cluster := range c.Clusters {
		if cluster.IsExternalService() {
			return true
		}
	}
	return false
}

func (c DefaultRouteConfigurer) routeAction() *envoy_route.RouteAction {
	routeAction := envoy_route.RouteAction{}
	if len(c.Clusters) != 0 {
		routeAction.Timeout = ptypes.DurationProto(c.Clusters[0].Timeout().GetHttp().GetRequestTimeout().AsDuration())
	}
	if len(c.Clusters) == 1 {
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: c.Clusters[0].Name(),
		}
	} else {
		var weightedClusters []*envoy_route.WeightedCluster_ClusterWeight
		var totalWeight uint32
		for _, cluster := range c.Clusters {
			weightedClusters = append(weightedClusters, &envoy_route.WeightedCluster_ClusterWeight{
				Name:   cluster.Name(),
				Weight: &wrappers.UInt32Value{Value: cluster.Weight()},
			})
			totalWeight += cluster.Weight()
		}
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_WeightedClusters{
			WeightedClusters: &envoy_route.WeightedCluster{
				Clusters:    weightedClusters,
				TotalWeight: &wrappers.UInt32Value{Value: totalWeight},
			},
		}
	}
	if c.hasExternal() {
		routeAction.HostRewriteSpecifier = &envoy_route.RouteAction_AutoHostRewrite{
			AutoHostRewrite: &wrappers.BoolValue{Value: true},
		}
	}
	return &routeAction
}
