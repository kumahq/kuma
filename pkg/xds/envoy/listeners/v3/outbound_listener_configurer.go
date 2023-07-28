package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type OutboundListenerConfigurer struct {
	Address  string
	Port     uint32
	Protocol core_xds.SocketAddressProtocol
}

func (c *OutboundListenerConfigurer) Configure(l *envoy_api.Listener) error {
	l.TrafficDirection = envoy_core.TrafficDirection_OUTBOUND
	l.Address = &envoy_core.Address{
		Address: &envoy_core.Address_SocketAddress{
			SocketAddress: &envoy_core.SocketAddress{
				Protocol: envoy_core.SocketAddress_Protocol(c.Protocol),
				Address:  c.Address,
				PortSpecifier: &envoy_core.SocketAddress_PortValue{
					PortValue: c.Port,
				},
			},
		},
	}
	// notice that filter chain configuration is left up to other configurers

	return nil
}
