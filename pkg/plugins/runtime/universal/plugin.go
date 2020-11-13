package universal

import (
	"context"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns"
)

var (
	log = core.Log.WithName("plugin").WithName("runtime").WithName("universal")
)

var _ core_plugins.RuntimePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	rt.DNSResolver().SetVIPsChangedHandler(func(list dns.VIPList) {
		if err := UpdateOutbounds(context.Background(), rt.ResourceManager(), list); err != nil {
			log.Error(err, "failed to update VIP outbounds")
		}
	})

	if err := addDNS(rt); err != nil {
		return err
	}
	return nil
}

func addDNS(rt core_runtime.Runtime) error {
	ipam, err := dns.NewSimpleIPAM(rt.Config().DNSServer.CIDR)
	if err != nil {
		return err
	}
	p := dns.NewMeshedPersistence(rt.ConfigManager())
	vipsAllocator, err := dns.NewVIPsAllocator(rt.ReadOnlyResourceManager(), p, ipam, rt.DNSResolver())
	if err != nil {
		return err
	}
	vipsSync, err := dns.NewVIPsSynchronizer(rt.DNSResolver(), p, rt.LeaderInfo())
	if err != nil {
		return err
	}
	server, err := dns.NewDNSServer(rt.Config().DNSServer.Port, rt.DNSResolver(), rt.Metrics())
	if err != nil {
		return err
	}
	return rt.Add(server, vipsSync, vipsAllocator)
}
