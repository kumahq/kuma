package listeners

import (
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type ListenerFilterChainConfigurerV3 struct {
	builder *FilterChainBuilder
}

func (c ListenerFilterChainConfigurerV3) Configure(listener *envoy_listener_v3.Listener) error {
	filterChain, err := c.builder.Build()
	if err != nil {
		return err
	}
	listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener_v3.FilterChain))
	return nil
}
