package tcp

import (
	"net"

	"github.com/kumahq/kuma/test/framework/network/ip"
)

// GenRandomAddressIPv4 will generate random *net.TCPAddr (IPv4) with provided port
func GenRandomAddressIPv4(port uint16) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   ip.GenRandomIPv4(),
		Port: int(port),
	}
}

// GenRandomAddressIPv6 will generate random *net.TCPAddr (IPv6) with provided port
func GenRandomAddressIPv6(port uint16) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   ip.GenRandomIPv6(),
		Port: int(port),
	}
}
