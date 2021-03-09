package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
)

// ListenerConfigurer is responsible for configuring a single aspect of the entire Envoy listener,
// such as filter chain, transparent proxying, etc.
type ListenerConfigurer interface {
	// Configure configures a single aspect on a given Envoy listener.
	Configure(listener *envoy_api.Listener) error
}

// FilterChainConfigurer is responsible for configuring a single aspect of the entire Envoy filter chain,
// such as TcpProxy filter, RBAC filter, access log, etc.
type FilterChainConfigurer interface {
	// Configure configures a single aspect on a given Envoy filter chain.
	Configure(filterChain *envoy_listener.FilterChain) error
}
