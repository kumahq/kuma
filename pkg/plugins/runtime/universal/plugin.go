package universal

import (
	"github.com/kumahq/kuma/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/dns"
)

var _ core_plugins.RuntimePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	if rt.Config().Mode == core.Global {
		return nil
	}

	if err := addDNS(rt); err != nil {
		return err
	}

	return nil
}

func addDNS(rt core_runtime.Runtime) error {
	vipsAllocator, err := dns.NewVIPsAllocator(
		rt.ReadOnlyResourceManager(),
		rt.ConfigManager(),
		rt.Config().DNSServer.CIDR,
		rt.DNSResolver(),
	)
	if err != nil {
		return err
	}
	return rt.Add(vipsAllocator)
}
