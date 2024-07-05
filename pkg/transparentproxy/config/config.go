package config

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type Owner struct {
	UID string
}

// ValueOrRangeList is a format acceptable by iptables in which
// single values are denoted by just a number e.g. 1000
// multiple values (lists) are denoted by a number separated by a comma e.g. 1000,1001
// ranges are denoted by a colon e.g. 1000:1003 meaning 1000,1001,1002,1003
// ranges and multiple values can be mixed e.g. 1000,1005:1006 meaning 1000,1005,1006
type ValueOrRangeList string

type UIDsToPorts struct {
	Protocol string
	UIDs     ValueOrRangeList
	Ports    ValueOrRangeList
}

// TrafficFlow is a struct for Inbound/Outbound configuration
type TrafficFlow struct {
	Enabled             bool
	Port                uint16
	PortIPv6            uint16
	Chain               Chain
	RedirectChain       Chain
	ExcludePorts        []uint16
	ExcludePortsForUIDs []UIDsToPorts
	IncludePorts        []uint16
}

type DNS struct {
	Enabled    bool
	CaptureAll bool
	Port       uint16
	// The iptables chain where the upstream DNS requests should be directed to.
	// It is only applied for IP V4. Use with care. (default "RETURN")
	UpstreamTargetChain string
	ConntrackZoneSplit  bool
	ResolvConfigPath    string
}

type InitializedDNS struct {
	DNS
	ServersIPv4            []string
	ServersIPv6            []string
	ConntrackZoneSplitIPv4 bool
	ConntrackZoneSplitIPv6 bool
}

// Initialize initializes the ServersIPv4 and ServersIPv6 fields by parsing
// the nameservers from the file specified in the ResolvConfigPath field of
// the input DNS struct.
func (c DNS) Initialize(
	l Logger,
	cfg Config,
	executables InitializedExecutablesIPvX,
) (InitializedDNS, error) {
	initialized := InitializedDNS{DNS: c}

	// We don't have to continue initialization if the DNS traffic shouldn't be
	// redirected
	if !c.Enabled {
		return initialized, nil
	}

	if c.ConntrackZoneSplit {
		warning := func(ipvx string) string {
			return fmt.Sprintf(
				"conntrack zone splitting for %s is disabled. "+
					"Functionality requires the 'conntrack' iptables module",
				ipvx,
			)
		}

		initialized.ConntrackZoneSplitIPv4 = executables.IPv4.Functionality.
			ConntrackZoneSplit()
		if !initialized.ConntrackZoneSplitIPv4 {
			l.Warn(warning("IPv4"))
		}

		initialized.ConntrackZoneSplitIPv6 = executables.IPv6.Functionality.
			ConntrackZoneSplit()
		if !initialized.ConntrackZoneSplitIPv4 {
			l.Warn(warning("IPv6"))
		}
	}

	// We don't have to get DNS servers if we want to capture all DNS traffic
	if c.CaptureAll {
		return initialized, nil
	}

	dnsConfig, err := dns.ClientConfigFromFile(c.ResolvConfigPath)
	if err != nil {
		return initialized, errors.Wrapf(
			err,
			"unable to read file %s",
			c.ResolvConfigPath,
		)
	}

	for _, address := range dnsConfig.Servers {
		parsed := net.ParseIP(address)
		if parsed.To4() != nil {
			initialized.ServersIPv4 = append(initialized.ServersIPv4, address)
		} else {
			initialized.ServersIPv6 = append(initialized.ServersIPv6, address)
		}
	}

	return initialized, nil
}

type VNet struct {
	Networks []string
}

type Redirect struct {
	// NamePrefix is a prefix which will be used go generate chains name
	NamePrefix string
	Inbound    TrafficFlow
	Outbound   TrafficFlow
	DNS        DNS
	VNet       VNet
}

type InitializedRedirect struct {
	Redirect
	DNS InitializedDNS
}

func (c Redirect) Initialize(
	l Logger,
	cfg Config,
	executables InitializedExecutablesIPvX,
) (InitializedRedirect, error) {
	var err error

	initialized := InitializedRedirect{Redirect: c}

	// .DNS
	initialized.DNS, err = c.DNS.Initialize(l, cfg, executables)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize .DNS")
	}

	return initialized, nil
}

type Chain struct {
	Name string
}

func (c Chain) GetFullName(prefix string) string {
	return prefix + c.Name
}

type Ebpf struct {
	Enabled    bool
	InstanceIP string
	BPFFSPath  string
	CgroupPath string
	// The name of network interface which TC ebpf programs should bind to,
	// when not provided, we'll try to automatically determine it
	TCAttachIface      string
	ProgramsSourcePath string
}

