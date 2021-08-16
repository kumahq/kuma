package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type SourceMatcherConfigurer struct {
	Address string
}

func (c *SourceMatcherConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	filterChain.FilterChainMatch = &envoy_listener.FilterChainMatch{
		SourcePrefixRanges: []*envoy_core.CidrRange{
			{
				AddressPrefix: c.Address,
				PrefixLen:     util_proto.UInt32(32),
			},
		},
	}
	return nil
}
