package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

<<<<<<< HEAD
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type TransparentProxyConfig struct {
	DryRun                    bool
	Verbose                   bool
	RedirectPortOutBound      string
	RedirectInBound           bool
	RedirectPortInBound       string
	RedirectPortInBoundV6     string
	ExcludeInboundPorts       string
	ExcludeOutboundPorts      string
	ExcludedOutboundsForUIDs  []string
	UID                       string
	GID                       string
	RedirectDNS               bool
	RedirectAllDNSTraffic     bool
	AgentDNSListenerPort      string
	DNSUpstreamTargetChain    string
	SkipDNSConntrackZoneSplit bool
	ExperimentalEngine        bool
	EbpfEnabled               bool
	EbpfInstanceIP            string
	EbpfBPFFSPath             string
	EbpfCgroupPath            string
	EbpfTCAttachIface         string
	EbpfProgramsSourcePath    string
	VnetNetworks              []string
	Stdout                    io.Writer
	Stderr                    io.Writer
	RestoreLegacy             bool
	Wait                      uint
	WaitInterval              uint
	MaxRetries                *int
	SleepBetweenRetries       time.Duration
}

const DebugLogLevel uint16 = 7

=======
	"github.com/miekg/dns"
	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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

<<<<<<< HEAD
=======
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

>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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

<<<<<<< HEAD
=======
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

>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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
	Enabled bool
	Level   uint16
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
	// DropInvalidPackets when set will enable configuration which should drop
	// packets in invalid states
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
	// Log is the place where configuration for logging iptables rules will
	// be placed
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
<<<<<<< HEAD
}

// ShouldDropInvalidPackets is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldDropInvalidPackets, Match(...), Jump(Drop()))
func (c Config) ShouldDropInvalidPackets() bool {
	return c.DropInvalidPackets
=======
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
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
}

// ShouldRedirectDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldRedirectDNS, Match(...), Jump(Drop()))
func (c Config) ShouldRedirectDNS() bool {
	return c.Redirect.DNS.Enabled
}

<<<<<<< HEAD
// ShouldFallbackDNSToUpstreamChain is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldFallbackDNSToUpstreamChain, Match(...), Jump(Drop()))
func (c Config) ShouldFallbackDNSToUpstreamChain() bool {
	return c.Redirect.DNS.UpstreamTargetChain != ""
}

=======
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
// ShouldCaptureAllDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldCaptureAllDNS, Match(...), Jump(Drop()))
func (c Config) ShouldCaptureAllDNS() bool {
	return c.Redirect.DNS.CaptureAll
}

<<<<<<< HEAD
// ShouldConntrackZoneSplit is a function which will check if DNS redirection and
// conntrack zone splitting settings are enabled (return false if not), and then
// will verify if there is conntrack iptables extension available to apply
// the DNS conntrack zone splitting iptables rules
func (c Config) ShouldConntrackZoneSplit(iptablesExecutable string) bool {
	if !c.Redirect.DNS.Enabled || !c.Redirect.DNS.ConntrackZoneSplit {
		return false
	}

	if iptablesExecutable == "" {
		iptablesExecutable = "iptables"
	}

	// There are situations where conntrack extension is not present (WSL2)
	// instead of failing the whole iptables application, we can log the warning,
	// skip conntrack related rules and move forward
	if err := exec.Command(iptablesExecutable, "-m", "conntrack", "--help").Run(); err != nil {
		_, _ = fmt.Fprintf(c.RuntimeStderr,
			"# [WARNING] error occurred when validating if 'conntrack' iptables "+
				"module is present. Rules for DNS conntrack zone "+
				"splitting won't be applied: %s\n", err,
		)

		return false
	}

	return true
}

