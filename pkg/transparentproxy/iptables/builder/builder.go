package builder

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/table"
)

const (
	iptables         = "iptables"
	ip6tables        = "ip6tables"
	iptablesRestore  = "iptables-restore"
	ip6tablesRestore = "ip6tables-restore"
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

func BuildIPTables(cfg config.Config, dnsServers []string, ipv6 bool) (string, error) {
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
		buildRawTable(cfg, dnsServers),
		natTable,
		buildMangleTable(cfg),
	).Build(cfg.Verbose), nil
}

// runtimeOutput is the file (should be os.Stdout by default) where we can dump generated
// rules for used to see and debug if something goes wrong, which can be overwritten
// in tests to not obfuscate the other, more relevant logs
func saveIPTablesRestoreFile(runtimeOutput io.Writer, f *os.File, content string) error {
	_, _ = fmt.Fprintln(runtimeOutput, "# writing following contents to rules file: ", f.Name())
	_, _ = fmt.Fprintln(runtimeOutput, content)

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

func runRestoreCmd(cmdName string, params []string) (string, error) {
	// #nosec G204
	cmd := exec.Command(cmdName, params...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executing command failed: %s (with output: %q)", err, output)
	}

	return string(output), nil
}

func restoreIPTables(cfg config.Config, dnsServers []string, ipv6 bool) (string, error) {
	rulesFile, err := createRulesFile(ipv6)
	if err != nil {
		return "", err
	}
	defer rulesFile.Close()
	defer os.Remove(rulesFile.Name())

	err = configureIPv6Address(ipv6)
	if err != nil {
		return "", err
	}

	rules, err := BuildIPTables(cfg, dnsServers, ipv6)
	if err != nil {
		return "", fmt.Errorf("unable to build iptable rules: %s", err)
	}

	if err := saveIPTablesRestoreFile(cfg.RuntimeStdout, rulesFile, rules); err != nil {
		return "", fmt.Errorf("unable to save iptables restore file: %s", err)
	}

	return restoreIPTablesWithRetry(cfg, rulesFile, ipv6)
}

func restoreIPTablesWithRetry(cfg config.Config, rulesFile *os.File, ipv6 bool) (string, error) {
	restoreLegacy, err := checkForIptablesRestoreLegacy(ipv6)
	if err != nil {
		return "", errors.Wrap(err, "cannot check if version of iptables-restore is legacy")
	}

	cmdName, params := buildRestore(cfg, rulesFile, restoreLegacy, ipv6)

	for i := 0; i <= cfg.Retry.MaxRetries; i++ {
		output, err := runRestoreCmd(cmdName, params)
		if err == nil {
			return output, nil
		}

		_, _ = cfg.RuntimeStderr.Write([]byte(fmt.Sprintf(
			"# [%d/%d] %s returned error: '%s'",
			i+1,
			cfg.Retry.MaxRetries+1,
			strings.Join(append([]string{cmdName}, params...), " "),
			err.Error(),
		)))

		if i < cfg.Retry.MaxRetries {
			_, _ = cfg.RuntimeStderr.Write([]byte(fmt.Sprintf(
				" will try again in %s",
				cfg.Retry.SleepBetweenReties.String(),
			)))

			time.Sleep(cfg.Retry.SleepBetweenReties)
		}

		_, _ = cfg.RuntimeStderr.Write([]byte("\n"))
	}

	_, _ = cfg.RuntimeStderr.Write([]byte("\n"))

	return "", errors.Errorf("%s failed", cmdName)
}

// checkForIptablesRestoreLegacy checks if the version of ip{6}tables-restore is
// legacy (non-nftables). The --wait and --wait-interval flags are only valid
// with legacy ip{6}tables-restore. These flags are invalid with nftables
// because nftables back end transactions are atomic and there is no need for
// the global xtables lock, which has proven problematic in environments with
// large and/or rapidly changing rulesets.
func checkForIptablesRestoreLegacy(ipv6 bool) (bool, error) {
	cmdName := iptablesRestore
	if ipv6 {
		cmdName = ip6tablesRestore
	}

	output, err := exec.Command(cmdName, "--version").Output()
	if err != nil {
		return false, err
	}

	r := regexp.MustCompile(`ip6?tables-restore v.*? \((.*?)\)`)
	match := r.FindStringSubmatch(string(output))

	return len(match) == 2 && match[1] == "legacy", nil
}

// RestoreIPTables
// TODO (bartsmykla): add validation if ip{,6}tables are available
func RestoreIPTables(cfg config.Config) (string, error) {
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

	output, err := restoreIPTables(cfg, dnsIpv4, false)
	if err != nil {
		return "", fmt.Errorf("cannot restore ipv4 iptable rules: %s", err)
	}

	if cfg.IPv6 {
		ipv6Output, err := restoreIPTables(cfg, dnsIpv6, true)
		if err != nil {
			return "", fmt.Errorf("cannot restore ipv6 iptable rules: %s", err)
		}

		output += ipv6Output
	}

	_, _ = cfg.RuntimeStdout.Write([]byte("# iptables set to diverge the traffic " +
		"to Envoy.\n"))

	return output, nil
}

// configureIPv6Address sets up a new IP address on local interface. This is needed
// for IPv6 but not IPv4, as IPv4 defaults to `netmask 255.0.0.0`, which allows binding to addresses
// in the 127.x.y.z range, while IPv6 defaults to `prefixlen 128` which allows binding only to ::1.
// Equivalent to `ip -6 addr add "::6/128" dev lo`
func configureIPv6Address(ipv6 bool) error {
	if !ipv6 {
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
