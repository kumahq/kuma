package dns

import (
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt runtime.Runtime) error {
	if rt.Config().Mode == config_core.Global {
		return nil
	}
	vipsSync := NewVIPsSynchronizer(
		rt.DNSResolver(),
		rt.ReadOnlyResourceManager(),
		rt.ConfigManager(),
		rt.LeaderInfo(),
	)
	if err := rt.Add(vipsSync); err != nil {
		return err
	}

	server, err := NewDNSServer(
		rt.Config().DNSServer.Port,
		rt.DNSResolver(),
		rt.Metrics(),
		DnsNameToKumaCompliant,
	)
	if err != nil {
		return err
	}
	return rt.Add(server)
}
