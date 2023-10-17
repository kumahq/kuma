package routes

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type RouteConfigurer interface {
	Configure(*envoy_config_route_v3.Route) error
}

type RouteBuilder struct {
	apiVersion  core_xds.APIVersion
	configurers []RouteConfigurer
	name        string
}

func NewRouteBuilder(apiVersion core_xds.APIVersion, name string) *RouteBuilder {
	return &RouteBuilder{
		apiVersion: apiVersion,
		name:       name,
	}
}

func (r *RouteBuilder) Configure(opts ...RouteConfigurer) *RouteBuilder {
	r.configurers = append(r.configurers, opts...)
	return r
}

func (r *RouteBuilder) Build() (envoy.NamedResource, error) {
	route := &envoy_config_route_v3.Route{
		Match:                &envoy_config_route_v3.RouteMatch{},
		Name:                 r.name,
		TypedPerFilterConfig: map[string]*anypb.Any{},
	}

	for _, c := range r.configurers {
		if err := c.Configure(route); err != nil {
			return nil, err
		}
	}

	return route, nil
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
