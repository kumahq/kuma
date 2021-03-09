package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type UpstreamBindConfigConfigurer struct {
	Address string
	Port    uint32
}

var _ ClusterConfigurer = &UpstreamBindConfigConfigurer{}

func (u *UpstreamBindConfigConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.UpstreamBindConfig = &envoy_core.BindConfig{
		SourceAddress: &envoy_core.SocketAddress{
			Address: u.Address,
			PortSpecifier: &envoy_core.SocketAddress_PortValue{
				PortValue: u.Port,
			},
		},
	}
	return nil
}
