package v2

import envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

type FilterChainMatchConfigurer struct {
	ServerNames       []string
	TransportProtocol string
}

func (f *FilterChainMatchConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		ServerNames:       f.ServerNames,
		TransportProtocol: f.TransportProtocol,
	}
	return nil
}
