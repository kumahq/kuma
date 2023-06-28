package v3

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type RouteConfigurer interface {
	Configure(*envoy_config_route_v3.Route) error
}

type RouteConfigureFunc func(*envoy_config_route_v3.Route) error

func (f RouteConfigureFunc) Configure(r *envoy_config_route_v3.Route) error {
	if f != nil {
		return f(r)
	}

	return nil
}

type RouteMustConfigureFunc func(*envoy_config_route_v3.Route)

func (f RouteMustConfigureFunc) Configure(r *envoy_config_route_v3.Route) error {
	if f != nil {
		f(r)
	}

	return nil
}
