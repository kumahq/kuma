package v3

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_routes_v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

type VirtualHostRouteConfigurer struct {
	MatchPath    string
	NewPath      string
	Cluster      string
	AllowGetOnly bool
}

func (c VirtualHostRouteConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	rb := envoy_routes.NewRouteBuilder(envoy_common.APIV3, "")
	if c.AllowGetOnly {
		rb.Configure(envoy_routes_v3.RouteMatchExactHeader(":method", "GET"))
	}
	rb.Configure(envoy_routes_v3.RouteMatchExactPath(c.MatchPath),
		envoy_routes_v3.RouteMustConfigureFunc(func(envoyRoute *envoy_config_route_v3.Route) {
			envoyRoute.Action = &envoy_config_route_v3.Route_Route{
				Route: &envoy_config_route_v3.RouteAction{
					RegexRewrite: &envoy_type_matcher.RegexMatchAndSubstitute{
						Pattern: &envoy_type_matcher.RegexMatcher{
							EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
								GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
							},
							Regex: `.*`,
						},
						Substitution: c.NewPath,
					},
					ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
						Cluster: c.Cluster,
					},
				},
			}
		}))
	r, err := rb.Build()
	if err != nil {
		return err
	}
	virtualHost.Routes = append(virtualHost.Routes, r.(*envoy_config_route_v3.Route))
	return nil
}
