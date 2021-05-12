package net

import (
	"fmt"
	"net"
	"sort"
	"strings"

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

// IsV4 return true if an ip is v4
func IsV4(ip string) bool {
	if strings.HasPrefix(ip, ":") {
		return false
	}
	parsedIp := net.ParseIP(ip)
	return parsedIp.To4() != nil
}

// ToV6 return self if ip6 other return the v4 prefixed with ::ffff:
func ToV6(ip string) string {
	parsedIp := net.ParseIP(ip)
	if parsedIp.To4() != nil {
		return fmt.Sprintf("::ffff:%x:%x", uint32(parsedIp[12])<<8+uint32(parsedIp[13]), uint32(parsedIp[14])<<8+uint32(parsedIp[15]))
	}
	return ip
}

func CidrIsIpV4(cidr string) bool {
	firstIp, _, _ := net.ParseCIDR(cidr)
	return firstIp.To4() != nil
}
