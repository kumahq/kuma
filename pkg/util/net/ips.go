package net

import (
	"net"
	"sort"

	"github.com/pkg/errors"
)

// GetAllIPs returns all IPs (IPv4 and IPv6) from the all network interfaces on the machine
func GetAllIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.Wrap(err, "could not list network interfaces")
	}
	var result []string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			result = append(result, ipnet.IP.String())
		}
	}
	sort.Strings(result) // sort so IPv4 are the first elements in the list
	return result, nil
}