type LogConfig struct {
	// Enabled determines whether iptables rules logging is activated. When
	// true, each packet matching an iptables rule will have its details logged,
	// aiding in diagnostics and monitoring of packet flows.
	Enabled bool
	// Level specifies the log level for iptables logging as defined by
	// netfilter. This level controls the verbosity and detail of the log
	// entries for matching packets. Higher values increase the verbosity.
	// Commonly used levels are: 1 (alerts), 4 (warnings), 5 (notices),
	// 7 (debugging). The exact behavior can depend on the system's syslog
	// configuration.
	Level uint16
}

type RetryConfig struct {
	// MaxRetries specifies the number of retries after the initial attempt.
	// A value of 0 means no retries, and only the initial attempt will be made.
	MaxRetries int
	// SleepBetweenRetries defines the duration to wait between retry attempts.
	// This delay helps in situations where immediate retries may not be
	// beneficial, allowing time for transient issues to resolve.
	SleepBetweenReties time.Duration
}

type Config struct {
	Owner    Owner
	Redirect Redirect
	Ebpf     Ebpf
	// DropInvalidPackets when enabled, kuma-dp will configure iptables to drop
	// packets that are considered invalid. This is useful in scenarios where
	// out-of-order packets bypass DNAT by iptables and reach the application
	// directly, causing connection resets.
	//
	// This behavior is typically observed during high-throughput requests (like
	// uploading large files). Enabling this option can improve application
	// stability by preventing these invalid packets from reaching the
	// application.
	//
	// However, enabling `DropInvalidPackets` might introduce slight performance
	// overhead. Consider the trade-off between connection stability and
	// performance before enabling this option.
	//
	// See also: https://kubernetes.io/blog/2019/03/29/kube-proxy-subtleties-debugging-an-intermittent-connection-reset/
	DropInvalidPackets bool
	// IPv6 when set will be used to configure iptables as well as ip6tables
	IPv6 bool
	// RuntimeStdout is the place where Any debugging, runtime information
	// will be placed (os.Stdout by default)
	RuntimeStdout io.Writer
	// RuntimeStderr is the place where error, runtime information will be
	// placed (os.Stderr by default)
	RuntimeStderr io.Writer
	// Verbose when set will generate iptables configuration with longer
	// argument/flag names, additional comments etc.
	Verbose bool
	// DryRun when set will not execute, but just display instructions which
	// otherwise would have served to install transparent proxy
	DryRun bool
	// Log configures logging for iptables rules using the LOG chain. When
	// enabled, this setting causes the kernel to log details about packets that
	// match the iptables rules, including IP/IPv6 headers. The logs are useful
	// for debugging and can be accessed via tools like dmesg or syslog. The
	// logging behavior is defined by the nested LogConfig struct.
	Log LogConfig
	// Wait is the amount of time, in seconds, that the application should wait
	// for the xtables exclusive lock before exiting. If the lock is not
	// available within the specified time, the application will exit with
	// an error. Default value *(0) means wait forever. To disable this behavior
	// and exit immediately if the xtables lock is not available, set this to
	// nil
	Wait uint
	// WaitInterval is the amount of time, in microseconds, that iptables should
	// wait between each iteration of the lock acquisition loop. This can be
	// useful if the xtables lock is being held by another application for
	// a long time, and you want to reduce the amount of CPU that iptables uses
	// while waiting for the lock
	WaitInterval uint
	// Retry allows you to configure the number of times that the system should
	// retry an installation if it fails
	Retry RetryConfig
	// StoreFirewalld when set, configures firewalld to store the generated
	// iptables rules.
	StoreFirewalld bool
	// Executables field holds configuration for the executables used to
	// interact with iptables (or ip6tables). It can handle both nft (nftables)
	// and legacy iptables modes, and supports IPv4 and IPv6 versions
	Executables ExecutablesNftLegacy
}

// InitializedConfig extends the Config struct by adding fields that require
// additional logic to retrieve their values. These values typically involve
// interacting with the system or external resources.
type InitializedConfig struct {
	Config
	// Redirect is an InitializedRedirect struct containing the initialized
	// redirection configuration. If DNS redirection is enabled this includes
	// the DNS servers retrieved from the specified resolv.conf file
	// (/etc/resolv.conf by default)
	Redirect InitializedRedirect
	// Executables field holds the initialized version of Config.Executables.
	// It attempts to locate the actual executable paths on the system based on
	// the provided configuration and verifies their functionality.
	Executables InitializedExecutablesIPvX
	// LoopbackInterfaceName represents the name of the loopback interface which
	// will be used to construct outbound iptable rules for outbound (i.e.
	// -A KUMA_MESH_OUTBOUND -s 127.0.0.6/32 -o lo -j RETURN)
	LoopbackInterfaceName string
	// Logger is utilized for recording logs across the entire lifecycle of the
	// InitializedConfig, from the initialization and configuration phases to
	// ongoing operations involving iptables, such as rule setup, modification,
	// and restoration. It ensures that logging capabilities are available not
	// only during the setup of system resources and configurations but also
	// throughout the execution of iptables-related activities.
	Logger Logger
}

