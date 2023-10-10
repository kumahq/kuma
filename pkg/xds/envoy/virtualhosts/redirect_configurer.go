package virtualhosts

import (
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type RedirectConfigurer struct {
	MatchPath    string
	NewPath      string
	Port         uint32
	AllowGetOnly bool
}

func (c RedirectConfigurer) Configure(virtualHost *envoy_config_route.VirtualHost) error {
	rb := envoy_routes.NewRouteBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
		Configure(envoy_routes.RouteMatchExactPath(c.MatchPath)).
		Configure(envoy_routes.RouteMustConfigureFunc(func(envoyRoute *envoy_config_route.Route) {
			envoyRoute.Action = &envoy_config_route.Route_Redirect{
				Redirect: &envoy_config_route.RedirectAction{
					PortRedirect: c.Port,
					PathRewriteSpecifier: &envoy_config_route.RedirectAction_PathRedirect{
						PathRedirect: c.NewPath,
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
