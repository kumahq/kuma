package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type SourceMatcherConfigurer struct {
	Address string
}

func (c *SourceMatcherConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		SourcePrefixRanges: []*envoy_core.CidrRange{
			{
				AddressPrefix: c.Address,
				PrefixLen:     &wrappers.UInt32Value{Value: 32},
			},
		},
	}
	return nil
}
