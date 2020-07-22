package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func UpstreamBindConfig(address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&upstreamBindConfigConfigurer{
			address: address,
			port:    port,
		})
	})
}

type upstreamBindConfigConfigurer struct {
	address string
	port    uint32
}

func (u *upstreamBindConfigConfigurer) Configure(c *envoy_api.Cluster) error {
	c.UpstreamBindConfig = &envoy_core.BindConfig{
		SourceAddress: &envoy_core.SocketAddress{
			Address: u.address,
			PortSpecifier: &envoy_core.SocketAddress_PortValue{
				PortValue: u.port,
			},
		},
	}
	return nil
}
