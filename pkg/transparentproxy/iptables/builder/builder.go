package builder

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"net"
	"os"
=======
	"slices"
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
<<<<<<< HEAD
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/table"
)

const (
	iptables  = "iptables"
	ip6tables = "ip6tables"
)

type IPTables struct {
	raw    *table.RawTable
	nat    *table.NatTable
	mangle *table.MangleTable
}

func newIPTables(
	raw *table.RawTable,
	nat *table.NatTable,
	mangle *table.MangleTable,
) *IPTables {
	return &IPTables{
		raw:    raw,
		nat:    nat,
		mangle: mangle,
	}
}

func (t *IPTables) Build(verbose bool) string {
	var tables []string

	raw := t.raw.Build(verbose)
	if raw != "" {
		tables = append(tables, raw)
	}

	nat := t.nat.Build(verbose)
	if nat != "" {
		tables = append(tables, nat)
	}

	mangle := t.mangle.Build(verbose)
	if mangle != "" {
		tables = append(tables, mangle)
	}

	separator := "\n"
	if verbose {
		separator = "\n\n"
	}

	return strings.Join(tables, separator) + "\n"
}

func BuildIPTables(
	cfg config.Config,
	dnsServers []string,
=======
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
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	ipv6 bool,
) (string, error) {
<<<<<<< HEAD
	cfg = config.MergeConfigWithDefaults(cfg)

	loopbackIface, err := getLoopback()
	if err != nil {
		return "", fmt.Errorf("cannot obtain loopback interface: %s", err)
	}

	natTable, err := buildNatTable(cfg, dnsServers, loopbackIface.Name, ipv6)
=======
	// Attempt to build the NAT table rules.
	natTable, err := buildNatTable(cfg, ipv6)
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	if err != nil {
		return "", fmt.Errorf("failed to build NAT table: %w", err)
	}

<<<<<<< HEAD
	return newIPTables(
		buildRawTable(cfg, dnsServers, iptablesExecutablePath),
		natTable,
		buildMangleTable(cfg),
	).Build(cfg.Verbose), nil
}

// runtimeOutput is the file (should be os.Stdout by default) where we can dump generated
// rules for used to see and debug if something goes wrong, which can be overwritten
// in tests to not obfuscate the other, more relevant logs
func (r *restorer) saveIPTablesRestoreFile(logPrefix string, f *os.File, content string) error {
	fmt.Fprintf(r.cfg.RuntimeStdout, "%s writing following contents to rules file: %s\n", logPrefix, f.Name())
	fmt.Fprint(r.cfg.RuntimeStdout, content)

	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(content)
	if err != nil {
		return fmt.Errorf("unable to write iptables-restore file: %s", err)
	}

	return writer.Flush()
}

func createRulesFile(ipv6 bool) (*os.File, error) {
	prefix := iptables
	if ipv6 {
		prefix = ip6tables
	}

	filename := fmt.Sprintf("%s-rules-%d.txt", prefix, time.Now().UnixNano())

	f, err := os.CreateTemp("", filename)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s rules file: %s", iptables, err)
	}

	return f, nil
}

type restorer struct {
	cfg         config.Config
	ipv6        bool
	dnsServers  []string
	executables *Executables
}

func newIPTablesRestorer(
	ctx context.Context,
	cfg config.Config,
	ipv6 bool,
	dnsServers []string,
) (*restorer, error) {
	executables, err := DetectIptablesExecutables(ctx, cfg, ipv6)
	if err != nil {
		return nil, fmt.Errorf("unable to detect iptables restore binaries: %s", err)
	}

	return &restorer{
		cfg:         cfg,
		ipv6:        ipv6,
		dnsServers:  dnsServers,
		executables: executables,
	}, nil
}

