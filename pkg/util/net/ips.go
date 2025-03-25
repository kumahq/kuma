package net

import (
	"fmt"
	"net"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

type AddressPredicate = func(address *net.IPNet) bool

func NonLoopback(address *net.IPNet) bool {
	return !address.IP.IsLoopback()
}

// GetAllIPs returns all IPs (IPv4 and IPv6) from the all network interfaces on the machine
func GetAllIPs(predicates ...AddressPredicate) ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.Wrap(err, "could not list network interfaces")
	}
	var result []string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			matchedPredicate := true
			for _, predicate := range predicates {
				if !predicate(ipnet) {
					matchedPredicate = false
					break
				}
			}
			if matchedPredicate {
				result = append(result, ipnet.IP.String())
			}
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

func JoinHostPort(h string, p uint32) string {
	return net.JoinHostPort(h, strconv.Itoa(int(p)))
}
