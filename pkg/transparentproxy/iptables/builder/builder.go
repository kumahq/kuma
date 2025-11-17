package builder

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/kumahq/kuma/v2/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/v2/pkg/transparentproxy/iptables/table"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
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
	ipv6 bool,
	iptablesExecutablePath string,
) (string, error) {
	cfg = config.MergeConfigWithDefaults(cfg)

	loopbackIface, err := getLoopback()
	if err != nil {
		return "", fmt.Errorf("cannot obtain loopback interface: %s", err)
	}

	natTable, err := buildNatTable(cfg, dnsServers, loopbackIface.Name, ipv6)
	if err != nil {
		return "", fmt.Errorf("build nat table: %s", err)
	}

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

	maxRetries := pointer.Deref(r.cfg.Retry.MaxRetries)

	for i := 0; i <= maxRetries; i++ {
		logPrefix := fmt.Sprintf("# [%d/%d]", i+1, maxRetries+1)

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

		if i < maxRetries {
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
	if err != nil {
		return "", err
	}

	output, err := ipv4Restorer.restore(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot restore ipv4 iptable rules: %s", err)
	}

	if cfg.IPv6 {
		ipv6Restorer, err := newIPTablesRestorer(ctx, cfg, true, dnsIpv6)
		if err != nil {
			return "", err
		}

		ipv6Output, err := ipv6Restorer.restore(ctx)
		if err != nil {
			return "", fmt.Errorf("cannot restore ipv6 iptable rules: %s", err)
		}

		output += ipv6Output
	}

	fmt.Fprintln(cfg.RuntimeStdout, "# iptables set to divert the traffic to Envoy")

	return output, nil
}

// configureIPv6Address sets up a new IP address on local interface. This is needed
// for IPv6 but not IPv4, as IPv4 defaults to `netmask 255.0.0.0`, which allows binding to addresses
// in the 127.x.y.z range, while IPv6 defaults to `prefixlen 128` which allows binding only to ::1.
// Equivalent to `ip -6 addr add "::6/128" dev lo`
func (r *restorer) configureIPv6Address() error {
	if !r.ipv6 {
		return nil
	}
	link, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("failed to find 'lo' link: %v", err)
	}
	// Equivalent to `ip -6 addr add "::6/128" dev lo`
	address := &net.IPNet{IP: net.ParseIP("::6"), Mask: net.CIDRMask(128, 128)}
	addr := &netlink.Addr{IPNet: address}

	err = netlink.AddrAdd(link, addr)
	if ignoreExists(err) != nil {
		return fmt.Errorf("failed to add IPv6 inbound address: %v", err)
	}
	return nil
}

func ignoreExists(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "file exists") {
		return nil
	}
	return err
}
