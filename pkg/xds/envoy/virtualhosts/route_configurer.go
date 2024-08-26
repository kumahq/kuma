package virtualhosts

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type VirtualHostRouteConfigurer struct {
	MatchPath    string
	NewPath      string
	Cluster      string
	AllowGetOnly bool
}

func (c VirtualHostRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	var headersMatcher []*envoy_config_route_v3.HeaderMatcher
	if c.AllowGetOnly {
		matcher := envoy_type_matcher_v3.StringMatcher{
			MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
				Exact: "GET",
			},
		}
		headersMatcher = []*envoy_config_route_v3.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_config_route_v3.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		}
	}
	virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
		Match: &envoy_config_route_v3.RouteMatch{
			PathSpecifier: &envoy_config_route_v3.RouteMatch_Path{
				Path: c.MatchPath,
			},
			Headers: headersMatcher,
		},
		Name: envoy_common.AnonymousResource,
		Action: &envoy_config_route_v3.Route_Route{
			Route: &envoy_config_route_v3.RouteAction{
				RegexRewrite: &envoy_type_matcher_v3.RegexMatchAndSubstitute{
					Pattern: &envoy_type_matcher_v3.RegexMatcher{
						Regex: `.*`,
					},
					Substitution: c.NewPath,
				},
				ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
					Cluster: c.Cluster,
				},
			},
		},
	})
	return nil
}

type VirtualHostBasicRouteConfigurer struct {
	Cluster string
}

func (c VirtualHostBasicRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
		Match: &envoy_config_route_v3.RouteMatch{
			PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &envoy_config_route_v3.Route_Route{
			Route: &envoy_config_route_v3.RouteAction{
				ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
					Cluster: c.Cluster,
				},
			},
		},
	})
	return nil
}

type VirtualHostDirectResponseRouteConfigurer struct {
	status      uint32
	responseMsg string
}

func (c VirtualHostDirectResponseRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	virtualHost.Routes = append(virtualHost.Routes, &envoy_config_route_v3.Route{
		Match: &envoy_config_route_v3.RouteMatch{
			PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &envoy_config_route_v3.Route_DirectResponse{
			DirectResponse: &envoy_config_route_v3.DirectResponseAction{
				Status: c.status,
				Body: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{
						InlineString: c.responseMsg,
					},
				},
			},
		},
	})
	return nil
}
