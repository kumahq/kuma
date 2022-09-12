package v3

import (
	"sort"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
				Route: c.routeAction(route.Clusters, route.Modify),
			},
		}

		typedPerFilterConfig, err := c.typedPerFilterConfig(&route)
		if err != nil {
			return err
		}
		envoyRoute.TypedPerFilterConfig = typedPerFilterConfig

		c.setHeadersModifications(envoyRoute, route.Modify)

		virtualHost.Routes = append(virtualHost.Routes, envoyRoute)
	}
	return nil
}

func (c RoutesConfigurer) setHeadersModifications(route *envoy_route.Route, modify *mesh_proto.TrafficRoute_Http_Modify) {
	for _, add := range modify.GetRequestHeaders().GetAdd() {
		route.RequestHeadersToAdd = append(route.RequestHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   add.Name,
				Value: add.Value,
			},
			Append: util_proto.Bool(add.Append),
		})
	}
	for _, remove := range modify.GetRequestHeaders().GetRemove() {
		route.RequestHeadersToRemove = append(route.RequestHeadersToRemove, remove.Name)
	}

	for _, add := range modify.GetResponseHeaders().GetAdd() {
		route.ResponseHeadersToAdd = append(route.ResponseHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   add.Name,
				Value: add.Value,
			},
			Append: util_proto.Bool(add.Append),
		})
	}
	for _, remove := range modify.GetResponseHeaders().GetRemove() {
		route.ResponseHeadersToRemove = append(route.ResponseHeadersToRemove, remove.Name)
	}
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
		stringMatcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: matcher.GetExact(),
			},
		}
		headerMatcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_StringMatch{
			StringMatch: &stringMatcher,
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

func (c RoutesConfigurer) routeAction(clusters []envoy_common.Cluster, modify *mesh_proto.TrafficRoute_Http_Modify) *envoy_route.RouteAction {
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
	c.setModifications(routeAction, modify)
	return routeAction
}

func (c RoutesConfigurer) setModifications(routeAction *envoy_route.RouteAction, modify *mesh_proto.TrafficRoute_Http_Modify) {
	if modify.GetPath() != nil {
		switch modify.GetPath().Type.(type) {
		case *mesh_proto.TrafficRoute_Http_Modify_Path_RewritePrefix:
			routeAction.PrefixRewrite = modify.GetPath().GetRewritePrefix()
		case *mesh_proto.TrafficRoute_Http_Modify_Path_Regex:
			routeAction.RegexRewrite = &envoy_type_matcher.RegexMatchAndSubstitute{
				Pattern: &envoy_type_matcher.RegexMatcher{
					EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
					},
					Regex: modify.GetPath().GetRegex().GetPattern(),
				},
				Substitution: modify.GetPath().GetRegex().GetSubstitution(),
			}
		}
	}

	if modify.GetHost() != nil {
		switch modify.GetHost().Type.(type) {
		case *mesh_proto.TrafficRoute_Http_Modify_Host_Value:
			routeAction.HostRewriteSpecifier = &envoy_route.RouteAction_HostRewriteLiteral{
				HostRewriteLiteral: modify.GetHost().GetValue(),
			}
		case *mesh_proto.TrafficRoute_Http_Modify_Host_FromPath:
			routeAction.HostRewriteSpecifier = &envoy_route.RouteAction_HostRewritePathRegex{
				HostRewritePathRegex: &envoy_type_matcher.RegexMatchAndSubstitute{
					Pattern: &envoy_type_matcher.RegexMatcher{
						EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
						},
						Regex: modify.GetHost().GetFromPath().GetPattern(),
					},
					Substitution: modify.GetHost().GetFromPath().GetSubstitution(),
				},
			}
		}
	}
}

func (c *RoutesConfigurer) typedPerFilterConfig(route *envoy_common.Route) (map[string]*anypb.Any, error) {
	typedPerFilterConfig := map[string]*anypb.Any{}

	if route.RateLimit != nil {
		rateLimit, err := NewRateLimitConfiguration(route.RateLimit.GetConf().GetHttp())
		if err != nil {
			return nil, err
		}
		typedPerFilterConfig["envoy.filters.http.local_ratelimit"] = rateLimit
	}

	return typedPerFilterConfig, nil
}
