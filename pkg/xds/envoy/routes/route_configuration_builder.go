package routes

import (
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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
	ApplyTo(config *RouteConfigurationBuilderConfig)
}

func NewRouteConfigurationBuilder(apiVersion core_xds.APIVersion) *RouteConfigurationBuilder {
	return &RouteConfigurationBuilder{
		apiVersion: apiVersion,
	}
}

// RouteConfigurationBuilder is responsible for generating an Envoy RouteConfiguration
// by applying a series of RouteConfigurationConfigurers.
type RouteConfigurationBuilder struct {
	apiVersion core_xds.APIVersion
	config     RouteConfigurationBuilderConfig
}

// Configure configures RouteConfigurationBuilder by adding individual RouteConfigurationConfigurers.
func (b *RouteConfigurationBuilder) Configure(opts ...RouteConfigurationBuilderOpt) *RouteConfigurationBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy RouteConfiguration by applying a series of RouteConfigurationConfigurers.
func (b *RouteConfigurationBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		routeConfiguration := envoy_route_v3.RouteConfiguration{}
		for _, configurer := range b.config.ConfigurersV3 {
			if err := configurer.Configure(&routeConfiguration); err != nil {
				return nil, err
			}
		}
		return &routeConfiguration, nil
	default:
		return nil, errors.New("unknown API")
	}
}

// RouteConfigurationBuilderConfig holds configuration of a RouteConfigurationBuilder.
type RouteConfigurationBuilderConfig struct {
	// A series of RouteConfigurationConfigurers to apply to Envoy RouteConfiguration.
	ConfigurersV3 []v3.RouteConfigurationConfigurer
}

// Add appends a given RouteConfigurationConfigurer to the end of the chain.
func (c *RouteConfigurationBuilderConfig) AddV3(configurer v3.RouteConfigurationConfigurer) {
	c.ConfigurersV3 = append(c.ConfigurersV3, configurer)
}

// RouteConfigurationBuilderOptFunc is a convenience type adapter.
type RouteConfigurationBuilderOptFunc func(config *RouteConfigurationBuilderConfig)

func (f RouteConfigurationBuilderOptFunc) ApplyTo(config *RouteConfigurationBuilderConfig) {
	f(config)
}

// AddRouteConfigurationConfigurer produces an option that adds the given
// configurer to the route coonfiguration builder.
func AddRouteConfigurationConfigurer(c v3.RouteConfigurationConfigurer) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.AddV3(c)
	})
}
