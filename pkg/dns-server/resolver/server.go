package resolver

import (
	"sync"

	config_manager "github.com/Kong/kuma/pkg/core/config/manager"

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
		domain      string
		address     string
		cidr        string
		isLeader    bool
		viplist     VIPList
		persistence *DNSPersistence
		handler     DNSHandler
		ipam        IPAM
	}
	ElectedDNSResolver struct {
		resolver DNSResolver
	}
)

func (d *SimpleDNSResolver) NeedLeaderElection() bool {
	return false
}

func NewSimpleDNSResolver(domain, ip, port, cidr string, configm config_manager.ConfigManager) (DNSResolver, error) {
	resolver := &SimpleDNSResolver{
		domain:  domain,
		address: ip + ":" + port,
		cidr:    cidr,
		viplist: VIPList{},
	}

	resolver.handler = NewSimpleDNSHandler(resolver)
	resolver.ipam = NewSimpleIPAM(cidr)
	resolver.persistence = NewDNSPersistence(configm)

	viplist := resolver.persistence.Get()
	if viplist != nil {
		resolver.viplist = viplist
	}

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

	if !d.isLeader {
		return "", errors.Errorf("Can't add a service when not a leader")
	}

	_, found := d.viplist[service]
	if !found {
		ip, err := d.ipam.AllocateIP()
		if err != nil {
			return "", err
		}

		d.viplist[service] = ip
		d.persistence.Set(d.viplist)
	}

	return d.viplist[service], nil
}

func (d *SimpleDNSResolver) RemoveService(service string) error {
	d.Lock()
	defer d.Unlock()

	if !d.isLeader {
		return errors.Errorf("Can't remove a service when not a leader")
	}

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

	if !d.isLeader {
		d.viplist = d.persistence.Get()
		return nil
	}

	services = d.normalizeServiceMap(services)

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

	d.persistence.Set(d.viplist)
	return errs
}

func (d *SimpleDNSResolver) ForwardLookup(service string) (string, error) {
	d.RLock()
	defer d.RUnlock()

	ip, found := d.viplist[service]
	if !found && !d.isLeader {
		d.viplist = d.persistence.Get()
		ip, found = d.viplist[service]
	}

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
	if !found && !d.isLeader {
		d.viplist = d.persistence.Get()
		ip, found = d.viplist[service]
	}

	if !found {
		return "", errors.Errorf("service [%s] not found in domain [%s].", service, domain)
	}

	return ip, nil
}

func (d *SimpleDNSResolver) ReverseLookup(ip string) (string, error) {
	d.RLock()
	defer d.RUnlock()
	if !d.isLeader {
		d.viplist = d.persistence.Get()
	}

	for service, serviceIP := range d.viplist {
		if serviceIP == ip {
			return service + "." + d.domain, nil
		}
	}

	return "", errors.Errorf("IP [%s] not found", ip)
}

func (d *SimpleDNSResolver) normalizeServiceMap(services map[string]bool) map[string]bool {
	result := map[string]bool{}

	for name, value := range services {
		service, err := d.serviceFromName(name)
		if err != nil {
			simpleDNSLog.Error(err, "unable to map service name", "name", name)
		} else {
			result[service] = value
		}
	}

	return result
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

func (d *SimpleDNSResolver) SetElectedLeader(elected bool) {
	d.Lock()
	defer d.Unlock()
	simpleDNSLog.Info("DNS elected as a leader.")
	d.isLeader = elected
}

func NewElectedDNSResolver(resolver DNSResolver) (*ElectedDNSResolver, error) {
	return &ElectedDNSResolver{
		resolver: resolver,
	}, nil
}

func (e *ElectedDNSResolver) Start(stop <-chan struct{}) error {
	e.resolver.SetElectedLeader(true)
	return nil
}

func (e *ElectedDNSResolver) NeedLeaderElection() bool {
	return true
}
