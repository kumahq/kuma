package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
)

// VirtualHostConfigurer is responsible for configuring a single aspect of the entire Envoy VirtualHost,
// such as Route, CORS, etc.
type VirtualHostConfigurer interface {
	// Configure configures a single aspect on a given Envoy VirtualHost.
	Configure(virtualHost *envoy_route.VirtualHost) error
}

// VirtualHostBuilderOpt is a configuration option for VirtualHostBuilder.
//
// The goal of VirtualHostBuilderOpt is to facilitate fluent VirtualHostBuilder API.
type VirtualHostBuilderOpt interface {
	// ApplyTo adds VirtualHostConfigurer(s) to the VirtualHostBuilder.
	ApplyTo(config *VirtualHostBuilderConfig)
}

func NewVirtualHostBuilder() *VirtualHostBuilder {
	return &VirtualHostBuilder{}
}

// VirtualHostBuilder is responsible for generating an Envoy VirtualHost
// by applying a series of VirtualHostConfigurers.
type VirtualHostBuilder struct {
	config VirtualHostBuilderConfig
}

// Configure configures VirtualHostBuilder by adding individual VirtualHostConfigurers.
func (b *VirtualHostBuilder) Configure(opts ...VirtualHostBuilderOpt) *VirtualHostBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy VirtualHost by applying a series of VirtualHostConfigurers.
func (b *VirtualHostBuilder) Build() (*envoy_route.VirtualHost, error) {
	virtualHost := envoy_route.VirtualHost{}
	for _, configurer := range b.config.Configurers {
		if err := configurer.Configure(&virtualHost); err != nil {
			return nil, err
		}
	}
	return &virtualHost, nil
}

// VirtualHostBuilderConfig holds configuration of a VirtualHostBuilder.
type VirtualHostBuilderConfig struct {
	// A series of VirtualHostConfigurers to apply to Envoy VirtualHost.
	Configurers []VirtualHostConfigurer
}

// Add appends a given VirtualHostConfigurer to the end of the chain.
func (c *VirtualHostBuilderConfig) Add(configurer VirtualHostConfigurer) {
	c.Configurers = append(c.Configurers, configurer)
}

// VirtualHostBuilderOptFunc is a convenience type adapter.
type VirtualHostBuilderOptFunc func(config *VirtualHostBuilderConfig)

func (f VirtualHostBuilderOptFunc) ApplyTo(config *VirtualHostBuilderConfig) {
	f(config)
}
