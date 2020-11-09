package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func OutboundListener(listenerName string, protocol mesh_core.Protocol, address string, port uint32) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&OutboundListenerConfigurer{
			listenerName: listenerName,
			protocol:     protocol,
			address:      address,
			port:         port,
		})
	})
}

type OutboundListenerConfigurer struct {
	listenerName string
	protocol     mesh_core.Protocol
	address      string
	port         uint32
}

func (c *OutboundListenerConfigurer) Configure(l *v2.Listener) error {
	l.Name = c.listenerName
	l.TrafficDirection = envoy_core.TrafficDirection_OUTBOUND
	var listenerProtocol envoy_core.SocketAddress_Protocol
	switch c.protocol {
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
				Address:  c.address,
				PortSpecifier: &envoy_core.SocketAddress_PortValue{
					PortValue: c.port,
				},
			},
		},
	}
	// notice that filter chain configuration is left up to other configurers

	return nil
}
