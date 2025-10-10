package config

import (
	"net"

	"github.com/pkg/errors"
)

// HasLocalIPv6 checks if the local system has an active non-loopback IPv6
// address. It scans through all network interfaces to find any IPv6 address
// that is not a loopback address.
func HasLocalIPv6() (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() == nil {
			return true, nil
		}
	}

	return false, errors.New("no local IPv6 addresses detected")
}
