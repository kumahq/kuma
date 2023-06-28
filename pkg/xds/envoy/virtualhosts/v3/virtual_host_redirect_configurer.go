package v3

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_routes_v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

type VirtualHostRedirectConfigurer struct {
	MatchPath    string
	NewPath      string
	Port         uint32
	AllowGetOnly bool
}

func (c VirtualHostRedirectConfigurer) Configure(virtualHost *envoy_config_route_v3.VirtualHost) error {
	rb := envoy_routes.NewRouteBuilder(envoy_common.APIV3, "")
	if c.AllowGetOnly {
		rb.Configure(envoy_routes_v3.RouteMatchExactHeader(":method", "GET"))
	}
	rb.Configure(envoy_routes_v3.RouteMatchExactPath(c.MatchPath),
		envoy_routes_v3.RouteMustConfigureFunc(func(envoyRoute *envoy_config_route_v3.Route) {
			envoyRoute.Action = &envoy_config_route_v3.Route_Redirect{
				Redirect: &envoy_config_route_v3.RedirectAction{
					PortRedirect: c.Port,
					PathRewriteSpecifier: &envoy_config_route_v3.RedirectAction_PathRedirect{
						PathRedirect: c.NewPath,
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
