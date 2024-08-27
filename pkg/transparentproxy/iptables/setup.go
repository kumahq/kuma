package iptables

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

func Setup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	if cfg.DryRun {
		return dryRun(cfg), nil
	}

	cfg.Logger.Info("cleaning up any existing transparent proxy iptables rules")

	if err := Cleanup(ctx, cfg); err != nil {
		return "", errors.Wrap(err, "cleanup failed during setup")
	}

	return builder.RestoreIPTables(ctx, cfg)
}

// Cleanup removes iptables rules and chains related to the transparent proxy
// for both IPv4 and IPv6 configurations. It calls the internal cleanupIPvX
// function for each IP version, ensuring that only the relevant rules andran
// chains are removed based on the presence of iptables comments. If either
// cleanup process fails, an error is returned.
func Cleanup(ctx context.Context, cfg config.InitializedConfig) error {
	if err := cleanupIPvX(ctx, cfg.IPv4); err != nil {
		return errors.Wrap(err, "failed to cleanup IPv4 rules")
	}

	if err := cleanupIPvX(ctx, cfg.IPv6); err != nil {
		return errors.Wrap(err, "failed to cleanup IPv6 rules")
	}

	return nil
}

// cleanupIPvX removes iptables rules and chains related to the transparent
// proxy, ensuring that only the relevant rules and chains are removed based on
// the presence of iptables comments and chain name prefixes. It verifies the
// new rules after cleanup and restores them if they are valid.
func cleanupIPvX(ctx context.Context, cfg config.InitializedConfigIPvX) error {
	if !cfg.Enabled() {
		return nil
	}

	cfg.Logger.Infof(
		"starting cleanup of existing transparent proxy rules. Any rule found in chains with names starting with %q or containing comments starting with %q will be deleted",
		cfg.Redirect.NamePrefix,
		cfg.Comments.Prefix,
	)

	// Execute iptables-save to retrieve current rules.
	stdout, _, err := cfg.Executables.IptablesSave.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to execute iptables-save command")
	}

	output := stdout.String()
	containsTProxyRules := strings.Contains(output, cfg.Redirect.NamePrefix)
	containsTProxyComments := strings.Contains(output, cfg.Comments.Prefix)

	switch {
	case !containsTProxyRules && !containsTProxyComments:
		// If there are no transparent proxy rules or chains, there is
		// nothing to do.
		cfg.Logger.Info("no transparent proxy rules detected. No cleanup necessary")
		return nil
	case containsTProxyRules && !containsTProxyComments:
		return errors.New("transparent proxy rules detected, but expected comments are missing. Cleanup cannot proceed safely without comments to identify rules. Please remove the transparent proxy iptables rules manually")
	}

	// Split the output into lines and remove lines related to transparent
	// proxy rules and chains.
	lines := strings.Split(output, "\n")
	linesCleaned := slices.DeleteFunc(
		lines,
		func(line string) bool {
			return strings.HasPrefix(line, "#") ||
				strings.Contains(line, cfg.Comments.Prefix) ||
				strings.Contains(line, cfg.Redirect.NamePrefix)
		},
	)
	newRules := strings.Join(linesCleaned, "\n")

	// Verify if the new rules after cleanup are correct.
	if _, err := cfg.Executables.RestoreTest(ctx, newRules); err != nil {
		return errors.Wrap(
			err,
			"verification of new rules after cleanup failed",
		)
	}

	if cfg.DryRun {
		cfg.Logger.Info("[dry-run]: rules after cleanup:")
		cfg.Logger.InfoWithoutPrefix(strings.TrimSpace(newRules))
		return nil
	}

	// Restore the new rules with flushing.
	if _, err := cfg.Executables.RestoreWithFlush(ctx, newRules, true); err != nil {
		return errors.Wrap(
			err,
			"failed to restore rules with flush after cleanup",
		)
	}

	cfg.Logger.Info("cleanup of existing transparent proxy rules completed successfully")

	return nil
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

func dryRunIPvX(cfg config.InitializedConfigIPvX, ipv6 bool) []string {
	if !cfg.Enabled() {
		return nil
	}

	return []string{
		fmt.Sprintf("### %s ###", consts.IPTypeMap[ipv6]),
		strings.TrimSpace(builder.BuildIPTablesForRestore(cfg)),
	}
}
