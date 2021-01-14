package v3_test

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

func EnvoySocketAddress(address string, port uint32) *envoy_core.Address {
	return &envoy_core.Address{
		Address: &envoy_core.Address_SocketAddress{
			SocketAddress: &envoy_core.SocketAddress{
				Address: address,
				PortSpecifier: &envoy_core.SocketAddress_PortValue{
					PortValue: port,
				},
			},
		},
	}
}
