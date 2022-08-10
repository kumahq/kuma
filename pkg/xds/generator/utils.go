package generator

import "net"

func IsAddressIPv6(address string) bool {
	if address == "" {
		return false
	}

	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}

	return ip.To4() == nil
}
