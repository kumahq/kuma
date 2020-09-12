package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func OutboundListener(listenerName string, address string, port uint32, isUDP bool) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&OutboundListenerConfigurer{
			listenerName: listenerName,
			address:      address,
			port:         port,
			isUDP:        isUDP,
		})
	})
}

type OutboundListenerConfigurer struct {
	listenerName string
	address      string
	port         uint32
	isUDP        bool
}

func (c *OutboundListenerConfigurer) Configure(l *v2.Listener) error {
	l.Name = c.listenerName
	l.TrafficDirection = envoy_core.TrafficDirection_OUTBOUND
	listenerProtocol := envoy_core.SocketAddress_TCP
	if c.isUDP {
		listenerProtocol = envoy_core.SocketAddress_UDP
		l.ReusePort = true
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
