package dns_server

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

func (d *SimpleDNSResolver) domainFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", fmt.Errorf("Wrong DNS name: %s", name)
	}
	return split[len(split)-1], nil
}

func (d *SimpleDNSResolver) serviceFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 2 {
		return "", fmt.Errorf("Wrong DNS name: %s", name)
	}

	service := strings.Join(split[:len(split)-1], ".")
	if service == "" {
		return "", fmt.Errorf("Wrong service in DNS name: %s", name)
	}

	return service, nil
}
