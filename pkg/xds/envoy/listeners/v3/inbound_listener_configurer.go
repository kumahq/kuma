package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type InboundListenerConfigurer struct {
	ListenerName string
	Address      string
	Port         uint32
}

func (c *InboundListenerConfigurer) Configure(l *envoy_listener.Listener) error {
	l.Name = c.ListenerName
	l.TrafficDirection = envoy_core.TrafficDirection_INBOUND
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
