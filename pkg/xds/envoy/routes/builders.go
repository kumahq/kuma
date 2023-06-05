package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type RouteConfigurer interface {
	Configure(*envoy_route.Route) error
}

type RouteBuilder struct {
	configurers []RouteConfigurer
}

func (r *RouteBuilder) Configure(configurers ...RouteConfigurer) *RouteBuilder {
	r.configurers = append(r.configurers, configurers...)

	return r
}

func (r *RouteBuilder) Build() (envoy.NamedResource, error) {
	route := &envoy_route.Route{
		Match: &envoy_route.RouteMatch{},
	}

	for _, c := range r.configurers {
		if err := c.Configure(route); err != nil {
			return nil, err
		}
	}

	return route, nil
}

type RouteConfigureFunc func(*envoy_route.Route) error

func (f RouteConfigureFunc) Configure(r *envoy_route.Route) error {
	if f != nil {
		return f(r)
	}

	return nil
}

type RouteMustConfigureFunc func(*envoy_route.Route)

func (f RouteMustConfigureFunc) Configure(r *envoy_route.Route) error {
	if f != nil {
		f(r)
	}

	return nil
}

type RouteRedirectConfigurer interface {
	Configure(redirect *envoy_route.RedirectAction) error
}

type RouteRedirectBuilder struct {
	configurers []RouteRedirectConfigurer
}

func (r *RouteRedirectBuilder) Configure(
	configurers ...RouteRedirectConfigurer,
) *RouteRedirectBuilder {
	r.configurers = append(r.configurers, configurers...)

	return r
}

func (r *RouteRedirectBuilder) Build() (*envoy_route.Route_Redirect, error) {
	routeRedirect := &envoy_route.Route_Redirect{
		Redirect: &envoy_route.RedirectAction{},
	}

	for _, c := range r.configurers {
		if err := c.Configure(routeRedirect.Redirect); err != nil {
			return nil, err
		}
	}

	return routeRedirect, nil
}

type RouteRedirectConfigureFunc func(redirect *envoy_route.RedirectAction) error

func (f RouteRedirectConfigureFunc) Configure(
	redirect *envoy_route.RedirectAction,
) error {
	if f != nil {
		return f(redirect)
	}

	return nil
}

type RouteRedirectMustConfigureFunc func(redirect *envoy_route.RedirectAction)

func (f RouteRedirectMustConfigureFunc) Configure(
	redirect *envoy_route.RedirectAction,
) error {
	if f != nil {
		f(redirect)
	}

	return nil
}
