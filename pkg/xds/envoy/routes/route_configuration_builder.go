package routes

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

// RouteConfigurationConfigurer is responsible for configuring a single aspect of the entire Envoy RouteConfiguration,
// such as VirtualHost, HTTP headers to add or remove, etc.
type RouteConfigurationConfigurer interface {
	// Configure configures a single aspect on a given Envoy RouteConfiguration.
	Configure(routeConfiguration *v2.RouteConfiguration) error
}

// RouteConfigurationBuilderOpt is a configuration option for RouteConfigurationBuilder.
//
// The goal of RouteConfigurationBuilderOpt is to facilitate fluent RouteConfigurationBuilder API.
type RouteConfigurationBuilderOpt interface {
	// ApplyTo adds RouteConfigurationConfigurer(s) to the RouteConfigurationBuilder.
	ApplyTo(config *RouteConfigurationBuilderConfig)
}

func NewRouteConfigurationBuilder() *RouteConfigurationBuilder {
	return &RouteConfigurationBuilder{}
}

// RouteConfigurationBuilder is responsible for generating an Envoy RouteConfiguration
// by applying a series of RouteConfigurationConfigurers.
type RouteConfigurationBuilder struct {
	config RouteConfigurationBuilderConfig
}

// Configure configures RouteConfigurationBuilder by adding individual RouteConfigurationConfigurers.
func (b *RouteConfigurationBuilder) Configure(opts ...RouteConfigurationBuilderOpt) *RouteConfigurationBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy RouteConfiguration by applying a series of RouteConfigurationConfigurers.
func (b *RouteConfigurationBuilder) Build() (*v2.RouteConfiguration, error) {
	routeConfiguration := v2.RouteConfiguration{}
	for _, configurer := range b.config.Configurers {
		if err := configurer.Configure(&routeConfiguration); err != nil {
			return nil, err
		}
	}
	return &routeConfiguration, nil
}

// RouteConfigurationBuilderConfig holds configuration of a RouteConfigurationBuilder.
type RouteConfigurationBuilderConfig struct {
	// A series of RouteConfigurationConfigurers to apply to Envoy RouteConfiguration.
	Configurers []RouteConfigurationConfigurer
}

// Add appends a given RouteConfigurationConfigurer to the end of the chain.
func (c *RouteConfigurationBuilderConfig) Add(configurer RouteConfigurationConfigurer) {
	c.Configurers = append(c.Configurers, configurer)
}

// RouteConfigurationBuilderOptFunc is a convenience type adapter.
type RouteConfigurationBuilderOptFunc func(config *RouteConfigurationBuilderConfig)

func (f RouteConfigurationBuilderOptFunc) ApplyTo(config *RouteConfigurationBuilderConfig) {
	f(config)
}
