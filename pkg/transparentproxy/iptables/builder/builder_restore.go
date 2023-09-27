package builder

import (
	"os"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

func buildRestore(
	cfg config.Config,
	rulesFile *os.File,
	restoreLegacy bool,
) (string, []string) {
	cmdName := iptablesRestore
	if cfg.IPv6 {
		cmdName = ip6tablesRestore
	}

	parameters := NewParameters().
		AppendIf(restoreLegacy, Wait(cfg.Wait), WaitInterval(cfg.WaitInterval)).
		Append(NoFlush())

	return cmdName, parameters.Build(cfg.Verbose, rulesFile.Name())
}
