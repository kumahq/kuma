package iptables

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

func Setup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	if cfg.DryRun {
		return dryRun(cfg), nil
	}

	return builder.RestoreIPTables(ctx, cfg)
}

func Cleanup(_ config.InitializedConfig) (string, error) {
	return "", errors.New("cleanup is not supported")
}

// dryRun simulates the setup of iptables rules for both IPv4 and IPv6
// configurations based on the provided config.InitializedConfig. It does not
// apply any changes to the system but generates and returns a string
// representation of what would be executed.
//
// The function operates as follows:
//  1. It defines a helper function, ipvxRun, which:
//     - Builds the iptables-restore content for either IPv4 or IPv6 depending
//     on the input.
//     - Prepends a header (### IPv4 ### or ### IPv6 ###) to distinguish between
//     the IP versions.
//     - Returns the formatted iptables rules or an error if the building
//     process fails.
//  2. Executes ipvxRun for IPv4 and, if enabled in the configuration, for IPv6.
//  3. Concatenates the results from IPv4 and IPv6 runs, separating them with
//     newlines for clarity.
//  4. Logs the final combined output using the configured logger without
//     prefixing, to ensure that the output is clear and unmodified, suitable
//     for review or documentation purposes.
//
// Args:
//
//   - cfg (config.InitializedConfig): Configuration settings that include flags
//     for dry run, logging, and IP version preferences.
//
// Returns:
//
//   - string: A combined string of formatted iptables commands for both IPv4
//     and IPv6.
//   - error: An error if there is a failure in generating the iptables commands
//     for any version.
func dryRun(cfg config.InitializedConfig) string {
	output := strings.Join(
		slices.Concat(
			dryRunIPvX(cfg.IPv4, false),
			dryRunIPvX(cfg.IPv6, true),
		),
		"\n\n",
	)

	cfg.Logger.InfoWithoutPrefix(output)

	return output
}

// dryRunIPvX generates iptables rules for either IPv4 or IPv6 based on the
// provided configuration. It returns a slice with a header indicating the
// IP version and the generated rules as a single string.
//
// Args:
//   - cfg (config.InitializedConfigIPvX): Configuration settings for IPv4 or
//     IPv6.
//   - ipv6 (bool): Indicates if the configuration is for IPv6.
//
// Returns:
//   - []string: A slice containing the header and the iptables rules for the
//     specified IP version.
func dryRunIPvX(cfg config.InitializedConfigIPvX, ipv6 bool) []string {
	if !cfg.Enabled() {
		return nil
	}

	return []string{
		fmt.Sprintf("### %s ###", consts.IPTypeMap[ipv6]),
		strings.TrimSpace(builder.BuildIPTablesForRestore(cfg)),
	}
}
