package net

import (
	"fmt"
	"net"
	"sort"
)

// GetAllIPs returns all IPs (IPv4 and IPv6) from the all network interfaces on the machine
func GetAllIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("could not list network interfaces: %w", err)
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

// ToV6 return self if ip6 other return the v4 prefixed with ::ffff:
func ToV6(ip string) string {
	parsedIp := net.ParseIP(ip)
	if parsedIp.To4() != nil {
		return fmt.Sprintf("::ffff:%x:%x", uint32(parsedIp[12])<<8+uint32(parsedIp[13]), uint32(parsedIp[14])<<8+uint32(parsedIp[15]))
	}
	return ip
}
