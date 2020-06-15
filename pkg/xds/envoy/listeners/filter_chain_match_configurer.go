package listeners

import envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

func FilterChainMatch(serverNames ...string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&FilterChainMatchConfigurer{
			serverNames: serverNames,
		})
	})
}

type FilterChainMatchConfigurer struct {
	serverNames []string
}

func (f *FilterChainMatchConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		ServerNames: f.serverNames,
	}
	return nil
}
