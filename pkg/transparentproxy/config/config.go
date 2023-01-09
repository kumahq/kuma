package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type TransparentProxyConfig struct {
	DryRun                         bool
	Verbose                        bool
	RedirectPortOutBound           string
	RedirectInBound                bool
	RedirectPortInBound            string
	RedirectPortInBoundV6          string
	ExcludeInboundPorts            string
	ExcludeOutboundPorts           string
	ExcludeOutboundTCPPortsForUIDs []string
	ExcludeOutboundUDPPortsForUIDs []string
	UID                            string
	GID                            string
	RedirectDNS                    bool
	RedirectAllDNSTraffic          bool
	AgentDNSListenerPort           string
	DNSUpstreamTargetChain         string
	SkipDNSConntrackZoneSplit      bool
	ExperimentalEngine             bool
	EbpfEnabled                    bool
	EbpfInstanceIP                 string
	EbpfBPFFSPath                  string
	EbpfCgroupPath                 string
	EbpfTCAttachIface              string
	EbpfProgramsSourcePath         string
	VnetNetworks                   []string
	Stdout                         io.Writer
	Stderr                         io.Writer
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
	Enabled            bool
	CaptureAll         bool
	Port               uint16
	ConntrackZoneSplit bool
	ResolvConfigPath   string
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
		_, _ = fmt.Fprintf(c.RuntimeStdout,
			"[WARNING] error occurred when validating if 'conntrack' iptables "+
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
			Level:   DebugLogLevel,
		},
	}
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

	return result
}
