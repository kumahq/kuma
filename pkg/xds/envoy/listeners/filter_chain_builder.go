package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
)

// FilterChainConfigurer is responsible for configuring a single aspect of the entire Envoy filter chain,
// such as TcpProxy filter, RBAC filter, access log, etc.
type FilterChainConfigurer interface {
	// Configure configures a single aspect on a given Envoy filter chain.
	Configure(filterChain *envoy_listener.FilterChain) error
}

// FilterChainBuilderOpt is a configuration option for FilterChainBuilder.
//
// The goal of FilterChainBuilderOpt is to facilitate fluent FilterChainBuilder API.
type FilterChainBuilderOpt interface {
	// ApplyTo adds FilterChainConfigurer(s) to the FilterChainBuilder.
	ApplyTo(config *FilterChainBuilderConfig)
}

func NewFilterChainBuilder() *FilterChainBuilder {
	return &FilterChainBuilder{}
}

// FilterChainBuilder is responsible for generating an Envoy filter chain
// by applying a series of FilterChainConfigurers.
type FilterChainBuilder struct {
	config FilterChainBuilderConfig
}

// Configure configures FilterChainBuilder by adding individual FilterChainConfigurers.
func (b *FilterChainBuilder) Configure(opts ...FilterChainBuilderOpt) *FilterChainBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy filter chain by applying a series of FilterChainConfigurers.
func (b *FilterChainBuilder) Build() (*envoy_listener.FilterChain, error) {
	filterChain := envoy_listener.FilterChain{}
	for _, configurer := range b.config.Configurers {
		if err := configurer.Configure(&filterChain); err != nil {
			return nil, err
		}
	}
	return &filterChain, nil
}

// FilterChainBuilderConfig holds configuration of a FilterChainBuilder.
type FilterChainBuilderConfig struct {
	// A series of FilterChainConfigurers to apply to Envoy filter chain.
	Configurers []FilterChainConfigurer
}

// Add appends a given FilterChainConfigurer to the end of the chain.
func (c *FilterChainBuilderConfig) Add(configurer FilterChainConfigurer) {
	c.Configurers = append(c.Configurers, configurer)
}

// FilterChainBuilderOptFunc is a convenience type adapter.
type FilterChainBuilderOptFunc func(config *FilterChainBuilderConfig)

func (f FilterChainBuilderOptFunc) ApplyTo(config *FilterChainBuilderConfig) {
	f(config)
}
