package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

func buildMangleTable(
	cfg config.InitializedConfig,
	ipv6 bool,
) *tables.MangleTable {
	mangle := tables.Mangle()

	if cfg.ShouldDropInvalidPackets(ipv6) {
		mangle.Prerouting().AddRules(
			rules.
				NewRule(
					Match(Conntrack(Ctstate(INVALID))),
					Jump(Drop()),
				),
		)
	}

	return mangle
}
