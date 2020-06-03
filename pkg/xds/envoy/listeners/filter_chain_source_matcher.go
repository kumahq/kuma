package listeners

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func SourceMatcher(address string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&SourceMatcherConfigurer{
			address: address,
		})
	})
}

type SourceMatcherConfigurer struct {
	address string
}

func (c *SourceMatcherConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		SourcePrefixRanges: []*envoy_core.CidrRange{
			{
				AddressPrefix: c.address,
				PrefixLen:     &wrappers.UInt32Value{Value: 32},
			},
		},
	}
	return nil
}
