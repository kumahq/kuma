package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type InboundListenerConfigurer struct {
	Protocol     core_xds.SocketAddressProtocol
	ListenerName string
	Address      string
	Port         uint32
}

func (c *InboundListenerConfigurer) Configure(l *envoy_listener.Listener) error {
	l.Name = c.ListenerName
	l.EnableReusePort = wrapperspb.Bool(c.Protocol == core_xds.SocketAddressProtocolUDP)
	l.TrafficDirection = envoy_core.TrafficDirection_INBOUND
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
