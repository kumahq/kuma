package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

func buildMangleTable(cfg config.Config) (*tables.MangleTable, error) {
	mangle := tables.Mangle()

	mangle.Prerouting().
		AddRuleIf(cfg.ShouldDropInvalidPackets,
			Match(Conntrack(Ctstate(INVALID))),
			Jump(Drop()),
		)

	return mangle, nil
}
