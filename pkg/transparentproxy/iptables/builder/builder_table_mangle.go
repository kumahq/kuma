package builder

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/table"
)

<<<<<<< HEAD
func buildMangleTable(cfg config.Config) *table.MangleTable {
	mangle := table.Mangle()

	mangle.Prerouting().
		AppendIf(cfg.ShouldDropInvalidPackets,
			Match(Conntrack(Ctstate(INVALID))),
			Jump(Drop()),
		)
=======
func buildMangleTable(
	cfg config.InitializedConfig,
	ipv6 bool,
) *tables.MangleTable {
	mangle := tables.Mangle()

	if cfg.ShouldDropInvalidPackets(ipv6) {
		mangle.Prerouting().
			AddRule(
				Match(Conntrack(Ctstate(INVALID))),
				Jump(Drop()),
			)
	}
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))

	return mangle
}
