package v3

import envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

type FilterChainMatchConfigurer struct {
	ServerNames          []string
	TransportProtocol    string
	ApplicationProtocols []string
}

func (f *FilterChainMatchConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		ServerNames: f.ServerNames,
	}
	if f.TransportProtocol != "" {
		filterChain.FilterChainMatch.TransportProtocol = f.TransportProtocol
	}
	if len(f.ApplicationProtocols) != 0 {
		filterChain.FilterChainMatch.ApplicationProtocols = f.ApplicationProtocols
	}
	return nil
}
