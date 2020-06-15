package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func InboundListener(listenerName string, address string, port uint32) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&InboundListenerConfigurer{
			listenerName: listenerName,
			address:      address,
			port:         port,
		})
	})
}

type InboundListenerConfigurer struct {
	listenerName string
	address      string
	port         uint32
}

func (c *InboundListenerConfigurer) Configure(l *v2.Listener) error {
	l.Name = c.listenerName
	l.TrafficDirection = envoy_core.TrafficDirection_INBOUND
	l.Address = &envoy_core.Address{
		Address: &envoy_core.Address_SocketAddress{
			SocketAddress: &envoy_core.SocketAddress{
				Protocol: envoy_core.SocketAddress_TCP,
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
