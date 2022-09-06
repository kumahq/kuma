package v3

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

type RedirectConfigurer struct {
	MatchPath    string
	NewPath      string
	Port         uint32
	AllowGetOnly bool
}

func (c RedirectConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
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
		Action: &envoy_route.Route_Redirect{
			Redirect: &envoy_route.RedirectAction{
				PortRedirect: c.Port,
				PathRewriteSpecifier: &envoy_route.RedirectAction_PathRedirect{
					PathRedirect: c.NewPath,
				},
			},
		},
	})
	return nil
}
