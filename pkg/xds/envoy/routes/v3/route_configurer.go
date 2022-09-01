package v3

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

type RouteConfigurer struct {
	MatchPath    string
	NewPath      string
	Cluster      string
	AllowGetOnly bool
}

func (c RouteConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	var headersMatcher []*envoy_route.HeaderMatcher
	if c.AllowGetOnly {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: "GET",
			},
		}
		headersMatcher = []*envoy_route.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		}
	}
	virtualHost.Routes = append(virtualHost.Routes, &envoy_route.Route{
		Match: &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: c.MatchPath,
			},
			Headers: headersMatcher,
		},
		Action: &envoy_route.Route_Route{
			Route: &envoy_route.RouteAction{
				RegexRewrite: &envoy_type_matcher.RegexMatchAndSubstitute{
					Pattern: &envoy_type_matcher.RegexMatcher{
						EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
						},
						Regex: `.*`,
					},
					Substitution: c.NewPath,
				},
				ClusterSpecifier: &envoy_route.RouteAction_Cluster{
					Cluster: c.Cluster,
				},
			},
		},
	})
	return nil
}
