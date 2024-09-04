package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

func buildMangleTable(cfg config.InitializedConfigIPvX) *tables.MangleTable {
	mangle := tables.Mangle()

	if cfg.DropInvalidPackets {
		mangle.Prerouting().AddRules(
			rules.
				NewAppendRule(
					Match(Conntrack(Ctstate(INVALID))),
					Jump(Drop()),
				).
				WithComment("drop packets in the INVALID state to prevent potential issues with malformed or out-of-order packets"),
		)
	}

	return mangle
}
