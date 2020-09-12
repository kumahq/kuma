package v2

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

type InboundListenerConfigurer struct {
	ListenerName string
	Address      string
	Port         uint32
}

func (c *InboundListenerConfigurer) Configure(l *v2.Listener) error {
	l.Name = c.ListenerName
	l.TrafficDirection = envoy_core.TrafficDirection_INBOUND
	listenerProtocol := envoy_core.SocketAddress_TCP
	if c.isUDP {
		listenerProtocol = envoy_core.SocketAddress_UDP
		l.ReusePort = true
	}
	l.Address = &envoy_core.Address{
		Address: &envoy_core.Address_SocketAddress{
			SocketAddress: &envoy_core.SocketAddress{
				Protocol: envoy_core.SocketAddress_TCP,
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
