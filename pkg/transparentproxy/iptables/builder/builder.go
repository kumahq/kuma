package builder

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

// BuildIPTablesForRestore constructs the complete set of iptables rules for
// restoring, based on the provided configuration and IP version.
//
// This function performs the following steps:
//  1. Builds the NAT table rules by calling `buildNatTable`. If this step
//     fails, it returns an error indicating the failure to build the NAT table.
//  2. Builds the rules for the raw, NAT, and mangle tables using the
//     `BuildRulesForRestore` function.
//  3. Filters out any empty rule sets from the results.
//  4. Joins the remaining rules with a separator based on the verbosity setting
//     from the configuration.
//  5. Returns the joined rules as a single string.
//
// Args:
//
//	cfg (config.InitializedConfig): The configuration used to initialize the
//	  iptables rules.
//	ipv6 (bool): A boolean indicating whether to build rules for IPv6.
//
// Returns:
//
//	string: The complete set of iptables rules as a single string.
//	error: An error if the NAT table rules cannot be built.
func BuildIPTablesForRestore(
	cfg config.InitializedConfig,
	ipv6 bool,
) (string, error) {
	// Attempt to build the NAT table rules.
	natTable, err := buildNatTable(cfg, ipv6)
	if err != nil {
		return "", fmt.Errorf("failed to build NAT table: %w", err)
	}

	// Build the rules for raw, NAT, and mangle tables, filtering out any empty
	// sets.
	result := slices.DeleteFunc([]string{
		tables.BuildRulesForRestore(buildRawTable(cfg, ipv6), cfg.Verbose),
		tables.BuildRulesForRestore(natTable, cfg.Verbose),
		tables.BuildRulesForRestore(buildMangleTable(cfg, ipv6), cfg.Verbose),
	}, func(s string) bool { return s == "" })

	// Determine the separator based on verbosity setting.
	separator := "\n"
	if cfg.Verbose {
		separator = "\n\n"
	}

	// Join the rules with the determined separator and return.
	return strings.Join(result, separator) + "\n", nil
}

func RestoreIPTables(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	cfg.Logger.Info("kumactl is about to apply the iptables rules that " +
		"will enable transparent proxying on the machine. The SSH connection " +
		"may drop. If that happens, just reconnect again.")

	rules, err := BuildIPTablesForRestore(cfg, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to build iptables rules")
	}

	output, err := cfg.Executables.IPv4.Restore(ctx, rules)
	if err != nil {
		return "", errors.Wrap(err, "unable to restore iptables rules")
	}

	if cfg.IPv6 {
		rules, err := BuildIPTablesForRestore(cfg, true)
		if err != nil {
			return "", errors.Wrap(err, "unable to build ip6tables rules")
		}

		ipv6Output, err := cfg.Executables.IPv6.Restore(ctx, rules)
		if err != nil {
			return "", errors.Wrap(err, "unable to restore ip6tables rules")
		}

		output += ipv6Output
	}

	cfg.Logger.Info("iptables set to divert the traffic to Envoy")

	return output, nil
}
