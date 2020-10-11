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
	return nil
}
