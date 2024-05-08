package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/table"
)

func buildMangleTable(cfg config.Config) (*table.MangleTable, error) {
	mangle := table.Mangle()

	mangle.Prerouting().
		AddRuleIf(cfg.ShouldDropInvalidPackets,
			Match(Conntrack(Ctstate(INVALID))),
			Jump(Drop()),
		)

	return mangle, nil
}
