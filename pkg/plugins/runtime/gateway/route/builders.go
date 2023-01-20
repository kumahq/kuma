package route

import (
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type RouteConfigurer interface {
	Configure(*envoy_config_route.Route) error
}

type RouteBuilder struct {
	configurers []RouteConfigurer
}

func (r *RouteBuilder) Configure(opts ...RouteConfigurer) *RouteBuilder {
	r.configurers = append(r.configurers, opts...)
	return r
}

func (r *RouteBuilder) Build() (envoy.NamedResource, error) {
	route := &envoy_config_route.Route{
		Match: &envoy_config_route.RouteMatch{},
	}

	for _, c := range r.configurers {
		if err := c.Configure(route); err != nil {
			return nil, err
		}
	}

	return route, nil
}

type RouteConfigureFunc func(*envoy_config_route.Route) error

func (f RouteConfigureFunc) Configure(r *envoy_config_route.Route) error {
	if f != nil {
		return f(r)
	}

	return nil
}

type RouteMustConfigureFunc func(*envoy_config_route.Route)

func (f RouteMustConfigureFunc) Configure(r *envoy_config_route.Route) error {
	if f != nil {
		f(r)
	}

	return nil
}
