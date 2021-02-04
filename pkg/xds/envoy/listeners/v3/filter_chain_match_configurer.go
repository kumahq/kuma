package v3

import envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

type FilterChainMatchConfigurer struct {
	ServerNames       []string
	TransportProtocol string
}

func (f *FilterChainMatchConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		ServerNames: f.ServerNames,
	}
	if f.TransportProtocol != "" {
		filterChain.FilterChainMatch.TransportProtocol = f.TransportProtocol
	}
	return nil
}