func (r *restorer) restore(ctx context.Context) (string, error) {
	rulesFile, err := createRulesFile(r.ipv6)
	if err != nil {
		return "", err
	}
	defer rulesFile.Close()
	defer os.Remove(rulesFile.Name())

	if err := r.configureIPv6Address(); err != nil {
		return "", err
	}

	for i := 0; i <= r.cfg.Retry.MaxRetries; i++ {
		logPrefix := fmt.Sprintf("# [%d/%d]", i+1, r.cfg.Retry.MaxRetries+1)
		fmt.Fprintf(r.cfg.RuntimeStderr, "\n# [%d/%d] ", i+1, r.cfg.Retry.MaxRetries+1)

		output, err := r.tryRestoreIPTables(ctx, logPrefix, r.executables, rulesFile)
		if err == nil {
			return output, nil
		}

		if r.executables.fallback != nil {
			fmt.Fprintf(r.cfg.RuntimeStdout, "%s trying fallback\n", logPrefix)

			output, err := r.tryRestoreIPTables(ctx, logPrefix, r.executables.fallback, rulesFile)
			if err == nil {
				return output, nil
			}
		}

		if i < r.cfg.Retry.MaxRetries {
			fmt.Fprintf(
				r.cfg.RuntimeStdout,
				"%s will try again in %s\n",
				logPrefix,
				r.cfg.Retry.SleepBetweenReties,
			)

			time.Sleep(r.cfg.Retry.SleepBetweenReties)
		}
	}

	return "", errors.Errorf("%s failed", r.executables.Restore.Path)
}

func (r *restorer) tryRestoreIPTables(
	ctx context.Context,
	logPrefix string,
	executables *Executables,
	rulesFile *os.File,
) (string, error) {
	if executables.foundDockerOutputChain {
		r.cfg.Redirect.DNS.UpstreamTargetChain = "DOCKER_OUTPUT"
	}

	rules, err := BuildIPTables(r.cfg, r.dnsServers, r.ipv6, executables.Iptables.Path)
	if err != nil {
		return "", fmt.Errorf("unable to build iptable rules: %s", err)
	}

	if err := r.saveIPTablesRestoreFile(logPrefix, rulesFile, rules); err != nil {
		return "", fmt.Errorf("unable to save iptables restore file: %s", err)
	}

	params := buildRestoreParameters(r.cfg, rulesFile, executables.legacy())

	fmt.Fprintf(
		r.cfg.RuntimeStdout,
		"%s %s %s\n",
		logPrefix,
		executables.Restore.Path,
		strings.Join(params, " "),
	)

	output, err := executables.Restore.exec(ctx, params...)
	if err == nil {
		return output.String(), nil
	}

	fmt.Fprintf(
		r.cfg.RuntimeStderr,
		"%s failed with error: '%s'\n",
		logPrefix,
		strings.ReplaceAll(err.Error(), "\n", ""),
	)

	return "", err
}

func RestoreIPTables(ctx context.Context, cfg config.Config) (string, error) {
	cfg = config.MergeConfigWithDefaults(cfg)

	_, _ = cfg.RuntimeStdout.Write([]byte("# kumactl is about to apply the " +
		"iptables rules that will enable transparent proxying on the machine. " +
		"The SSH connection may drop. If that happens, just reconnect again.\n"))

	var err error
	var dnsIpv6, dnsIpv4 []string

	if cfg.ShouldRedirectDNS() && !cfg.ShouldCaptureAllDNS() {
		dnsIpv4, dnsIpv6, err = GetDnsServers(cfg.Redirect.DNS.ResolvConfigPath)
		if err != nil {
			return "", err
		}
	}

	ipv4Restorer, err := newIPTablesRestorer(ctx, cfg, false, dnsIpv4)
=======
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
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	if err != nil {
		return "", errors.Wrap(err, "unable to build iptables rules")
	}

	output, err := cfg.Executables.IPv4.Restore(ctx, rules)
	if err != nil {
		return "", errors.Wrap(err, "unable to restore iptables rules")
	}

	if cfg.IPv6 {
<<<<<<< HEAD
		ipv6Restorer, err := newIPTablesRestorer(ctx, cfg, true, dnsIpv6)
=======
		rules, err := BuildIPTablesForRestore(cfg, true)
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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
