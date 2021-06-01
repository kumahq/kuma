package v2

import (
	"sort"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type RoutesConfigurer struct {
	Routes envoy_common.Routes
}

func (c RoutesConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	for _, route := range c.Routes {
		envoyRoute := &envoy_route.Route{
			Match: c.routeMatch(route.Match),
			Action: &envoy_route.Route_Route{
				Route: c.routeAction(route.Clusters),
			},
			TypedPerFilterConfig: route.TypedPerFilterConfig,
		}
		virtualHost.Routes = append(virtualHost.Routes, envoyRoute)
	}
	return nil
}

func (c RoutesConfigurer) routeMatch(match *mesh_proto.TrafficRoute_Http_Match) *envoy_route.RouteMatch {
	envoyMatch := &envoy_route.RouteMatch{}

	if match.GetPath() != nil {
		c.setPathMatcher(match.GetPath(), envoyMatch)
	} else {
		// Path match is required on Envoy config so if there is only matching by header in TrafficRoute, we need to place
		// the default route match anyways.
		envoyMatch.PathSpecifier = &envoy_route.RouteMatch_Prefix{
			Prefix: "/",
		}
	}

	var headers []string
	for headerName := range match.GetHeaders() {
		headers = append(headers, headerName)
	}
	sort.Strings(headers) // sort for stability of Envoy config
	for _, headerName := range headers {
		envoyMatch.Headers = append(envoyMatch.Headers, c.headerMatcher(headerName, match.Headers[headerName]))
	}
	if match.GetMethod() != nil {
		envoyMatch.Headers = append(envoyMatch.Headers, c.headerMatcher(":method", match.GetMethod()))
	}

	return envoyMatch
}

func (c RoutesConfigurer) headerMatcher(name string, matcher *mesh_proto.TrafficRoute_Http_Match_StringMatcher) *envoy_route.HeaderMatcher {
	headerMatcher := &envoy_route.HeaderMatcher{
		Name: name,
	}
	switch matcher.MatcherType.(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		headerMatcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_PrefixMatch{
			PrefixMatch: matcher.GetPrefix(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
		headerMatcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_ExactMatch{
			ExactMatch: matcher.GetExact(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		headerMatcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: &envoy_type_matcher.RegexMatcher{
				EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
				},
				Regex: matcher.GetRegex(),
			},
		}
	}
	return headerMatcher
}

func (c RoutesConfigurer) setPathMatcher(
	matcher *mesh_proto.TrafficRoute_Http_Match_StringMatcher,
	routeMatch *envoy_route.RouteMatch,
) {
	switch matcher.MatcherType.(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		routeMatch.PathSpecifier = &envoy_route.RouteMatch_Prefix{
			Prefix: matcher.GetPrefix(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
		routeMatch.PathSpecifier = &envoy_route.RouteMatch_Path{
			Path: matcher.GetExact(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		routeMatch.PathSpecifier = &envoy_route.RouteMatch_SafeRegex{
			SafeRegex: &envoy_type_matcher.RegexMatcher{
				EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
				},
				Regex: matcher.GetRegex(),
			},
		}
	}
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
	routeAction := envoy_route.RouteAction{}
	if len(clusters) != 0 {
		routeAction.Timeout = ptypes.DurationProto(clusters[0].Timeout().GetHttp().GetRequestTimeout().AsDuration())
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
	if c.hasExternal(clusters) {
		routeAction.HostRewriteSpecifier = &envoy_route.RouteAction_AutoHostRewrite{
			AutoHostRewrite: &wrappers.BoolValue{Value: true},
		}
	}
	return &routeAction
}
