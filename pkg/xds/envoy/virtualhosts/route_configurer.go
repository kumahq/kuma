package virtualhosts

import (
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type VirtualHostRouteConfigurer struct {
	MatchPath    string
	NewPath      string
	Cluster      string
	AllowGetOnly bool
}

func (c VirtualHostRouteConfigurer) Configure(virtualHost *envoy_config_route.VirtualHost) error {
	rb := envoy_routes.NewRouteBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
		Configure(envoy_routes.RouteMatchExactPath(c.MatchPath)).
		Configure(envoy_routes.RouteMustConfigureFunc(func(envoyRoute *envoy_config_route.Route) {
			envoyRoute.Action = &envoy_config_route.Route_Route{
				Route: &envoy_config_route.RouteAction{
					RegexRewrite: &envoy_type_matcher.RegexMatchAndSubstitute{
						Pattern: &envoy_type_matcher.RegexMatcher{
							EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
								GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
							},
							Regex: `.*`,
						},
						Substitution: c.NewPath,
					},
					ClusterSpecifier: &envoy_config_route.RouteAction_Cluster{
						Cluster: c.Cluster,
					},
				},
			}
		}))
	if c.AllowGetOnly {
		rb.Configure(envoy_routes.RouteMatchExactMethod("GET"))
	}
	route, err := rb.Build()
	if err != nil {
		return err
	}
	virtualHost.Routes = append(virtualHost.Routes, route)
	return nil
}
