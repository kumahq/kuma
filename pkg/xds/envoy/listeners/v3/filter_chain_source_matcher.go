package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type SourceMatcherConfigurer struct {
	Address string
}

func (c *SourceMatcherConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		SourcePrefixRanges: []*envoy_core.CidrRange{
			{
				AddressPrefix: c.Address,
				PrefixLen:     &wrapperspb.UInt32Value{Value: 32},
			},
		},
	}
	return nil
}
