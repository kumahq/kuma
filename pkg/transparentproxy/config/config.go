package config

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/miekg/dns"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/pointer"
)

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

type InitializedDNS struct {
	DNS
	ServersIPv4 []string
	ServersIPv6 []string
}

// Initialize initializes the ServersIPv4 and ServersIPv6 fields by parsing
// the nameservers from the file specified in the ResolvConfigPath field of
// the input DNS struct.
func (c DNS) Initialize() (InitializedDNS, error) {
	initialized := InitializedDNS{DNS: c}

	// We don't have to get DNS servers if DNS traffic shouldn't be redirected,
	// or if we want to capture all DNS traffic
	if !c.Enabled || c.CaptureAll {
		return initialized, nil
	}

	dnsConfig, err := dns.ClientConfigFromFile(c.ResolvConfigPath)
	if err != nil {
		return initialized, errors.Errorf("unable to read file %s: %s", c.ResolvConfigPath, err)
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

func (c Redirect) Initialize() (InitializedRedirect, error) {
	var err error

	initialized := InitializedRedirect{Redirect: c}

	// .DNS
	initialized.DNS, err = c.DNS.Initialize()
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
	// StoreFirewalld when set, configures firewalld to store the generated
	// iptables rules.
	StoreFirewalld bool
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
	// LoopbackInterfaceName represents the name of the loopback interface which
	// will be used to construct outbound iptable rules for outbound (i.e.
	// -A KUMA_MESH_OUTBOUND -s 127.0.0.6/32 -o lo -j RETURN)
	LoopbackInterfaceName string
}

// ShouldDropInvalidPackets is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AddRuleIf(ShouldDropInvalidPackets, Match(...), Jump(Drop()))
func (c InitializedConfig) ShouldDropInvalidPackets() bool {
	return c.DropInvalidPackets
}

// ShouldRedirectDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AddRuleIf(ShouldRedirectDNS, Match(...), Jump(Drop()))
func (c InitializedConfig) ShouldRedirectDNS() bool {
	return c.Redirect.DNS.Enabled
}

// ShouldFallbackDNSToUpstreamChain is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AddRuleIf(ShouldFallbackDNSToUpstreamChain, Match(...), Jump(Drop()))
func (c InitializedConfig) ShouldFallbackDNSToUpstreamChain() bool {
	return c.Redirect.DNS.UpstreamTargetChain != ""
}

// ShouldCaptureAllDNS is just a convenience function which can be used in
// iptables conditional command generations instead of inlining anonymous functions
// i.e. AddRuleIf(ShouldCaptureAllDNS, Match(...), Jump(Drop()))
func (c InitializedConfig) ShouldCaptureAllDNS() bool {
	return c.Redirect.DNS.CaptureAll
}

// ShouldConntrackZoneSplit is a function which will check if DNS redirection and
// conntrack zone splitting settings are enabled (return false if not), and then
// will verify if there is conntrack iptables extension available to apply
// the DNS conntrack zone splitting iptables rules
func (c InitializedConfig) ShouldConntrackZoneSplit(iptablesExecutable string) bool {
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

func getLoopbackInterfaceName() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", errors.Wrap(err, "unable to list network interfaces")
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			return iface.Name, nil
		}
	}

	return "", errors.New("loopback interface not found")
}

func (c Config) Initialize() (InitializedConfig, error) {
	var err error

	initialized := InitializedConfig{Config: c}

	initialized.Redirect, err = c.Redirect.Initialize()
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
