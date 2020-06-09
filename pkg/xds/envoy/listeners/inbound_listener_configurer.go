package listeners

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	_struct "github.com/golang/protobuf/ptypes/struct"
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

func InboundListenerTLSInspector(listenerName string, address string, port uint32) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&InboundListenerConfigurer{
			listenerName: listenerName,
			address:      address,
			port:         port,
			tlsInspector: true,
		})
	})
}

type InboundListenerConfigurer struct {
	listenerName string
	address      string
	port         uint32
	tlsInspector bool
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
	if c.tlsInspector {
		l.ListenerFilters = append(l.ListenerFilters, &envoy_api_v2_listener.ListenerFilter{
			Name: wellknown.TlsInspector,
			ConfigType: &envoy_api_v2_listener.ListenerFilter_Config{
				Config: &_struct.Struct{},
			},
		})
	}
	// notice that filter chain configuration is left up to other configurers

	return nil
}
