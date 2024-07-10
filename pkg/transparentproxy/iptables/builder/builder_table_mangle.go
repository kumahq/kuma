package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

// buildMangleTable constructs the mangle table for iptables with the necessary
// rules for handling invalid packets. This table is configured based on the
// provided configuration and handles both IPv4 and IPv6 traffic.
func buildMangleTable(cfg config.InitializedConfigIPvX) *tables.MangleTable {
	mangle := tables.Mangle()

	if cfg.DropInvalidPackets {
		mangle.Prerouting().AddRules(
			rules.
				NewRule(
					Match(Conntrack(Ctstate(INVALID))),
					Jump(Drop()),
				).
				WithComment("drop packets in the INVALID state to prevent potential issues with malformed or out-of-order packets"),
		)
	}

	return mangle
}
