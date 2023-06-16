package v3

import envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

type FilterChainNameConfigurer struct {
	Name string
}

func (f *FilterChainNameConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.Name = f.Name
	return nil
}
