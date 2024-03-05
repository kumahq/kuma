package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type AdditionalAddressConfigurer struct {
	Addresses []mesh_proto.OutboundInterface
}

func (c *AdditionalAddressConfigurer) Configure(l *listenerv3.Listener) error {
	if len(c.Addresses) < 1 || l.Address == nil {
		return nil
	}

	var addresses []*listenerv3.AdditionalAddress
	for _, addr := range c.Addresses {
		address := makeSocketAddress(addr.DataplaneIP, addr.DataplanePort, l.Address.GetSocketAddress().GetProtocol())
		addresses = append(addresses, address)
	}
	l.AdditionalAddresses = addresses
	return nil
}

func makeSocketAddress(addr string, port uint32, protocol envoy_core.SocketAddress_Protocol) *listenerv3.AdditionalAddress {
	return &listenerv3.AdditionalAddress{
		Address: &envoy_core.Address{
			Address: &envoy_core.Address_SocketAddress{
				SocketAddress: &envoy_core.SocketAddress{
					Protocol: protocol,
					Address:  addr,
					PortSpecifier: &envoy_core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
	}
}
