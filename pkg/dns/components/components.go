package components

import (
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/dns"
)

func SetupServer(rt runtime.Runtime) error {
	server := dns.NewDNSServer(rt.Config().DNSServer.Port, rt.DNSResolver())
	persistence := dns.NewDNSPersistence(rt.ConfigManager())
	vipsSync, err := dns.NewVIPsSynchronizer(rt.ReadOnlyResourceManager(), rt.DNSResolver(), persistence, rt.LeaderInfo())
	if err != nil {
		return err
	}
	ipam, err := dns.NewSimpleIPAM(rt.Config().DNSServer.CIDR)
	if err != nil {
		return err
	}
	vipsAllocator, err := dns.NewVIPsAllocator(rt.ReadOnlyResourceManager(), persistence, ipam, rt.DNSResolver())
	if err != nil {
		return err
	}

	return rt.Add(server, vipsAllocator, vipsSync)
}
