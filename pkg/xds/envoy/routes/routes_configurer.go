package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func Route(matchPath, newPath, cluster string, allowGetOnly bool) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.Add(&RoutesConfigurer{
			matchPath:    matchPath,
			newPath:      newPath,
			cluster:      cluster,
			allowGetOnly: allowGetOnly,
		})
	})
}

type RoutesConfigurer struct {
	matchPath    string
	newPath      string
	cluster      string
	allowGetOnly bool
}

func (c RoutesConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	var headersMatcher []*envoy_route.HeaderMatcher
	if c.allowGetOnly {
		headersMatcher = []*envoy_route.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_ExactMatch{
					ExactMatch: "GET",
				},
			},
		}
	}
	virtualHost.Routes = append(virtualHost.Routes, &envoy_route.Route{
		Match: &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: c.matchPath,
			},
			Headers: headersMatcher,
		},
		Action: &envoy_route.Route_Route{
			Route: &envoy_route.RouteAction{
				RegexRewrite: &envoy_type_matcher.RegexMatchAndSubstitute{
					Pattern: &envoy_type_matcher.RegexMatcher{
						EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{
								MaxProgramSize: &wrappers.UInt32Value{Value: 500},
							},
						},
						Regex: `.*`,
					},
					Substitution: c.newPath,
				},
				ClusterSpecifier: &envoy_route.RouteAction_Cluster{
					Cluster: c.cluster,
				},
			},
		},
	})
	return nil
}
