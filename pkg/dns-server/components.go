package dns_server

import (
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/dns-server/resolver"
	"github.com/Kong/kuma/pkg/dns-server/synchronizer"
)

func SetupServer(rt runtime.Runtime) error {
	resourceSynchronizer, err := synchronizer.NewResourceSynchronizer(rt.ReadOnlyResourceManager(), rt.DNSResolver())
	if err != nil {
		return err
	}

	electedResolver, err := resolver.NewElectedDNSResolver(rt.DNSResolver())
	if err != nil {
		return err
	}

	return rt.Add(rt.DNSResolver(), electedResolver, resourceSynchronizer)
}
