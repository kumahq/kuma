package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type OutboundListenerConfigurer struct {
	ListenerName string
	Address      string
	Port         uint32
	Protocol     mesh_core.Protocol
}

func (c *OutboundListenerConfigurer) Configure(l *envoy_api.Listener) error {
	l.Name = c.ListenerName
	l.TrafficDirection = envoy_core.TrafficDirection_OUTBOUND
	var listenerProtocol envoy_core.SocketAddress_Protocol
	switch c.Protocol {
	case mesh_core.ProtocolUDP:
		listenerProtocol = envoy_core.SocketAddress_UDP
		l.ReusePort = true
	case mesh_core.ProtocolTCP:
		listenerProtocol = envoy_core.SocketAddress_TCP
	}
	l.Address = &envoy_core.Address{
		Address: &envoy_core.Address_SocketAddress{
			SocketAddress: &envoy_core.SocketAddress{
				Protocol: listenerProtocol,
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
