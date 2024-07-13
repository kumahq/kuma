package builder

import (
	"context"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

// BuildIPTablesForRestore constructs the complete set of iptables rules for
// restoring based on the provided configuration.
//
// This function performs the following steps:
//  1. Builds the rules for the raw, nat, and mangle tables using the
//     `BuildRulesForRestore` function, and filters out any empty rule sets.
//  2. Determines the separator based on the verbosity setting from the
//     configuration.
//  3. Joins the remaining rules with the determined separator and returns
//     the result.
//
// Args:
//
//   - cfg (config.InitializedConfigIPvX): The configuration used to initialize
//     the iptables rules.
//
// Returns:
//
//   - string: The complete set of iptables rules as a single string.
func BuildIPTablesForRestore(cfg config.InitializedConfigIPvX) string {
	// Build the rules for raw, NAT, and mangle tables, filtering out any empty
	// sets.
	result := slices.DeleteFunc([]string{
		tables.BuildRulesForRestore(cfg, buildRawTable(cfg)),
		tables.BuildRulesForRestore(cfg, buildNatTable(cfg)),
		tables.BuildRulesForRestore(cfg, buildMangleTable(cfg)),
	}, func(s string) bool { return s == "" })

	// Determine the separator based on verbosity setting.
	separator := "\n"
	if cfg.Verbose {
		separator = "\n\n"
	}

	// Join the rules with the determined separator and return.
	return strings.Join(result, separator) + "\n"
}

func RestoreIPTables(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	cfg.Logger.Info("kumactl is about to apply the iptables rules that " +
		"will enable transparent proxying on the machine. The SSH connection " +
		"may drop. If that happens, just reconnect again.")

	output, err := cfg.IPv4.Executables.Restore(
		ctx,
		BuildIPTablesForRestore(cfg.IPv4),
		false,
	)
	if err != nil {
		return "", errors.Wrap(err, "unable to restore iptables rules")
	}

	if cfg.IPv6.Enabled() {
		ipv6Output, err := cfg.IPv6.Executables.Restore(
			ctx,
			BuildIPTablesForRestore(cfg.IPv6),
			false,
		)
		if err != nil {
			return "", errors.Wrap(err, "unable to restore ip6tables rules")
		}

		output += ipv6Output
	}

	cfg.Logger.Info("iptables set to divert the traffic to Envoy")

	return output, nil
}
