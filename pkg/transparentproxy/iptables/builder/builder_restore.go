package builder

import (
	"os"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

func buildRestore(cfg config.Config, rulesFile *os.File) (string, []string) {
	cmdName := iptablesRestore
	if cfg.IPv6 {
		cmdName = ip6tablesRestore
	}

	return cmdName, NewParameters(
		NoFlush(),
		Wait(cfg.Wait),
		WaitInterval(cfg.WaitInterval),
	).Build(cfg.Verbose, rulesFile.Name())
}
