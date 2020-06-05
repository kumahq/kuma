package dns_server

import (
	"strconv"

	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/dns-server/resolver"
	"github.com/Kong/kuma/pkg/dns-server/synchronizer"
)

const topLevelDomain = ".kuma"

func SetupServer(rt runtime.Runtime) error {
	cfg := rt.Config()

	dnsResolver, err := resolver.NewSimpleDNSResolver(
		cfg.General.AdvertisedHostname,
		strconv.FormatUint(uint64(cfg.DNSServer.Port), 10),
		cfg.DNSServer.CIDR)
	if err != nil {
		return err
	}

	resourceSynchronizer, err := synchronizer.NewResourceSynchronizer(topLevelDomain, rt.ResourceManager(), dnsResolver)
	if err != nil {
		return err
	}

	return rt.Add(dnsResolver, resourceSynchronizer)
}
