package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/table"
)

func buildRawTable(
	cfg config.Config,
	dnsServers []string,
) *table.RawTable {
	raw := table.Raw()

	if cfg.ShouldConntrackZoneSplit() {
		raw.Output().
			Append(
				Protocol(Udp(DestinationPort(DNSPort))),
				Match(Owner(Uid(cfg.Owner.UID))),
				Jump(Ct(Zone("1"))),
			).
			Append(
				Protocol(Udp(SourcePort(cfg.Redirect.DNS.Port))),
				Match(Owner(Uid(cfg.Owner.UID))),
				Jump(Ct(Zone("2"))),
			)

		if cfg.ShouldCaptureAllDNS() {
			raw.Output().Append(
				Protocol(Udp(DestinationPort(DNSPort))),
				Jump(Ct(Zone("2"))),
			)

			raw.Prerouting().
				Append(
					Protocol(Udp(SourcePort(DNSPort))),
					Jump(Ct(Zone("1"))),
				)
		} else {
			for _, ip := range dnsServers {
				raw.Output().Append(
					Destination(ip),
					Protocol(Udp(DestinationPort(DNSPort))),
					Jump(Ct(Zone("2"))),
				)
				raw.Prerouting().
					Append(
						Destination(ip),
						Protocol(Udp(SourcePort(DNSPort))),
						Jump(Ct(Zone("1"))),
					)
			}
		}
	}

	return raw
}
