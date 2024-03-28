package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/kumahq/kuma/pkg/util/pointer"
)

type TransparentProxyConfig struct {
	DryRun                    bool
	Verbose                   bool
	RedirectPortOutBound      string
	RedirectInBound           bool
	RedirectPortInBound       string
	RedirectPortInBoundV6     string
	IpFamilyMode              string
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
	MaxRetries         *int
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
}

// ShouldDropInvalidPackets is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldDropInvalidPackets, Match(...), Jump(Drop()))
func (c Config) ShouldDropInvalidPackets() bool {
	return c.DropInvalidPackets
}

// ShouldRedirectDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldRedirectDNS, Match(...), Jump(Drop()))
func (c Config) ShouldRedirectDNS() bool {
	return c.Redirect.DNS.Enabled
}

// ShouldFallbackDNSToUpstreamChain is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldFallbackDNSToUpstreamChain, Match(...), Jump(Drop()))
func (c Config) ShouldFallbackDNSToUpstreamChain() bool {
	return c.Redirect.DNS.UpstreamTargetChain != ""
}

// ShouldCaptureAllDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AppendIf(ShouldCaptureAllDNS, Match(...), Jump(Drop()))
func (c Config) ShouldCaptureAllDNS() bool {
	return c.Redirect.DNS.CaptureAll
}

// ShouldConntrackZoneSplit is a function which will check if DNS redirection and
// conntrack zone splitting settings are enabled (return false if not), and then
// will verify if there is conntrack iptables extension available to apply
// the DNS conntrack zone splitting iptables rules
func (c Config) ShouldConntrackZoneSplit() bool {
	if !c.Redirect.DNS.Enabled || !c.Redirect.DNS.ConntrackZoneSplit {
		return false
	}

	// There are situations where conntrack extension is not present (WSL2)
	// instead of failing the whole iptables application, we can log the warning,
	// skip conntrack related rules and move forward
	if err := exec.Command("iptables", "-m", "conntrack", "--help").Run(); err != nil {
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
			ProgramsSourcePath: "/tmp/kuma-ebpf",
		},
		DropInvalidPackets: false,
		IPv6:               false,
		RuntimeStdout:      os.Stdout,
		RuntimeStderr:      os.Stderr,
		Verbose:            true,
		DryRun:             false,
		Log: LogConfig{
			Enabled: false,
			Level:   DebugLogLevel,
		},
		Wait:         5,
		WaitInterval: 0,
		Retry: RetryConfig{
			MaxRetries:         pointer.To(4),
			SleepBetweenReties: 2 * time.Second,
		},
	}
}

func DefaultConfig() Config {
	return defaultConfig()
}

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
}
