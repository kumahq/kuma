package iptables

import (
	"context"
	"errors"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
)

func Setup(ctx context.Context, cfg config.Config) (string, error) {
	if cfg.DryRun {
<<<<<<< HEAD
		// TODO (bartsmykla): we should generate IPv4 and IPv6 when cfg.IPv6 is
		//  set, but currently in DryRun mode we would just display IPv6
		//  configuration when cfg.IPv6 is set
		// TODO (bartsmykla): I think dns servers should be provided as a config
		//  value instead of explicit function parameter here
		iptablesExecutablePath := "iptables"
		if executables, err := builder.DetectIptablesExecutables(ctx, cfg, cfg.IPv6); err == nil && executables != nil {
			iptablesExecutablePath = executables.Iptables.Path
		}

		output, err := builder.BuildIPTables(cfg, nil, cfg.IPv6, iptablesExecutablePath)
		if err != nil {
			return "", err
		}

		_, _ = cfg.RuntimeStdout.Write([]byte(output))

		return output, nil
=======
		return dryRun(cfg)
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	}

	return builder.RestoreIPTables(ctx, cfg)
}

func Cleanup(cfg config.Config) (string, error) {
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
//  3. Concatenates the results from IPv4 and IPv6 runs, separating them with a
//     newlines for clarity.
//  4. Logs the final combined output using the configured logger without
//     prefixing, to ensure that the output is clear and unmodified, suitable
//     for review or documentation purposes.
//
// Args:
//
//	cfg (config.InitializedConfig): Configuration settings that include flags
//	 for dry run, logging, and IP version preferences.
//
// Returns:
//
//	string: A combined string of formatted iptables commands for both IPv4 and
//	 IPv6.
//	error: An error if there is a failure in generating the iptables commands
//	 for any version.
func dryRun(cfg config.InitializedConfig) (string, error) {
	ipvxRun := func(ipv6 bool) ([]string, error) {
		var result []string

		output, err := builder.BuildIPTablesForRestore(cfg, ipv6)
		if err != nil {
			return nil, err
		}

		if !ipv6 {
			result = append(result, "### IPv4 ###")
		} else {
			result = append(result, "### IPv6 ###")
		}

		result = append(result, strings.TrimSpace(output))

		return result, nil
	}

	output, err := ipvxRun(false)
	if err != nil {
		return "", err
	}

	if cfg.IPv6 {
		ipv6Output, err := ipvxRun(true)
		if err != nil {
			return "", err
		}

		output = append(output, ipv6Output...)
	}

	combinedOutput := strings.Join(output, "\n\n")

	cfg.Logger.InfoWithoutPrefix(combinedOutput)

	return combinedOutput, nil
}
