package routes

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

// RouteConfigurationBuilderOpt is a configuration option for RouteConfigurationBuilder.
//
// The goal of RouteConfigurationBuilderOpt is to facilitate fluent RouteConfigurationBuilder API.
type RouteConfigurationBuilderOpt interface {
	// ApplyTo adds RouteConfigurationConfigurer(s) to the RouteConfigurationBuilder.
	ApplyTo(builder *RouteConfigurationBuilder)
}

func NewRouteConfigurationBuilder(apiVersion core_xds.APIVersion, name string) *RouteConfigurationBuilder {
	return &RouteConfigurationBuilder{
		apiVersion: apiVersion,
		name:       name,
	}
}

// RouteConfigurationBuilder is responsible for generating an Envoy RouteConfiguration
// by applying a series of RouteConfigurationConfigurers.
type RouteConfigurationBuilder struct {
	apiVersion  core_xds.APIVersion
	configurers []v3.RouteConfigurationConfigurer
	name        string
}

// Configure configures RouteConfigurationBuilder by adding individual RouteConfigurationConfigurers.
func (b *RouteConfigurationBuilder) Configure(opts ...RouteConfigurationBuilderOpt) *RouteConfigurationBuilder {
	for _, opt := range opts {
		opt.ApplyTo(b)
	}
	return b
}

// Build generates an Envoy RouteConfiguration by applying a series of RouteConfigurationConfigurers.
func (b *RouteConfigurationBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		routeConfiguration := envoy_config_route_v3.RouteConfiguration{
			Name: b.name,
		}
		for _, configurer := range b.configurers {
			if err := configurer.Configure(&routeConfiguration); err != nil {
				return nil, err
			}
		}
		if routeConfiguration.GetName() == "" {
			return nil, errors.New("route configuration name is undefined")
		}
		return &routeConfiguration, nil
	default:
		return nil, errors.New("unknown API")
	}
}

// AddConfigurer appends a given RouteConfigurationConfigurer to the end of the chain.
func (b *RouteConfigurationBuilder) AddConfigurer(configurer v3.RouteConfigurationConfigurer) {
	b.configurers = append(b.configurers, configurer)
}

// RouteConfigurationBuilderOptFunc is a convenience type adapter.
type RouteConfigurationBuilderOptFunc func(builder *RouteConfigurationBuilder)

func (f RouteConfigurationBuilderOptFunc) ApplyTo(builder *RouteConfigurationBuilder) {
	f(builder)
}

// AddRouteConfigurationConfigurer produces an option that adds the given
// configurer to the route configuration builder.
func AddRouteConfigurationConfigurer(c v3.RouteConfigurationConfigurer) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(builder *RouteConfigurationBuilder) {
		builder.AddConfigurer(c)
	})
}
