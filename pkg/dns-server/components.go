package dns_server

import (
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/dns-server/resolver"
	"github.com/Kong/kuma/pkg/dns-server/synchroniser"
	"strconv"
)

func SetupServer(rt runtime.Runtime) error {
	cfg := rt.Config()
	dnsResolver, err := resolver.NewSimpleDNSResolver(
		cfg.General.AdvertisedHostname,
		strconv.FormatUint(uint64(cfg.DNSServer.Port), 10),
		cfg.DNSServer.CIDR)
	if err != nil {
		return err
	}

	resourceSynchroniser, err := synchroniser.NewResourceSynchroniser(rt.ResourceManager(), dnsResolver)
	if err != nil {
		return err
	}

	return rt.Add(dnsResolver, resourceSynchroniser)
}