// ShouldDropInvalidPackets determines whether the configuration indicates
// dropping invalid packets based on the configured behavior and the presence of
// the mangle table for the specified IP version.
//
// Args:
//
//	ipv6 (bool): Flag indicating if the check is for IPv6 or IPv4 packets.
//
// Returns:
//
//	bool: True if the configuration indicates dropping invalid packets for the
//	      specified IP version, and the corresponding mangle table is present.
//	      False otherwise.
//
// This method considers the following factors:
//   - `DropInvalidPackets` configuration setting: This setting should be enabled
//     for dropping invalid packets.
//   - Presence of Mangle Table: The mangle table is required for implementing
//     packet filtering rules. The method checks for the appropriate mangle table
//     based on the provided `ipv6` flag.
func (c InitializedConfig) ShouldDropInvalidPackets(ipv6 bool) bool {
	mangleTablePresent := c.Executables.IPv4.Functionality.Tables.Mangle
	if ipv6 {
		mangleTablePresent = c.Executables.IPv6.Functionality.Tables.Mangle
	}

	return c.DropInvalidPackets && mangleTablePresent
}

// ShouldRedirectDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AddRuleIf(ShouldRedirectDNS, Match(...), Jump(Drop()))
func (c InitializedConfig) ShouldRedirectDNS() bool {
	return c.Redirect.DNS.Enabled
}

// ShouldCaptureAllDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AddRuleIf(ShouldCaptureAllDNS, Match(...), Jump(Drop()))
func (c InitializedConfig) ShouldCaptureAllDNS() bool {
	return c.Redirect.DNS.CaptureAll
}

func (c Config) Initialize(ctx context.Context) (InitializedConfig, error) {
	var err error

	l := Logger{
		stdout: c.RuntimeStdout,
		stderr: c.RuntimeStderr,
		maxTry: c.Retry.MaxRetries + 1,
	}

	initialized := InitializedConfig{Config: c, Logger: l}

	initialized.Executables, err = c.Executables.Initialize(ctx, l, c)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize Executables configuration")
	}

	initialized.Redirect, err = c.Redirect.Initialize(l, c, initialized.Executables)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize Redirect configuration")
	}

	initialized.LoopbackInterfaceName, err = getLoopbackInterfaceName()
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize LoopbackInterfaceName")
	}

	return initialized, nil
}

func DefaultConfig() Config {
	return Config{
		Owner: Owner{UID: "5678"},
		Redirect: Redirect{
			NamePrefix: "KUMA_",
			Inbound: TrafficFlow{
				Enabled:       true,
				Port:          DefaultRedirectInbountPort,
				PortIPv6:      DefaultRedirectInbountPortIPv6,
				Chain:         Chain{Name: "MESH_INBOUND"},
				RedirectChain: Chain{Name: "MESH_INBOUND_REDIRECT"},
				ExcludePorts:  []uint16{},
				IncludePorts:  []uint16{},
			},
			Outbound: TrafficFlow{
				Enabled:       true,
				Port:          DefaultRedirectOutboundPort,
				Chain:         Chain{Name: "MESH_OUTBOUND"},
				RedirectChain: Chain{Name: "MESH_OUTBOUND_REDIRECT"},
				ExcludePorts:  []uint16{},
				IncludePorts:  []uint16{},
			},
			DNS: DNS{
				Port:               DefaultRedirectDNSPort,
				Enabled:            false,
				CaptureAll:         false,
				ConntrackZoneSplit: true,
				ResolvConfigPath:   "/etc/resolv.conf",
			},
			VNet: VNet{
				Networks: []string{},
			},
		},
		Ebpf: Ebpf{
			Enabled:            false,
			CgroupPath:         "/sys/fs/cgroup",
			BPFFSPath:          "/run/kuma/bpf",
			ProgramsSourcePath: "/tmp/kuma-ebpf",
		},
		DropInvalidPackets: false,
		IPv6:               false,
		RuntimeStdout:      os.Stdout,
		RuntimeStderr:      os.Stderr,
		Verbose:            false,
		DryRun:             false,
		Log: LogConfig{
			Enabled: false,
			Level:   LogLevelDebug,
		},
		Wait:         5,
		WaitInterval: 0,
		Retry: RetryConfig{
			// Specifies the number of retries after the initial attempt,
			// totaling 5 tries
			MaxRetries:         4,
			SleepBetweenReties: 2 * time.Second,
		},
		Executables: NewExecutablesNftLegacy(),
	}
}

// getLoopbackInterfaceName retrieves the name of the loopback interface on the
// system. This function iterates over all network interfaces and checks if the
// 'net.FlagLoopback' flag is set. If a loopback interface is found, its name is
// returned. Otherwise, an error message indicating that no loopback interface
// was found is returned.
func getLoopbackInterfaceName() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve network interfaces")
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			return iface.Name, nil
		}
	}

	return "", errors.New("no loopback interface found on the system")
}
