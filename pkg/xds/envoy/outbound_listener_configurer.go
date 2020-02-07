package envoy

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
)

func OutboundListener(listenerName string, address string, port uint32) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&OutboundListenerConfigurer{
			listenerName: listenerName,
			address:      address,
			port:         port,
		})
	})
}

type OutboundListenerConfigurer struct {
	listenerName string
	address      string
	port         uint32
}

func (c *OutboundListenerConfigurer) Configure(l *v2.Listener) error {
	l.Name = c.listenerName
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
	l.FilterChains = []*envoy_listener.FilterChain{
		{}, // 1 filter chain that will be configured later on
	}

	return nil
}
