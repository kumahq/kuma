package listeners

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type ListenerFilterChainConfigurerV2 struct {
	builder *FilterChainBuilder
}

func (c ListenerFilterChainConfigurerV2) Configure(listener *envoy_api_v2.Listener) error {
	filterChain, err := c.builder.Build()
	if err != nil {
		return err
	}
	listener.FilterChains = append(listener.FilterChains, filterChain.(*envoy_listener_v2.FilterChain))
	return nil
}

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
