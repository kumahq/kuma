package xds

import (
	"strings"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds/filters"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type RoutesConfigurer struct {
	Hash                    common_api.MatchesHash
	Match                   api.Match
	Filters                 []api.Filter
	Split                   []envoy_common.Split
	BackendRefToClusterName map[common_api.BackendRefHash]string
}

func (c RoutesConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	matches := c.routeMatch(c.Match)

	for _, match := range matches {
		rb := envoy_routes.NewRouteBuilder(envoy_common.APIV3, string(c.Hash)).
			Configure(envoy_routes.RouteMustConfigureFunc(func(envoyRoute *envoy_route.Route) {
				// todo: create configurers for Match and Action
				envoyRoute.Match = match.routeMatch
				envoyRoute.Action = &envoy_route.Route_Route{
					Route: c.routeAction(c.Split),
				}
			}))

		// We pass the information about whether this match was created from
		// a prefix match along to the filters because it's no longer
		// possible to know for sure with just an envoy_route.Route
		for _, filter := range c.Filters {
			switch filter.Type {
			case api.RequestHeaderModifierType:
				rb.Configure(filters.NewRequestHeaderModifier(*filter.RequestHeaderModifier))
			case api.ResponseHeaderModifierType:
				rb.Configure(filters.NewResponseHeaderModifier(*filter.ResponseHeaderModifier))
			case api.RequestRedirectType:
				rb.Configure(filters.NewRequestRedirect(*filter.RequestRedirect, match.prefixMatch))
			case api.URLRewriteType:
				rb.Configure(filters.NewURLRewrite(*filter.URLRewrite, match.prefixMatch))
			case api.RequestMirrorType:
				rb.Configure(filters.NewRequestMirror(*filter.RequestMirror, c.BackendRefToClusterName))
			}
		}

		r, err := rb.Build()
		if err != nil {
			return err
		}

		virtualHost.Routes = append(virtualHost.Routes, r.(*envoy_route.Route))
	}

	return nil
}

type routeMatch struct {
	routeMatch  *envoy_route.RouteMatch
	prefixMatch bool
}

// routeMatch returns a list of RouteMatches given a list of MeshHTTPRoute matches
// Note that some MeshHTTPRoute matches result in multiple Envoy matches because
// of prefix + rewrite handling. That's why we return the wrapper type as well.
func (c RoutesConfigurer) routeMatch(match api.Match) []routeMatch {
	var allEnvoyMatches []routeMatch

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
		if match.QueryParams != nil {
			routeQueryParamsMatch(envoyMatch, match.QueryParams)
		}
		routeHeadersMatch(envoyMatch, match.Headers)
		if match.Path != nil && match.Path.Type == api.PathPrefix {
			allEnvoyMatches = append(allEnvoyMatches, routeMatch{envoyMatch, true})
		} else {
			allEnvoyMatches = append(allEnvoyMatches, routeMatch{envoyMatch, false})
		}
	}

	return allEnvoyMatches
}

// Not every API match maps cleanly to a single envoy match
func (c RoutesConfigurer) routePathMatch(match api.PathMatch) []*envoy_route.RouteMatch {
	switch match.Type {
	case api.Exact:
		return []*envoy_route.RouteMatch{{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: match.Value,
			},
		}}
	case api.PathPrefix:
		if match.Value == "/" {
			return []*envoy_route.RouteMatch{{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{
					Prefix: match.Value,
				},
			}}
		}
		// This case forces us to create two different envoy matches to
		// replicate the "path-separated prefixes only" semantics.
		//
		// It needs to be replicated instead of using envoy's native
		// `path_separated_prefix` (https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routematch-path-separated-prefix)
		// because of an edge case, when if we want to drop the prefix:
		//
		//   ```
		//   - match:
		//       path_separated_prefix: "/prefix"
		//     route:
		//       prefix_rewrite: "/"
		//   ```
		//
		// results in the rewrites:
		//
		//   * `/prefix` -> `/`
		//   * `/prefix/path` -> `//path`
		//
		// ref. https://github.com/envoyproxy/envoy/issues/23130
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
		return []*envoy_route.RouteMatch{{
			PathSpecifier: &envoy_route.RouteMatch_SafeRegex{
				SafeRegex: regexMatcher(match.Value),
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

func routeQueryParamsMatch(envoyMatch *envoy_route.RouteMatch, matches []api.QueryParamsMatch) {
	// We ignore multiple matchers for the same name, though this is also
	// validated
	matchedNames := map[string]struct{}{}

	for _, match := range matches {
		if _, ok := matchedNames[match.Name]; ok {
			continue
		}
		matchedNames[match.Name] = struct{}{}

		var matcher envoy_type_matcher.StringMatcher
		switch match.Type {
		case api.ExactQueryMatch:
			matcher = envoy_type_matcher.StringMatcher{
				MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
					Exact: match.Value,
				},
			}
		case api.RegularExpressionQueryMatch:
			matcher = envoy_type_matcher.StringMatcher{
				MatchPattern: &envoy_type_matcher.StringMatcher_SafeRegex{
					SafeRegex: regexMatcher(match.Value),
				},
			}
		default:
			panic("impossible")
		}

		envoyMatch.QueryParameters = append(envoyMatch.QueryParameters,
			&envoy_route.QueryParameterMatcher{
				Name: match.Name,
				QueryParameterMatchSpecifier: &envoy_route.QueryParameterMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		)
	}
}

func (c RoutesConfigurer) hasExternal(split []envoy_common.Split) bool {
	for _, s := range split {
		if s.HasExternalService() {
			return true
		}
	}
	return false
}

func (c RoutesConfigurer) routeAction(split []envoy_common.Split) *envoy_route.RouteAction {
	routeAction := &envoy_route.RouteAction{
		// this timeout should be updated by the MeshTimeout plugin
		Timeout: util_proto.Duration(0),
	}
	if len(split) == 1 {
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: split[0].ClusterName(),
		}
	} else {
		var weightedClusters []*envoy_route.WeightedCluster_ClusterWeight
		for _, s := range split {
			weightedClusters = append(weightedClusters, &envoy_route.WeightedCluster_ClusterWeight{
				Name:   s.ClusterName(),
				Weight: util_proto.UInt32(s.Weight()),
			})
		}
		routeAction.ClusterSpecifier = &envoy_route.RouteAction_WeightedClusters{
			WeightedClusters: &envoy_route.WeightedCluster{
				Clusters: weightedClusters,
			},
		}
	}
	if c.hasExternal(split) {
		routeAction.HostRewriteSpecifier = &envoy_route.RouteAction_AutoHostRewrite{
			AutoHostRewrite: util_proto.Bool(true),
		}
	}
	return routeAction
}
