package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

// buildRawTable constructs the raw table for iptables with the necessary rules
// for handling DNS traffic and connection tracking zone splitting. This table
// is configured based on the provided configuration, including handling IPv4
// and IPv6 DNS servers
func buildRawTable(cfg config.InitializedConfigIPvX) *tables.RawTable {
	raw := tables.Raw()

	if cfg.Redirect.DNS.ConntrackZoneSplit {
		raw.Output().AddRules(
			rules.
				NewAppendRule(
					Protocol(Udp(DestinationPort(consts.DNSPort))),
					Match(Owner(Uid(cfg.KumaDPUser))),
					Jump(Ct(Zone("1"))),
				).
				WithCommentf("assign connection tracking zone 1 to DNS traffic from the kuma-dp user (UID %s)", cfg.KumaDPUser),
			rules.
				NewAppendRule(
					Protocol(Udp(SourcePort(cfg.Redirect.DNS.Port))),
					Match(Owner(Uid(cfg.KumaDPUser))),
					Jump(Ct(Zone("2"))),
				).
				WithComment("assign connection tracking zone 2 to DNS responses from the kuma-dp DNS proxy"),
		)

		if cfg.Redirect.DNS.CaptureAll {
			raw.Output().AddRules(
				rules.
					NewAppendRule(
						Protocol(Udp(DestinationPort(consts.DNSPort))),
						Jump(Ct(Zone("2"))),
					).
					WithComment("assign connection tracking zone 2 to all DNS requests"),
			)

			raw.Prerouting().AddRules(
				rules.
					NewAppendRule(
						Protocol(Udp(SourcePort(consts.DNSPort))),
						Jump(Ct(Zone("1"))),
					).
					WithComment("assign connection tracking zone 1 to all DNS responses"),
			)
		} else {
			for _, ip := range cfg.Redirect.DNS.Servers {
				raw.Output().AddRules(
					rules.
						NewAppendRule(
							Destination(ip),
							Protocol(Udp(DestinationPort(consts.DNSPort))),
							Jump(Ct(Zone("2"))),
						).
						WithCommentf("assign connection tracking zone 2 to DNS requests destined for %s", ip.String()),
				)

				// IsLoopback checks if the address is local (e.g., 127.0.0.1 or 127.0.0.11, which is common
				// in Docker containers within custom networks). This rule addresses an issue where the
				// transparent proxy is installed in a Docker container that's part of a custom network.
				// In such cases, Docker NATs the destination port to a random one. This rule is applied
				// only when the source is a local address and the destination is localhost to prevent
				// unexpected behavior in untested scenarios
				if ip.IsLoopback() {
					raw.Output().AddRules(
						rules.
							NewAppendRule(
								Source(ip),
								Destination(cfg.LocalhostCIDR),
								Protocol(Udp()),
								Jump(Ct(Zone("1"))),
							).
							WithComment("assign conntrack zone 1 to DNS responses from the local DNS server to localhost, needed when the DNS query port is altered by a DNAT iptables rule, such as with Docker containers in a custom network"),
					)
				}

				raw.Prerouting().AddRules(
					rules.
						NewAppendRule(
							Source(ip),
							Protocol(Udp(SourcePort(consts.DNSPort))),
							Jump(Ct(Zone("1"))),
						).
						WithCommentf("assign connection tracking zone 1 to DNS responses from %s", ip.String()),
				)
			}
		}
	}

	return raw
}
