package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

type UpstreamBindConfigConfigurer struct {
	Address string
	Port    uint32
}

var _ ClusterConfigurer = &UpstreamBindConfigConfigurer{}

func (u *UpstreamBindConfigConfigurer) Configure(c *envoy_api.Cluster) error {
	c.UpstreamBindConfig = &envoy_core.BindConfig{
		SourceAddress: &envoy_core.SocketAddress{
			Address: u.Address,
			PortSpecifier: &envoy_core.SocketAddress_PortValue{
				PortValue: u.Port,
			},
			Ipv4Compat: true,
		},
	}
	return nil
}
