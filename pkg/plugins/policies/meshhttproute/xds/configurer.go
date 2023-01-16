package xds

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type RoutesConfigurer struct {
	Matches  []api.Match
	Filters  []api.Filter
	Clusters []envoy_common.Cluster
}

func (c RoutesConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	envoyRoute := &envoy_route.Route{
		Match: c.routeMatch(c.Matches),
		Action: &envoy_route.Route_Route{
			Route: c.routeAction(c.Clusters),
		},
		TypedPerFilterConfig: map[string]*anypb.Any{},
	}

	applyRouteFilters(c.Filters, envoyRoute)

	virtualHost.Routes = append(virtualHost.Routes, envoyRoute)
	return nil
}

func (c RoutesConfigurer) routeMatch(matches []api.Match) *envoy_route.RouteMatch {
	envoyMatch := &envoy_route.RouteMatch{}

	for _, match := range matches {
		if match.Path.Prefix != "" {
			envoyMatch.PathSpecifier = &envoy_route.RouteMatch_Prefix{
				Prefix: match.Path.Prefix,
			}
		}
	}

	return envoyMatch
}

func (c RoutesConfigurer) hasExternal(clusters []envoy_common.Cluster) bool {
	for _, cluster := range clusters {
		if cluster.IsExternalService() {
			return true
		}
	}
	return false
}

func (c RoutesConfigurer) routeAction(clusters []envoy_common.Cluster) *envoy_route.RouteAction {
	routeAction := &envoy_route.RouteAction{}
	if len(clusters) != 0 {
		routeAction.Timeout = util_proto.Duration(clusters[0].Timeout().GetHttp().GetRequestTimeout().AsDuration())
	}
	if len(clusters) == 1 {
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: clusters[0].Name(),
		}
	} else {
		var weightedClusters []*envoy_route.WeightedCluster_ClusterWeight
		var totalWeight uint32
		for _, cluster := range clusters {
			weightedClusters = append(weightedClusters, &envoy_route.WeightedCluster_ClusterWeight{
				Name:   cluster.Name(),
				Weight: util_proto.UInt32(cluster.Weight()),
			})
			totalWeight += cluster.Weight()
		}
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_WeightedClusters{
			WeightedClusters: &envoy_route.WeightedCluster{
				Clusters:    weightedClusters,
				TotalWeight: util_proto.UInt32(totalWeight),
			},
		}
	}
	if c.hasExternal(clusters) {
		routeAction.HostRewriteSpecifier = &envoy_route.RouteAction_AutoHostRewrite{
			AutoHostRewrite: util_proto.Bool(true),
		}
	}
	return routeAction
}

func applyRouteFilters(filters []api.Filter, route *envoy_route.Route) {
	for _, filter := range filters {
		routeFilter(filter, route)
	}
}
