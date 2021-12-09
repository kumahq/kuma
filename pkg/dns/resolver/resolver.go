package resolver

import (
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/dns/vips"
)

type DNSResolver interface {
	GetDomain() string
	SetVIPs(map[vips.HostnameEntry]string)
	ForwardLookupFQDN(name string) (string, error)
}

type dnsResolver struct {
	sync.RWMutex
	domain  string
	viplist map[vips.HostnameEntry]string
}

var _ DNSResolver = &dnsResolver{}

func NewDNSResolver(domain string) DNSResolver {
	return &dnsResolver{
		domain: domain,
	}
}

func (d *dnsResolver) GetDomain() string {
	return d.domain
}

func (s *dnsResolver) SetVIPs(list map[vips.HostnameEntry]string) {
	s.Lock()
	defer s.Unlock()
	s.viplist = list
}

func (s *dnsResolver) ForwardLookupFQDN(name string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	ipFqdn, foundFqdn := s.viplist[vips.NewFqdnEntry(strings.TrimSuffix(name, "."))]

	domain, err := s.domainFromName(name)
	if err != nil {
		return "", err
	}

	if domain == s.domain {
		service, err := s.serviceFromName(name)
		if err != nil {
			return "", err
		}

		ip, found := s.viplist[vips.NewServiceEntry(service)]
		if found {
			return ip, nil
		} else if foundFqdn {
			return ipFqdn, nil
		}
		return "", errors.Errorf("service [%s] not found in domain [%s].", service, domain)
	} else if foundFqdn {
		return ipFqdn, nil
	}
	return "", errors.Errorf("domain [%s] not found.", domain)
}

func (s *dnsResolver) domainFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", errors.Errorf("wrong DNS name: %s", name)
	}

	return split[len(split)-1], nil
}

func (s *dnsResolver) serviceFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", errors.Errorf("wrong DNS name: %s", name)
	}

	// If it terminates with the domain we remove it.
	if split[len(split)-1] == s.domain {
		split = split[0 : len(split)-1]
	}

	return strings.Join(split, "."), nil
}
