package routes

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

type RouteBuilder struct {
	apiVersion  core_xds.APIVersion
	configurers []v3.RouteConfigurer
	name        string
}

func NewRouteBuilder(apiVersion core_xds.APIVersion, name string) *RouteBuilder {
	return &RouteBuilder{
		apiVersion: apiVersion,
		name:       name,
	}
}

func (r *RouteBuilder) Configure(opts ...v3.RouteConfigurer) *RouteBuilder {
	r.configurers = append(r.configurers, opts...)
	return r
}

func (r *RouteBuilder) Build() (envoy.NamedResource, error) {
	route := &envoy_config_route_v3.Route{
		Match: &envoy_config_route_v3.RouteMatch{},
		Name:  r.name,
	}

	for _, c := range r.configurers {
		if err := c.Configure(route); err != nil {
			return nil, err
		}
	}

	return route, nil
}
