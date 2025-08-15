package virtualhosts

import (
	"sort"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes_v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

type RoutesConfigurer struct {
	Routes envoy_common.Routes
}

func (c RoutesConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	for i := range c.Routes {
		route := c.Routes[i]
		envoyRoute := &envoy_config_route_v3.Route{
			Match: c.routeMatch(route.Match),
			Name:  envoy_common.AnonymousResource,
			Action: &envoy_config_route_v3.Route_Route{
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

func (c RoutesConfigurer) setHeadersModifications(route *envoy_config_route_v3.Route, modify *mesh_proto.TrafficRoute_Http_Modify) {
	for _, add := range modify.GetRequestHeaders().GetAdd() {
		appendAction := envoy_config_core_v3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		if add.Append {
			appendAction = envoy_config_core_v3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		}
		route.RequestHeadersToAdd = append(route.RequestHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   add.Name,
				Value: add.Value,
			},
			AppendAction: appendAction,
		})
	}
	for _, remove := range modify.GetRequestHeaders().GetRemove() {
		route.RequestHeadersToRemove = append(route.RequestHeadersToRemove, remove.Name)
	}

	for _, add := range modify.GetResponseHeaders().GetAdd() {
		appendAction := envoy_config_core_v3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		if add.Append {
			appendAction = envoy_config_core_v3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		}
		route.ResponseHeadersToAdd = append(route.ResponseHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   add.Name,
				Value: add.Value,
			},
			AppendAction: appendAction,
		})
	}
	for _, remove := range modify.GetResponseHeaders().GetRemove() {
		route.ResponseHeadersToRemove = append(route.ResponseHeadersToRemove, remove.Name)
	}
}

func (c RoutesConfigurer) routeMatch(match *mesh_proto.TrafficRoute_Http_Match) *envoy_config_route_v3.RouteMatch {
	envoyMatch := &envoy_config_route_v3.RouteMatch{}

	if match.GetPath() != nil {
		c.setPathMatcher(match.GetPath(), envoyMatch)
	} else {
		// Path match is required on Envoy config so if there is only matching by header in TrafficRoute, we need to place
		// the default route match anyways.
		envoyMatch.PathSpecifier = &envoy_config_route_v3.RouteMatch_Prefix{
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

func (c RoutesConfigurer) headerMatcher(name string, matcher *mesh_proto.TrafficRoute_Http_Match_StringMatcher) *envoy_config_route_v3.HeaderMatcher {
	headerMatcher := &envoy_config_route_v3.HeaderMatcher{
		Name: name,
	}
	switch matcher.MatcherType.(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		headerMatcher.HeaderMatchSpecifier = &envoy_config_route_v3.HeaderMatcher_PrefixMatch{
			PrefixMatch: matcher.GetPrefix(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
		stringMatcher := envoy_type_matcher_v3.StringMatcher{
			MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
				Exact: matcher.GetExact(),
			},
		}
		headerMatcher.HeaderMatchSpecifier = &envoy_config_route_v3.HeaderMatcher_StringMatch{
			StringMatch: &stringMatcher,
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		headerMatcher.HeaderMatchSpecifier = &envoy_config_route_v3.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: &envoy_type_matcher_v3.RegexMatcher{
				Regex: matcher.GetRegex(),
			},
		}
	}
	return headerMatcher
}

func (c RoutesConfigurer) setPathMatcher(
	matcher *mesh_proto.TrafficRoute_Http_Match_StringMatcher,
	routeMatch *envoy_config_route_v3.RouteMatch,
) {
	switch matcher.MatcherType.(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		routeMatch.PathSpecifier = &envoy_config_route_v3.RouteMatch_Prefix{
			Prefix: matcher.GetPrefix(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
		routeMatch.PathSpecifier = &envoy_config_route_v3.RouteMatch_Path{
			Path: matcher.GetExact(),
		}
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		routeMatch.PathSpecifier = &envoy_config_route_v3.RouteMatch_SafeRegex{
			SafeRegex: &envoy_type_matcher_v3.RegexMatcher{
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

func (c RoutesConfigurer) routeAction(clusters []envoy_common.Cluster, modify *mesh_proto.TrafficRoute_Http_Modify) *envoy_config_route_v3.RouteAction {
	routeAction := &envoy_config_route_v3.RouteAction{}
	if len(clusters) != 0 {
		// Timeout can be configured only per outbound listener. So all clusters in the split
		// must have the same timeout. That's why we can take the timeout from the first cluster.
		if cluster, ok := clusters[0].(*envoy_common.ClusterImpl); ok {
			routeAction.Timeout = util_proto.Duration(cluster.Timeout().GetHttp().GetRequestTimeout().AsDuration())
		} else {
			routeAction.Timeout = util_proto.Duration(0)
		}
	}
	if len(clusters) == 1 {
		routeAction.ClusterSpecifier = &envoy_config_route_v3.RouteAction_Cluster{
			Cluster: clusters[0].Name(),
		}
	} else {
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
	c.setModifications(routeAction, modify)
	return routeAction
}

func (c RoutesConfigurer) setModifications(routeAction *envoy_config_route_v3.RouteAction, modify *mesh_proto.TrafficRoute_Http_Modify) {
	if modify.GetPath() != nil {
		switch modify.GetPath().Type.(type) {
		case *mesh_proto.TrafficRoute_Http_Modify_Path_RewritePrefix:
			routeAction.PrefixRewrite = modify.GetPath().GetRewritePrefix()
		case *mesh_proto.TrafficRoute_Http_Modify_Path_Regex:
			routeAction.RegexRewrite = &envoy_type_matcher_v3.RegexMatchAndSubstitute{
				Pattern: &envoy_type_matcher_v3.RegexMatcher{
					Regex: modify.GetPath().GetRegex().GetPattern(),
				},
				Substitution: modify.GetPath().GetRegex().GetSubstitution(),
			}
		}
	}

	if modify.GetHost() != nil {
		switch modify.GetHost().Type.(type) {
		case *mesh_proto.TrafficRoute_Http_Modify_Host_Value:
			routeAction.HostRewriteSpecifier = &envoy_config_route_v3.RouteAction_HostRewriteLiteral{
				HostRewriteLiteral: modify.GetHost().GetValue(),
			}
		case *mesh_proto.TrafficRoute_Http_Modify_Host_FromPath:
			routeAction.HostRewriteSpecifier = &envoy_config_route_v3.RouteAction_HostRewritePathRegex{
				HostRewritePathRegex: &envoy_type_matcher_v3.RegexMatchAndSubstitute{
					Pattern: &envoy_type_matcher_v3.RegexMatcher{
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
		rateLimit, err := envoy_routes_v3.NewRateLimitConfiguration(envoy_routes_v3.RateLimitConfigurationFromProto(route.RateLimit))
		if err != nil {
			return nil, err
		}
		typedPerFilterConfig["envoy.filters.http.local_ratelimit"] = rateLimit
	}

	return typedPerFilterConfig, nil
}
