package resolver

import (
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/miekg/dns"

	"github.com/Kong/kuma/pkg/core"
)

var (
	simpleDNSLog = core.Log.WithName("dns-server-resolver")
)

type (
	VIPList           map[string]string
	SimpleDNSResolver struct {
		sync.RWMutex
		domain  string
		address string
		cidr    string
		viplist VIPList
		handler DNSHandler
		ipam    IPAM
	}
)

func (d *SimpleDNSResolver) NeedLeaderElection() bool {
	return false
}

func NewSimpleDNSResolver(domain, ip, port, cidr string) (DNSResolver, error) {
	resolver := &SimpleDNSResolver{
		domain:  domain,
		address: ip + ":" + port,
		cidr:    cidr,
		viplist: VIPList{},
	}

	resolver.handler = NewSimpleDNSHandler(resolver)
	resolver.ipam = NewSimpleIPAM(cidr)

	return resolver, nil
}

func (d *SimpleDNSResolver) Start(stop <-chan struct{}) error {
	server := &dns.Server{
		Addr: d.address,
		Net:  "udp",
	}

	errChan := make(chan error)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			simpleDNSLog.Error(err, "failed to start the DNS listener.")
			errChan <- err
		}
	}()

	simpleDNSLog.Info("starting", "address", d.address, "CIDR", d.cidr)
	select {
	case <-stop:
		simpleDNSLog.Info("shutting down the DNS Server")
		return server.Shutdown()
	case err := <-errChan:
		return err
	}
}

func (d *SimpleDNSResolver) GetDomain() string {
	return d.domain
}

func (d *SimpleDNSResolver) AddService(service string) (string, error) {
	d.Lock()
	defer d.Unlock()

	_, found := d.viplist[service]
	if !found {
		ip, err := d.ipam.AllocateIP()
		if err != nil {
			return "", err
		}

		d.viplist[service] = ip
	}

	return d.viplist[service], nil
}

func (d *SimpleDNSResolver) RemoveService(service string) error {
	d.Lock()
	defer d.Unlock()

	ip, found := d.viplist[service]
	if !found {
		return errors.Errorf("service [%s] not found in domain [%s].", service, d.domain)
	}

	delete(d.viplist, service)

	err := d.ipam.FreeIP(ip)
	if err != nil {
		return err
	}

	return nil
}

func (d *SimpleDNSResolver) SyncServices(services map[string]bool) (errs error) {
	d.Lock()
	defer d.Unlock()

	// ensure all services have entries in the domain
	for service := range services {
		_, found := d.viplist[service]
		if !found {
			ip, err := d.ipam.AllocateIP()
			if err != nil {
				errs = multierr.Append(errs, errors.Wrapf(err, "unable to allocate an ip for service %s", service))
			} else {
				d.viplist[service] = ip
				simpleDNSLog.Info("Adding ", "service", service, "ip", ip)
			}
		}
	}

	// ensure all entries in the domain are present in the service list, and delete them otherwise
	for service := range d.viplist {
		_, found := services[service]
		if !found {
			_ = d.ipam.FreeIP(d.viplist[service])
			delete(d.viplist, service)
		}
	}

	return errs
}

func (d *SimpleDNSResolver) ForwardLookup(service string) (string, error) {
	d.RLock()
	defer d.RUnlock()

	ip, found := d.viplist[service]
	if !found {
		return "", errors.Errorf("service [%s] not found in domain [%s].", service, d.domain)
	}
	return ip, nil
}

func (d *SimpleDNSResolver) ForwardLookupFQDN(name string) (string, error) {
	d.RLock()
	defer d.RUnlock()
	domain, err := d.domainFromName(name)
	if err != nil {
		return "", err
	}

	if domain != d.domain {
		return "", errors.Errorf("domain [%s] not found.", domain)
	}

	service, err := d.serviceFromName(name)
	if err != nil {
		return "", err
	}

	ip, found := d.viplist[service]
	if !found {
		return "", errors.Errorf("service [%s] not found in domain [%s].", service, domain)
	}

	return ip, nil
}

func (d *SimpleDNSResolver) ReverseLookup(ip string) (string, error) {
	d.RLock()
	defer d.RUnlock()
	for service, serviceIP := range d.viplist {
		if serviceIP == ip {
			return service + "." + d.domain, nil
		}
	}

	return "", errors.Errorf("IP [%s] not found", ip)
}

func (d *SimpleDNSResolver) domainFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", errors.Errorf("wrong DNS name: %s", name)
	}

	return split[len(split)-1], nil
}

func (d *SimpleDNSResolver) serviceFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", errors.Errorf("wrong DNS name: %s", name)
	}

	service := split[0]

	return service, nil
}
