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

// FilterChainMustConfigureFunc adapts a FilterChain configuration function that
// never fails to the FilterChainConfigurer interface.
type FilterChainMustConfigureFunc func(chain *envoy_listener.FilterChain)

func (f FilterChainMustConfigureFunc) Configure(chain *envoy_listener.FilterChain) error {
	if f != nil {
		f(chain)
	}

	return nil
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
