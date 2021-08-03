package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

// ListenerConfigurer is responsible for configuring a single aspect of the entire Envoy listener,
// such as filter chain, transparent proxying, etc.
type ListenerConfigurer interface {
	// Configure configures a single aspect on a given Envoy listener.
	Configure(listener *envoy_listener.Listener) error
}

// FilterChainConfigurer is responsible for configuring a single aspect of the entire Envoy filter chain,
// such as TcpProxy filter, RBAC filter, access log, etc.
type FilterChainConfigurer interface {
	// Configure configures a single aspect on a given Envoy filter chain.
	Configure(filterChain *envoy_listener.FilterChain) error
}

// ListenerConfigureFunc adapts a configuration function to the
// ListenerConfigurer interface.
type ListenerConfigureFunc func(listener *envoy_listener.Listener) error

func (f ListenerConfigureFunc) Configure(listener *envoy_listener.Listener) error {
	if f != nil {
		return f(listener)
	}

	return nil
}

// ListenerMustConfigureFunc adapts a configuration function that never
// fails to the ListenerConfigurer interface.
type ListenerMustConfigureFunc func(listener *envoy_listener.Listener)

func (f ListenerMustConfigureFunc) Configure(listener *envoy_listener.Listener) error {
	if f != nil {
		f(listener)
	}

	return nil
}
