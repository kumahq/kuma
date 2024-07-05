package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

func buildRawTable(cfg config.InitializedConfig, ipv6 bool) *tables.RawTable {
	raw := tables.Raw()

	dnsServers := cfg.Redirect.DNS.ServersIPv4
	conntractZoneSplit := cfg.Redirect.DNS.ConntrackZoneSplitIPv4
	if ipv6 {
		dnsServers = cfg.Redirect.DNS.ServersIPv6
		conntractZoneSplit = cfg.Redirect.DNS.ConntrackZoneSplitIPv6
	}

	if conntractZoneSplit {
		raw.Output().
			AddRule(
				Protocol(Udp(DestinationPort(DNSPort))),
				Match(Owner(Uid(cfg.Owner.UID))),
				Jump(Ct(Zone("1"))),
			).
			AddRule(
				Protocol(Udp(SourcePort(cfg.Redirect.DNS.Port))),
				Match(Owner(Uid(cfg.Owner.UID))),
				Jump(Ct(Zone("2"))),
			)

		if cfg.ShouldCaptureAllDNS() {
			raw.Output().AddRule(
				Protocol(Udp(DestinationPort(DNSPort))),
				Jump(Ct(Zone("2"))),
			)

			raw.Prerouting().
				AddRule(
					Protocol(Udp(SourcePort(DNSPort))),
					Jump(Ct(Zone("1"))),
				)
		} else {
			for _, ip := range dnsServers {
				raw.Output().AddRule(
					Destination(ip),
					Protocol(Udp(DestinationPort(DNSPort))),
					Jump(Ct(Zone("2"))),
				)
				raw.Prerouting().
					AddRule(
						Destination(ip),
						Protocol(Udp(SourcePort(DNSPort))),
						Jump(Ct(Zone("1"))),
					)
			}
		}
	}

	return raw
}
