package dns

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt runtime.Runtime) error {
	vipsSync := NewVIPsSynchronizer(
		rt.DNSResolver(),
		rt.ReadOnlyResourceManager(),
		rt.ConfigManager(),
		rt.LeaderInfo(),
	)
	if err := rt.Add(vipsSync); err != nil {
		return err
	}

	if rt.Config().DNSServer.Enabled {
		if server, err := NewDNSServer(
			rt.Config().DNSServer.Port,
			rt.DNSResolver(),
			rt.Metrics(),
			DnsNameToKumaCompliant,
		); err != nil {
			return err
		} else {
			return rt.Add(server)
		}
	}
	return nil
}
