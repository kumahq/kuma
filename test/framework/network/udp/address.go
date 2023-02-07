package udp

import (
	"net"

	"github.com/kumahq/kuma/test/framework/network/ip"
)

// GenRandomAddressIPv4 will generate random *net.UDPAddr (IPv4) with provided port
func GenRandomAddressIPv4(port uint16) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   ip.GenRandomIPv4(),
		Port: int(port),
	}
}

// GenRandomAddressIPv6 will generate random *net.UDPAddr (IPv6) with provided port
func GenRandomAddressIPv6(port uint16) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   ip.GenRandomIPv6(),
		Port: int(port),
	}
}
