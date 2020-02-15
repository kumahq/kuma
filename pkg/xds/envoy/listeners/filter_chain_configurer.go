package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func FilterChain(builder *FilterChainBuilder) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&ListenerFilterChainConfigurer{
			builder: builder,
		})
	})
}

type ListenerFilterChainConfigurer struct {
	builder *FilterChainBuilder
}

func (c ListenerFilterChainConfigurer) Configure(listener *v2.Listener) error {
	filterChain, err := c.builder.Build()
	if err != nil {
		return err
	}
	listener.FilterChains = append(listener.FilterChains, filterChain)
	return nil
}
