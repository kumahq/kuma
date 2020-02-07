package envoy

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

// ListenerConfigurer is responsible for configuring a single aspect of the entire Envoy listener,
// such as TcpProxy filter, RBAC filter, access log, etc.
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

// ListenerBuilder is responsible for generating an Envoy listener
// by applying a series of ListenerConfigurers.
type ListenerBuilder interface {
	// Configure configures ListenerBuilder by adding individual ListenerConfigurers.
	Configure(opts ...ListenerBuilderOpt) ListenerBuilder
	// Build generates an Envoy listener by applying a series of ListenerConfigurers.
	Build() (*v2.Listener, error)
}

func NewListenerBuilder() ListenerBuilder {
	return &listenerBuilder{}
}

// ListenerBuilderConfig holds configuration of a ListenerBuilder.
type ListenerBuilderConfig struct {
	// A series of ListenerConfigurers to apply to Envoy listener.
	Configurers []ListenerConfigurer
}

func (c *ListenerBuilderConfig) Add(configurer ListenerConfigurer) {
	c.Configurers = append(c.Configurers, configurer)
}

type listenerBuilder struct {
	config ListenerBuilderConfig
}

// Configure configures ListenerBuilder by adding individual ListenerConfigurers.
func (b *listenerBuilder) Configure(opts ...ListenerBuilderOpt) ListenerBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy listener by applying a series of ListenerConfigurers.
func (b *listenerBuilder) Build() (*v2.Listener, error) {
	listener := v2.Listener{}
	for _, configurer := range b.config.Configurers {
		if err := configurer.Configure(&listener); err != nil {
			return nil, err
		}
	}
	return &listener, nil
}

// ListenerBuilderOptFunc is a convenience type adapter.
type ListenerBuilderOptFunc func(config *ListenerBuilderConfig)

func (f ListenerBuilderOptFunc) ApplyTo(config *ListenerBuilderConfig) {
	f(config)
}
