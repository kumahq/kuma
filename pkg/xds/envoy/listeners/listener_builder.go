package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

// ListenerConfigurer is responsible for configuring a single aspect of the entire Envoy listener,
// such as filter chain, transparent proxying, etc.
type ListenerConfigurer interface {
	// Configure configures a single aspect on a given Envoy listener.
	Configure(listener *v2.Listener) error
}

// ListenerBuilderOpt is a configuration option for ListenerBuilder.
//
// The goal of ListenerBuilderOpt is to facilitate fluent ListenerBuilder API.
type ListenerBuilderOpt interface {
	// ApplyTo adds ListenerConfigurer(s) to the ListenerBuilder.
	ApplyTo(config *ListenerBuilderConfig)
}

func NewListenerBuilder() *ListenerBuilder {
	return &ListenerBuilder{}
}

// ListenerBuilder is responsible for generating an Envoy listener
// by applying a series of ListenerConfigurers.
type ListenerBuilder struct {
	config ListenerBuilderConfig
}

// Configure configures ListenerBuilder by adding individual ListenerConfigurers.
func (b *ListenerBuilder) Configure(opts ...ListenerBuilderOpt) *ListenerBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy listener by applying a series of ListenerConfigurers.
func (b *ListenerBuilder) Build() (*v2.Listener, error) {
	listener := v2.Listener{}
	for _, configurer := range b.config.Configurers {
		if err := configurer.Configure(&listener); err != nil {
			return nil, err
		}
	}
	return &listener, nil
}

// ListenerBuilderConfig holds configuration of a ListenerBuilder.
type ListenerBuilderConfig struct {
	// A series of ListenerConfigurers to apply to Envoy listener.
	Configurers []ListenerConfigurer
}

// Add appends a given ListenerConfigurer to the end of the chain.
func (c *ListenerBuilderConfig) Add(configurer ListenerConfigurer) {
	c.Configurers = append(c.Configurers, configurer)
}

// ListenerBuilderOptFunc is a convenience type adapter.
type ListenerBuilderOptFunc func(config *ListenerBuilderConfig)

func (f ListenerBuilderOptFunc) ApplyTo(config *ListenerBuilderConfig) {
	f(config)
}
