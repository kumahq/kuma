package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
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

// FilterChainConfigureFunc adapts a FilterChain configuration function to the
// FilterChainConfigurer interface.
type FilterChainConfigureFunc func(chain *envoy_listener.FilterChain) error

func (f FilterChainConfigureFunc) Configure(chain *envoy_listener.FilterChain) error {
	if f != nil {
		return f(chain)
	}

	return nil
}

// FilterChainConfigureFunc adapts a FilterChain configuration function that
// never fails to the FilterChainConfigurer interface.
type FilterChainMustConfigureFunc func(chain *envoy_listener.FilterChain)

func (f FilterChainMustConfigureFunc) Configure(chain *envoy_listener.FilterChain) error {
	if f != nil {
		f(chain)
	}

	return nil
}

// HttpConnectionManagerConfigureFunc adapts a HttpConnectionManager
// configuration function to the FilterChainConfigurer interface.
type HttpConnectionManagerConfigureFunc func(hcm *envoy_hcm.HttpConnectionManager) error

func (f HttpConnectionManagerConfigureFunc) Configure(filterChain *envoy_listener.FilterChain) error {
	if f != nil {
		return UpdateHTTPConnectionManager(filterChain, f)
	}

	return nil
}

// HttpConnectionManagerMustConfigureFunc adapts a HttpConnectionManager
// configuration function that never fails to the FilterChainConfigurer
// interface.
type HttpConnectionManagerMustConfigureFunc func(hcm *envoy_hcm.HttpConnectionManager)

func (f HttpConnectionManagerMustConfigureFunc) Configure(filterChain *envoy_listener.FilterChain) error {
	if f != nil {
		return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
			f(hcm)
			return nil
		})
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

// ListenerMustConfigureFunc adapts a configuration function that never
// fails to the ListenerConfigurer interface.
type ListenerMustConfigureFunc func(listener *envoy_listener.Listener)

func (f ListenerMustConfigureFunc) Configure(listener *envoy_listener.Listener) error {
	if f != nil {
		f(listener)
	}

	return nil
}
