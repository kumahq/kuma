package builder

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

func GetDnsServers(cfgPath string) ([]string, []string, error) {
	dnsConfig, err := dns.ClientConfigFromFile(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read file %s: %s", cfgPath, err)
	}
	ipv4, ipv6 := groupIps(dnsConfig.Servers)
	return ipv4, ipv6, nil
}

func groupIps(addresses []string) ([]string, []string) {
	var ipv4 []string
	var ipv6 []string
	for _, address := range addresses {
		parsed := net.ParseIP(address)
		if parsed.To4() != nil {
			ipv4 = append(ipv4, address)
		} else {
			ipv6 = append(ipv6, address)
		}
	}
	return ipv4, ipv6
}