func defaultConfig() Config {
=======
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
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	return Config{
		Owner: Owner{UID: "5678"},
		Redirect: Redirect{
			NamePrefix: "",
			Inbound: TrafficFlow{
				Enabled:       true,
				Port:          15006,
				PortIPv6:      15010,
				Chain:         Chain{Name: "MESH_INBOUND"},
				RedirectChain: Chain{Name: "MESH_INBOUND_REDIRECT"},
				ExcludePorts:  []uint16{},
				IncludePorts:  []uint16{},
			},
			Outbound: TrafficFlow{
				Enabled:       true,
				Port:          15001,
				Chain:         Chain{Name: "MESH_OUTBOUND"},
				RedirectChain: Chain{Name: "MESH_OUTBOUND_REDIRECT"},
				ExcludePorts:  []uint16{},
				IncludePorts:  []uint16{},
			},
			DNS: DNS{
				Port:               15053,
				Enabled:            false,
				CaptureAll:         true,
				ConntrackZoneSplit: true,
				ResolvConfigPath:   "/etc/resolv.conf",
			},
			VNet: VNet{
				Networks: []string{},
			},
		},
		Ebpf: Ebpf{
			Enabled:            false,
			BPFFSPath:          "/run/kuma/bpf",
			ProgramsSourcePath: "/kuma/ebpf",
		},
		DropInvalidPackets: false,
		IPv6:               false,
		RuntimeStdout:      os.Stdout,
		RuntimeStderr:      os.Stderr,
		Verbose:            true,
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

<<<<<<< HEAD
func MergeConfigWithDefaults(cfg Config) Config {
	result := defaultConfig()

	// .Owner
	if cfg.Owner.UID != "" {
		result.Owner.UID = cfg.Owner.UID
	}

	// .Redirect
	if cfg.Redirect.NamePrefix != "" {
		result.Redirect.NamePrefix = cfg.Redirect.NamePrefix
	}

	// .Redirect.Inbound
	result.Redirect.Inbound.Enabled = cfg.Redirect.Inbound.Enabled
	if cfg.Redirect.Inbound.Port != 0 {
		result.Redirect.Inbound.Port = cfg.Redirect.Inbound.Port
	}

	if cfg.Redirect.Inbound.PortIPv6 != 0 {
		result.Redirect.Inbound.PortIPv6 = cfg.Redirect.Inbound.PortIPv6
	}

	if cfg.Redirect.Inbound.Chain.Name != "" {
		result.Redirect.Inbound.Chain.Name = cfg.Redirect.Inbound.Chain.Name
	}

	if cfg.Redirect.Inbound.RedirectChain.Name != "" {
		result.Redirect.Inbound.RedirectChain.Name = cfg.Redirect.Inbound.RedirectChain.Name
	}

	if len(cfg.Redirect.Inbound.ExcludePorts) > 0 {
		result.Redirect.Inbound.ExcludePorts = cfg.Redirect.Inbound.ExcludePorts
	}

	if len(cfg.Redirect.Inbound.IncludePorts) > 0 {
		result.Redirect.Inbound.IncludePorts = cfg.Redirect.Inbound.IncludePorts
	}

	// .Redirect.Outbound
	result.Redirect.Outbound.Enabled = cfg.Redirect.Outbound.Enabled
	if cfg.Redirect.Outbound.Port != 0 {
		result.Redirect.Outbound.Port = cfg.Redirect.Outbound.Port
	}

	if cfg.Redirect.Outbound.Chain.Name != "" {
		result.Redirect.Outbound.Chain.Name = cfg.Redirect.Outbound.Chain.Name
	}

	if cfg.Redirect.Outbound.RedirectChain.Name != "" {
		result.Redirect.Outbound.RedirectChain.Name = cfg.Redirect.Outbound.RedirectChain.Name
	}

	if len(cfg.Redirect.Outbound.ExcludePorts) > 0 {
		result.Redirect.Outbound.ExcludePorts = cfg.Redirect.Outbound.ExcludePorts
	}

	if len(cfg.Redirect.Outbound.IncludePorts) > 0 {
		result.Redirect.Outbound.IncludePorts = cfg.Redirect.Outbound.IncludePorts
	}

	if len(cfg.Redirect.Outbound.ExcludePortsForUIDs) > 0 {
		result.Redirect.Outbound.ExcludePortsForUIDs = cfg.Redirect.Outbound.ExcludePortsForUIDs
	}

	// .Redirect.DNS
	result.Redirect.DNS.Enabled = cfg.Redirect.DNS.Enabled
	result.Redirect.DNS.ConntrackZoneSplit = cfg.Redirect.DNS.ConntrackZoneSplit
	result.Redirect.DNS.CaptureAll = cfg.Redirect.DNS.CaptureAll
	if cfg.Redirect.DNS.ResolvConfigPath != "" {
		result.Redirect.DNS.ResolvConfigPath = cfg.Redirect.DNS.ResolvConfigPath
	}

	if cfg.Redirect.DNS.UpstreamTargetChain != "" {
		result.Redirect.DNS.UpstreamTargetChain = cfg.Redirect.DNS.UpstreamTargetChain
	}

	if cfg.Redirect.DNS.Port != 0 {
		result.Redirect.DNS.Port = cfg.Redirect.DNS.Port
	}

	// .Redirect.VNet
	if len(cfg.Redirect.VNet.Networks) > 0 {
		result.Redirect.VNet.Networks = cfg.Redirect.VNet.Networks
	}

	// .Ebpf
	result.Ebpf.Enabled = cfg.Ebpf.Enabled
	if cfg.Ebpf.InstanceIP != "" {
		result.Ebpf.InstanceIP = cfg.Ebpf.InstanceIP
	}

	if cfg.Ebpf.BPFFSPath != "" {
		result.Ebpf.BPFFSPath = cfg.Ebpf.BPFFSPath
	}

	if cfg.Ebpf.ProgramsSourcePath != "" {
		result.Ebpf.ProgramsSourcePath = cfg.Ebpf.ProgramsSourcePath
	}

	// .DropInvalidPackets
	result.DropInvalidPackets = cfg.DropInvalidPackets

	// .IPv6
	result.IPv6 = cfg.IPv6

	// .RuntimeStdout
	if cfg.RuntimeStdout != nil {
		result.RuntimeStdout = cfg.RuntimeStdout
	}

	// .RuntimeStderr
	if cfg.RuntimeStderr != nil {
		result.RuntimeStderr = cfg.RuntimeStderr
	}

	// .Verbose
	result.Verbose = cfg.Verbose

	// .DryRun
	result.DryRun = cfg.DryRun

	// .Log
	result.Log.Enabled = cfg.Log.Enabled
	if cfg.Log.Level != DebugLogLevel {
		result.Log.Level = cfg.Log.Level
	}

	// .Wait
	result.Wait = cfg.Wait

	// .WaitInterval
	result.WaitInterval = cfg.WaitInterval

	// .Retry
	if cfg.Retry.MaxRetries != nil {
		result.Retry.MaxRetries = cfg.Retry.MaxRetries
	}

	if cfg.Retry.SleepBetweenReties != 0 {
		result.Retry.SleepBetweenReties = cfg.Retry.SleepBetweenReties
	}

	return result
=======
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
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
}
