package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/table"
)

<<<<<<< HEAD
func buildRawTable(
	cfg config.Config,
	dnsServers []string,
) *table.RawTable {
	raw := table.Raw()

	if cfg.ShouldConntrackZoneSplit() {
=======
func buildRawTable(cfg config.InitializedConfig, ipv6 bool) *tables.RawTable {
	raw := tables.Raw()

	dnsServers := cfg.Redirect.DNS.ServersIPv4
	if ipv6 {
		dnsServers = cfg.Redirect.DNS.ServersIPv6
	}

	conntractZoneSplit := cfg.Redirect.DNS.ConntrackZoneSplitIPv4
	if ipv6 {
		conntractZoneSplit = cfg.Redirect.DNS.ConntrackZoneSplitIPv6
	}

	if conntractZoneSplit {
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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
