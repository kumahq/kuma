package universal

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/dns"
)

var _ core_plugins.RuntimePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	if rt.Config().Environment != core.UniversalEnvironment {
		return nil
	}
	if rt.Config().Mode == core.Global {
		return nil
	}

	if err := addDNS(rt); err != nil {
		return err
	}

	return nil
}

func addDNS(rt core_runtime.Runtime) error {
	zone := ""
	if rt.Config().Multizone != nil && rt.Config().Multizone.Zone != nil {
		zone = rt.Config().Multizone.Zone.Name
	}
	vipsAllocator, err := dns.NewVIPsAllocator(
		rt.ReadOnlyResourceManager(),
		rt.ConfigManager(),
		*rt.Config().DNSServer,
		rt.Config().Experimental,
		zone,
		rt.Metrics(),
	)
	if err != nil {
		return err
	}
	return rt.Add(component.LeaderComponentFunc(func(stop <-chan struct{}) error {
		ticker := time.NewTicker(rt.Config().Runtime.Universal.VIPRefreshInterval.Duration)
		defer ticker.Stop()

		dns.Log.Info("starting the DNS VIPs allocator")
		ctx := user.Ctx(context.Background(), user.ControlPlane)
		for {
			select {
			case <-ticker.C:
				if err := vipsAllocator.CreateOrUpdateVIPConfigs(ctx); err != nil {
					dns.Log.Error(err, "errors during updating VIP configs")
				}
			case <-stop:
				dns.Log.Info("stopping")
				return nil
			}
		}
	}))
}
