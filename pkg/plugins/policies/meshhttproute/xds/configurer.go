package xds

import (
	"strings"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
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
	matches := c.routeMatch(c.Matches)

	var envoyRoutes []*envoy_route.Route
	for _, match := range matches {
		envoyRoutes = append(envoyRoutes, &envoy_route.Route{
			Match: match,
			Action: &envoy_route.Route_Route{
				Route: c.routeAction(c.Clusters),
			},
			TypedPerFilterConfig: map[string]*anypb.Any{},
		})
	}

	for _, envoyRoute := range envoyRoutes {
		applyRouteFilters(c.Filters, envoyRoute)
	}

	virtualHost.Routes = append(virtualHost.Routes, envoyRoutes...)
	return nil
}

func (c RoutesConfigurer) routeMatch(matches []api.Match) []*envoy_route.RouteMatch {
	var allEnvoyMatches []*envoy_route.RouteMatch

	for _, match := range matches {
		var envoyMatches []*envoy_route.RouteMatch

		if match.Path != nil {
			envoyMatches = c.routePathMatch(*match.Path)
		} else {
			envoyMatches = []*envoy_route.RouteMatch{{}}
		}

		for _, envoyMatch := range envoyMatches {
			if match.Method != nil {
				c.routeMethodMatch(envoyMatch, *match.Method)
			}
		}

		allEnvoyMatches = append(allEnvoyMatches, envoyMatches...)
	}

	return allEnvoyMatches
}

func (c RoutesConfigurer) routePathMatch(match api.PathMatch) []*envoy_route.RouteMatch {
	switch match.Type {
	case api.Exact:
		return []*envoy_route.RouteMatch{{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: match.Value,
			},
		}}
	case api.Prefix:
		if match.Value == "/" {
			return []*envoy_route.RouteMatch{{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{
					Prefix: match.Value,
				},
			}}
		}
		// This case forces us to create two different envoy matches to
		// replicate the "path-separated prefixes only" semantics
		trimmed := strings.TrimSuffix(match.Value, "/")
		return []*envoy_route.RouteMatch{{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: trimmed,
			},
		}, {
			PathSpecifier: &envoy_route.RouteMatch_Prefix{
				Prefix: trimmed + "/",
			},
		}}
	case api.RegularExpression:
		matcher := &envoy_type_matcher.RegexMatcher{
			Regex: match.Value,
			EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
				GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
			},
		}
		return []*envoy_route.RouteMatch{{
			PathSpecifier: &envoy_route.RouteMatch_SafeRegex{
				SafeRegex: matcher,
			},
		}}
	default:
		panic("impossible")
	}
}

func (c RoutesConfigurer) routeMethodMatch(envoyMatch *envoy_route.RouteMatch, method api.Method) {
	matcher := envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
			Exact: string(method),
		},
	}
	envoyMatch.Headers = append(envoyMatch.Headers,
		&envoy_route.HeaderMatcher{
			Name: ":method",
			HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
				StringMatch: &matcher,
			},
		},
	)
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
