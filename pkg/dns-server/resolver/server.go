package resolver

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"

	"github.com/Kong/kuma/pkg/core"
)

var (
	simpleDNSLog = core.Log.WithName("dns-server-resolver")
)

type (
	VIPList           map[string]string
	DomainList        map[string]VIPList
	SimpleDNSResolver struct {
		address string
		domains DomainList
		ipam    *IPAM
	}
)

func NewSimpleDNSResolver(ip, port, cidr string) (DNSResolver, error) {
	return newSimpleDNSResolver(ip, port, cidr)
}

func newSimpleDNSResolver(ip, port, cidr string) (*SimpleDNSResolver, error) {
	resolver := &SimpleDNSResolver{
		address: ip + ":" + port,
		domains: DomainList{},
	}

	err := resolver.initIPAM(cidr)
	if err != nil {
		simpleDNSLog.Error(err, "Unale to init the IPAM module in the DNS resolver.")
		return nil, err
	}

	return resolver, nil
}

func (d *SimpleDNSResolver) Start(stop <-chan struct{}) error {
	// handlers for all domains
	d.registerDNSHandlers()

	server := &dns.Server{
		Addr: d.address,
		Net:  "udp",
	}

	errChan := make(chan error)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			simpleDNSLog.Error(err, "Failed to start the DNS listener.")
			errChan <- err
		}
	}()

	simpleDNSLog.Info("starting", "address", d.address, "CIDR", d.ipam.CIDR)
	select {
	case <-stop:
		simpleDNSLog.Info("Shutting down the DNS Server")
		return server.Shutdown()
	case err := <-errChan:
		return err
	}
}

func (d *SimpleDNSResolver) AddDomain(domain string) error {
	domain, err := d.domainFromName(domain)
	if err != nil {
		return err
	}

	_, found := d.domains[domain]
	if !found {
		d.domains[domain] = VIPList{}
	}

	return nil
}

func (d *SimpleDNSResolver) RemoveDomain(domain string) error {
	domain, err := d.domainFromName(domain)
	if err != nil {
		return err
	}

	_, found := d.domains[domain]
	if !found {
		return fmt.Errorf("Deleting domain [%s] not found.", domain)
	}

	delete(d.domains, domain)

	return nil
}

func (d *SimpleDNSResolver) AddServiceToDomain(service string, domain string) (string, error) {
	entry, found := d.domains[domain]
	if !found {
		return "", fmt.Errorf("Domain [%s] not found.", domain)
	}

	_, found = entry[service]
	if !found {
		ip, err := d.allocateIP()
		if err != nil {
			return "", err
		}

		entry[service] = ip
	}

	return entry[service], nil
}

func (d *SimpleDNSResolver) RemoveServiceFromDomain(service string, domain string) error {
	entry, found := d.domains[domain]
	if !found {
		return fmt.Errorf("Domain [%s] not found.", domain)
	}

	ip, found := entry[service]
	if !found {
		return fmt.Errorf("Service [%s] not found in domain [%s].", service, domain)
	}

	delete(entry, service)

	err := d.freeIP(ip)
	if err != nil {
		return err
	}

	return nil
}

func (d *SimpleDNSResolver) SyncServicesForDomain(services map[string]bool, domain string) error {
	entry, found := d.domains[domain]
	if !found {
		return fmt.Errorf("Domain [%s] not found.", domain)
	}

	errors := []string{}
	// ensure all services have entries in the domain
	for service := range services {
		_, found = entry[service]
		if !found {
			ip, err := d.allocateIP()
			if err != nil {
				errors = append(errors, fmt.Sprintf("unable to allocate an ip for service %s [%v]", service, err))
			} else {
				entry[service] = ip
			}
		}
	}

	// ensure all entries in the domain are present in the service list, and delete them otherwise
	for service := range entry {
		_, found := services[service]
		if !found {
			delete(entry, service)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, ","))
	}

	return nil
}

func (d *SimpleDNSResolver) ForwardLookup(name string) (string, error) {
	domain, err := d.domainFromName(name)
	if err != nil {
		return "", err
	}

	entry, found := d.domains[domain]
	if !found {
		return "", fmt.Errorf("Domain [%s] not found.", domain)
	}

	service, err := d.serviceFromName(name)
	if err != nil {
		return "", err
	}

	ip, found := entry[service]
	if !found {
		return "", fmt.Errorf("Service [%s] not found in domain [%s].", service, domain)
	}

	return ip, nil
}

func (d *SimpleDNSResolver) ReverseLookup(ip string) (string, error) {
	for domain, entry := range d.domains {
		for service, serviceIP := range entry {
			if serviceIP == ip {
				return service + "." + domain, nil
			}
		}
	}

	return "", fmt.Errorf("IP [%s] not found", ip)
}
