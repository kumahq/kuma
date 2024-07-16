package config

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

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

// NewValueOrRangeList creates a ValueOrRangeList from a given value or range of
// values. It accepts a parameter of type []uint16, uint16, or string and
// converts it to a ValueOrRangeList, which is a comma-separated string
// representation of the values.
//
// The function panics if an unsupported type is provided, although the type
// constraints should prevent this from occurring.
func NewValueOrRangeList[T ~[]uint16 | ~uint16 | ~string](v T) ValueOrRangeList {
	switch value := any(v).(type) {
	case []uint16:
		var ports []string
		for _, port := range value {
			ports = append(ports, strconv.Itoa(int(port)))
		}
		return ValueOrRangeList(strings.Join(ports, ","))
	case uint16:
		return ValueOrRangeList(strconv.Itoa(int(value)))
	case string:
		return ValueOrRangeList(value)
	default:
		// Shouldn't be possible to catch this
		panic(errors.Errorf("invalid value type: %T", value))
	}
}

type Exclusion struct {
	Protocol ProtocolL4
	Address  string
	UIDs     ValueOrRangeList
	Ports    ValueOrRangeList
}

// TrafficFlow is a struct for Inbound/Outbound configuration
type TrafficFlow struct {
	Enabled             bool
	Port                uint16
	ChainName           string
	RedirectChainName   string
	ExcludePorts        []uint16
	ExcludePortsForUIDs []string
	ExcludePortsForIPs  []string
	IncludePorts        []uint16
}

func (c TrafficFlow) Initialize(
	ipv6 bool,
	chainNamePrefix string,
) (InitializedTrafficFlow, error) {
	initialized := InitializedTrafficFlow{TrafficFlow: c, Port: c.Port}

	if c.ChainName == "" {
		return InitializedTrafficFlow{}, errors.New("no chain name provided")
	}
	initialized.ChainName = fmt.Sprintf("%s_%s", chainNamePrefix, c.ChainName)

	if c.RedirectChainName == "" {
		return InitializedTrafficFlow{}, errors.New("no redirect chain name provided")
	}
	initialized.RedirectChainName = fmt.Sprintf("%s_%s", chainNamePrefix, c.RedirectChainName)

	excludePortsForUIDs, err := parseExcludePortsForUIDs(c.ExcludePortsForUIDs)
	if err != nil {
		return initialized, errors.Wrap(
			err,
			"parsing excluded outbound ports for uids failed",
		)
	}

	excludePortsForIPs, err := parseExcludePortsForIPs(c.ExcludePortsForIPs, ipv6)
	if err != nil {
		return initialized, errors.Wrap(
			err,
			"parsing excluded outbound ports for IPs failed",
		)
	}

	initialized.Exclusions = slices.Concat(
		initialized.Exclusions,
		excludePortsForUIDs,
		excludePortsForIPs,
	)

	return initialized, nil
}

type InitializedTrafficFlow struct {
	TrafficFlow
	Exclusions        []Exclusion
	Port              uint16
	ChainName         string
	RedirectChainName string
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
	Servers            []string
	ConntrackZoneSplit bool
	Enabled            bool
}

// Initialize initializes the ServersIPv4 and ServersIPv6 fields by parsing
// the nameservers from the file specified in the ResolvConfigPath field of
// the input DNS struct.
func (c DNS) Initialize(
	l Logger,
	executables InitializedExecutablesIPvX,
	ipv6 bool,
) (InitializedDNS, error) {
	initialized := InitializedDNS{DNS: c, Enabled: c.Enabled}

	// We don't have to continue initialization if the DNS traffic shouldn't be
	// redirected
	if !c.Enabled {
		return initialized, nil
	}

	if c.ConntrackZoneSplit {
		initialized.ConntrackZoneSplit = executables.Functionality.ConntrackZoneSplit()
		if !initialized.ConntrackZoneSplit {
			l.Warn("conntrack zone splitting is disabled. Functionality requires the 'conntrack' iptables module")
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

	// Loop through each DNS server address parsed from the resolv.conf file.
	for _, address := range dnsConfig.Servers {
		parsed := net.ParseIP(address)
		// Check if the address matches the expected IP version.
		// - If config is not for IPv6 and the address is IPv4, add to the list.
		// - If config is for IPv6 and the address is IPv6, add to the list.
		if !ipv6 && parsed.To4() != nil || ipv6 && parsed.To4() == nil {
			initialized.Servers = append(initialized.Servers, address)
		}
	}

	if len(initialized.Servers) == 0 {
		initialized.Enabled = false
		initialized.ConntrackZoneSplit = false

		l.Warnf(
			"couldn't find any %s servers in %s file. Capturing %[1]s DNS traffic will be disabled",
			IPTypeMap[ipv6],
			c.ResolvConfigPath,
		)
	}

	return initialized, nil
}

type VNet struct {
	// Networks specifies virtual networks using the format interfaceName:CIDR.
	// The interface name can be exact or a prefix followed by a "+", allowing
	// matching for interfaces starting with the given prefix.
	// Examples:
	// - "docker0:172.17.0.0/16"
	// - "br+:172.18.0.0/16" (matches any interface starting with "br")
	// - "iface:::1/64"
	Networks []string
}

// Initialize processes the virtual networks specified in the VNet struct and
// separates them into IPv4 and IPv6 categories based on their CIDR notation.
// It returns an InitializedVNet struct that contains the parsed interface
// names and corresponding CIDRs for the specified IP version (IPv4 or IPv6).
//
// This method performs the following steps:
//  1. Iterates through each network definition in the Networks slice.
//  2. Splits each network definition into an interface name and a CIDR block
//     using the first colon (":") as the delimiter.
//  3. Validates the format of the network definition, returning an error if it
//     is invalid.
//  4. Parses the CIDR block to determine whether it is an IPv4 or IPv6 address.
//     - If the CIDR block is valid and matches the specified IP version,
//     it is added to the InterfaceCIDRs map.
//  5. Constructs and returns an InitializedVNet struct containing the
//     populated InterfaceCIDRs map.
func (c VNet) Initialize(ipv6 bool) (InitializedVNet, error) {
	initialized := InitializedVNet{InterfaceCIDRs: map[string]string{}}

	for _, network := range c.Networks {
		// We accept only the first ":" so in case of IPv6 there should be no
		// problem with parsing
		pair := strings.SplitN(network, ":", 2)
		if len(pair) < 2 {
			return InitializedVNet{}, errors.Errorf(
				"invalid virtual network definition: %s",
				network,
			)
		}

		address, _, err := net.ParseCIDR(pair[1])
		if err != nil {
			return InitializedVNet{}, errors.Wrapf(
				err,
				"invalid CIDR definition for %s",
				pair[1],
			)
		}

		// Add the address to the map if it matches the specified IP version
		if (!ipv6 && address.To4() != nil) || (ipv6 && address.To4() == nil) {
			initialized.InterfaceCIDRs[pair[0]] = pair[1]
		}
	}

	return initialized, nil
}

type InitializedVNet struct {
	// InterfaceCIDRs is a map where the keys are interface names and the values
	// are IP addresses in CIDR notation, representing the parsed and validated
	// virtual network configurations.
	InterfaceCIDRs map[string]string
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
	DNS      InitializedDNS
	VNet     InitializedVNet
	Inbound  InitializedTrafficFlow
	Outbound InitializedTrafficFlow
}

func (c Redirect) Initialize(
	l Logger,
	executables InitializedExecutablesIPvX,
	ipv6 bool,
) (InitializedRedirect, error) {
	var err error

	initialized := InitializedRedirect{Redirect: c}

	// .DNS
	initialized.DNS, err = c.DNS.Initialize(l, executables, ipv6)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize .DNS")
	}

	// .VNet
	initialized.VNet, err = c.VNet.Initialize(ipv6)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize .VNet")
	}

	// .Inbound
	initialized.Inbound, err = c.Inbound.Initialize(ipv6, c.NamePrefix)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize .Inbound")
	}

	// .Outbound
	initialized.Outbound, err = c.Outbound.Initialize(ipv6, c.NamePrefix)
	if err != nil {
		return initialized, errors.Wrap(err, "unable to initialize .Outbound")
	}

	return initialized, nil
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

// Comment struct contains the configuration for iptables rule comments.
// It includes an option to enable or disable comments.
type Comment struct {
	Disabled bool
}

// InitializedComment struct contains the processed configuration for iptables
// rule comments. It indicates whether comments are enabled and the prefix to
// use for comment text.
type InitializedComment struct {
	// Enabled indicates whether iptables rule comments are enabled based on
	// the initial configuration and system capabilities.
	Enabled bool
	// Prefix defines the prefix to be used for comments on iptables rules,
	// aiding in identifying and organizing rules created by the transparent
	// proxy.
	Prefix string
}

// Initialize processes the Comment configuration and determines whether
// iptables rule comments should be enabled. It checks the system's
// functionality to see if the comment module is available and returns
// an InitializedComment struct with the result.
func (c Comment) Initialize(e InitializedExecutablesIPvX) InitializedComment {
	return InitializedComment{
		Enabled: !c.Disabled && e.Functionality.Modules.Comment,
		Prefix:  IptablesRuleCommentPrefix,
	}
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
	// Comment configures the prefix and enable/disable status for iptables rule
	// comments. This setting helps in identifying and organizing iptables rules
	// created by the transparent proxy, making them easier to manage and debug.
	Comment Comment
}

// InitializedConfigIPvX extends the Config struct by adding fields that require
// additional logic to retrieve their values. These values typically involve
// interacting with the system or external resources.
type InitializedConfigIPvX struct {
	Config
	// Logger is utilized for detailed logging throughout the lifecycle of the
	// InitializedConfigIPvX. This includes specific logging for iptables
	// operations such as rule setup, modification, and restoration. The Logger
	// in this struct ensures detailed, step-by-step logs are available for
	// operations related to the corresponding IP version (IPv4 or IPv6), aiding
	// in diagnostics and debugging.
	Logger Logger
	// Redirect is an InitializedRedirect struct containing the initialized
	// redirection configuration. If DNS redirection is enabled this includes
	// the DNS servers retrieved from the specified resolv.conf file
	// (/etc/resolv.conf by default)
	Redirect InitializedRedirect
	// Executables field holds the initialized version of Config.Executables.
	// It attempts to locate the actual executable paths on the system based on
	// the provided configuration and verifies their functionality.
	Executables InitializedExecutablesIPvX
	// DropInvalidPackets when enabled, kuma-dp will configure iptables to drop
	// packets that are considered invalid. This is useful in scenarios where
	// out-of-order packets bypass DNAT by iptables and reach the application
	// directly, causing connection resets. This field is set during
	// configuration initialization and considers whether the mangle table is
	// available for the corresponding IP version (IPv4 or IPv6).
	DropInvalidPackets bool
	// LoopbackInterfaceName represents the name of the loopback interface which
	// will be used to construct outbound iptable rules for outbound (i.e.
	// -A KUMA_MESH_OUTBOUND -s 127.0.0.6/32 -o lo -j RETURN)
	LoopbackInterfaceName string
	// LocalhostCIDR is a string representing the CIDR notation of the localhost
	// address for the given IP version (IPv4 or IPv6). This is used to
	// construct rules related to the loopback interface.
	LocalhostCIDR string
	// InboundPassthroughCIDR is a string representing the CIDR notation of the
	// address used for inbound passthrough traffic. This is used to construct
	// rules allowing specific traffic to bypass normal proxying.
	InboundPassthroughCIDR string
	// Comment holds the processed configuration for iptables rule comments,
	// indicating whether comments are enabled and the prefix to use for comment
	// text. This helps in identifying and organizing iptables rules created by
	// the transparent proxy, making them easier to manage and debug.
	Comment InitializedComment

	enabled bool
}

// Enabled returns the state of the 'enabled' field, indicating whether the
// IP version-specific configuration is enabled.
//
// This method simply returns the value of the 'enabled' field which
// determines if the corresponding IPv4 or IPv6 configuration is active.
func (c InitializedConfigIPvX) Enabled() bool {
	return c.enabled
}

type InitializedConfig struct {
	// Logger is utilized for recording general logs during the lifecycle of the
	// InitializedConfig, including the initialization and finalization phases
	// of the transparent proxy installation process. This logger is used to log
	// high-level information and statuses, while more specific logging related
	// to iptables operations is handled by the Logger in InitializedConfigIPvX.
	Logger Logger
	// DryRun when set will not execute, but just display instructions which
	// otherwise would have served to install transparent proxy
	DryRun bool
	// IPv4 contains the initialized configuration specific to IPv4. This
	// includes all settings, executables, and rules relevant to IPv4 iptables
	// management.
	IPv4 InitializedConfigIPvX
	// IPv6 contains the initialized configuration specific to IPv6. This
	// includes all settings, executables, and rules relevant to IPv6 ip6tables
	// management.
	IPv6 InitializedConfigIPvX
}

func (c Config) Initialize(ctx context.Context) (InitializedConfig, error) {
	var err error

	l := Logger{
		stdout: c.RuntimeStdout,
		stderr: c.RuntimeStderr,
		maxTry: c.Retry.MaxRetries + 1,
	}

	loggerIPv4 := l.WithPrefix(IptablesCommandByFamily[false])
	loggerIPv6 := l.WithPrefix(IptablesCommandByFamily[true])

	initialized := InitializedConfig{
		Logger: l,
		IPv4: InitializedConfigIPvX{
			Config:                 c,
			Logger:                 loggerIPv4,
			LocalhostCIDR:          LocalhostCIDRIPv4,
			InboundPassthroughCIDR: InboundPassthroughSourceAddressCIDRIPv4,
			enabled:                true,
		},
		DryRun: c.DryRun,
	}

	e, err := c.Executables.Initialize(ctx, l, c)
	if err != nil {
		return initialized, errors.Wrap(
			err,
			"unable to initialize Executables configuration",
		)
	}
	initialized.IPv4.Executables = e.IPv4

	ipv4Redirect, err := c.Redirect.Initialize(loggerIPv4, e.IPv4, false)
	if err != nil {
		return initialized, errors.Wrap(
			err,
			"unable to initialize IPv4 Redirect configuration",
		)
	}
	initialized.IPv4.Redirect = ipv4Redirect

	loopbackInterfaceName, err := getLoopbackInterfaceName()
	if err != nil {
		return initialized, errors.Wrap(
			err,
			"unable to initialize LoopbackInterfaceName",
		)
	}
	initialized.IPv4.LoopbackInterfaceName = loopbackInterfaceName

	initialized.IPv4.Comment = c.Comment.Initialize(e.IPv4)
	initialized.IPv4.DropInvalidPackets = c.DropInvalidPackets && e.IPv4.Functionality.Tables.Mangle

	if !c.IPv6 {
		return initialized, nil
	}

	if ok, err := hasLocalIPv6(); !ok || err != nil {
		if c.Verbose {
			loggerIPv6.Warn("IPv6 executables initialization skipped:", err)
		}
		return initialized, nil
	}

	if err := configureIPv6OutboundAddress(); err != nil {
		if c.Verbose {
			loggerIPv6.Warn(
				"failed to configure IPv6 outbound address. IPv6 rules will be skipped:",
				err,
			)
		}
		return initialized, nil
	}

	initialized.IPv6 = InitializedConfigIPvX{
		Config:                 c,
		Logger:                 loggerIPv6,
		Executables:            e.IPv6,
		LoopbackInterfaceName:  loopbackInterfaceName,
		LocalhostCIDR:          LocalhostCIDRIPv6,
		InboundPassthroughCIDR: InboundPassthroughSourceAddressCIDRIPv6,
		Comment:                c.Comment.Initialize(e.IPv6),
		DropInvalidPackets:     c.DropInvalidPackets && e.IPv6.Functionality.Tables.Mangle,
		enabled:                true,
	}

	if initialized.IPv6.Redirect, err = c.Redirect.Initialize(loggerIPv6, e.IPv6, true); err != nil {
		return initialized, errors.Wrap(err, "unable to initialize IPv6 Redirect configuration")
	}

	return initialized, nil
}

func DefaultConfig() Config {
	return Config{
		Owner: Owner{UID: "5678"},
		Redirect: Redirect{
			NamePrefix: IptablesChainsPrefix,
			Inbound: TrafficFlow{
				Enabled:           true,
				Port:              DefaultRedirectInbountPort,
				ChainName:         "INBOUND",
				RedirectChainName: "INBOUND_REDIRECT",
				ExcludePorts:      []uint16{},
				IncludePorts:      []uint16{},
			},
			Outbound: TrafficFlow{
				Enabled:           true,
				Port:              DefaultRedirectOutboundPort,
				ChainName:         "OUTBOUND",
				RedirectChainName: "OUTBOUND_REDIRECT",
				ExcludePorts:      []uint16{},
				IncludePorts:      []uint16{},
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
		IPv6:               true,
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
		Comment: Comment{
			Disabled: false,
		},
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

// parseExcludePortsForUIDs parses a slice of strings representing port
// exclusion rules based on UIDs and returns a slice of Exclusion structs.
//
// Each input string should follow the format: <protocol:>?<ports:>?<uids>.
// This means the string can contain optional protocol and port values,
// followed by mandatory UID values. Examples of valid formats include:
//   - "tcp:22:1000-2000" (TCP protocol, port 22, UIDs from 1000 to 2000)
//   - "udp:53:1001" (UDP protocol, port 53, UID 1001)
//   - "80:1002" (Any protocol, port 80, UID 1002)
//   - "1003" (Any protocol, any port, UID 1003)
func parseExcludePortsForUIDs(exclusionRules []string) ([]Exclusion, error) {
	var result []Exclusion

	for _, elem := range exclusionRules {
		parts := strings.Split(elem, ":")
		if len(parts) == 0 || len(parts) > 3 {
			return nil, errors.Errorf(
				"invalid format for excluding ports by UIDs: '%s'. Expected format: <protocol:>?<ports:>?<uids>",
				elem,
			)
		}

		var portValuesOrRange, protocolOpts, uidValuesOrRange string

		switch len(parts) {
		case 1:
			protocolOpts = "*"
			portValuesOrRange = "*"
			uidValuesOrRange = parts[0]
		case 2:
			protocolOpts = "*"
			portValuesOrRange = parts[0]
			uidValuesOrRange = parts[1]
		case 3:
			protocolOpts = parts[0]
			portValuesOrRange = parts[1]
			uidValuesOrRange = parts[2]
		}

		if uidValuesOrRange == "*" {
			return nil, errors.New("wildcard '*' is not allowed for UIDs")
		}

		if portValuesOrRange == "*" || portValuesOrRange == "" {
			portValuesOrRange = "1-65535"
		}

		if err := validateUintValueOrRange(portValuesOrRange); err != nil {
			return nil, errors.Wrap(err, "invalid port range")
		}

		if strings.Contains(uidValuesOrRange, ",") {
			return nil, errors.Errorf(
				"invalid UID entry: '%s'. It should either be a single item or a range",
				uidValuesOrRange,
			)
		}

		if err := validateUintValueOrRange(uidValuesOrRange); err != nil {
			return nil, errors.Wrap(err, "invalid UID range")
		}

		var protocols []ProtocolL4
		if protocolOpts == "" || protocolOpts == "*" {
			protocols = []ProtocolL4{ProtocolTCP, ProtocolUDP}
		} else {
			for _, s := range strings.Split(protocolOpts, ",") {
				if p := ParseProtocolL4(s); p != ProtocolUndefined {
					protocols = append(protocols, p)
					continue
				}

				return nil, errors.Errorf(
					"invalid or unsupported protocol: '%s'",
					s,
				)
			}
		}

		for _, p := range protocols {
			ports := strings.ReplaceAll(portValuesOrRange, "-", ":")
			uids := strings.ReplaceAll(uidValuesOrRange, "-", ":")

			result = append(result, Exclusion{
				Ports:    ValueOrRangeList(ports),
				UIDs:     ValueOrRangeList(uids),
				Protocol: p,
			})
		}
	}

	return result, nil
}

// parseExcludePortsForIPs parses a slice of strings representing port exclusion
// rules based on IP addresses and returns a slice of IPToPorts structs.
//
// This function currently allows each exclusion rule to be a valid IPv4 or IPv6
// address, with or without a CIDR suffix. It is designed to potentially support
// more complex exclusion rules in the future.
func parseExcludePortsForIPs(
	exclusionRules []string,
	ipv6 bool,
) ([]Exclusion, error) {
	var result []Exclusion

	for _, rule := range exclusionRules {
		if rule == "" {
			return nil, errors.New(
				"invalid exclusion rule: the rule cannot be empty",
			)
		}

		for _, address := range strings.Split(rule, ",") {
			err, isExpectedIPVersion := validateIP(address, ipv6)
			if err != nil {
				return nil, errors.Wrap(err, "invalid exclusion rule")
			}

			if isExpectedIPVersion {
				result = append(result, Exclusion{Address: address})
			}
		}
	}

	return result, nil
}

// validateUintValueOrRange validates whether a given string represents a valid
// single uint16 value or a range of uint16 values. The input string can contain
// multiple comma-separated values or ranges (e.g., "80,1000-2000").
func validateUintValueOrRange(valueOrRange string) error {
	for _, element := range strings.Split(valueOrRange, ",") {
		for _, port := range strings.Split(element, "-") {
			if _, err := parseUint16(port); err != nil {
				return errors.Wrapf(
					err,
					"validation failed for value or range '%s'",
					valueOrRange,
				)
			}
		}
	}

	return nil
}

// validateIP validates an IP address or CIDR and checks if it matches the
// expected IP version (IPv4 or IPv6).
func validateIP(address string, ipv6 bool) (error, bool) {
	// Attempt to parse the address as a CIDR.
	ip, _, err := net.ParseCIDR(address)
	// If parsing as CIDR fails, attempt to parse it as a plain IP address.
	if err != nil {
		ip = net.ParseIP(address)
	}

	// If parsing as both CIDR and IP address fails, return an error with a
	// message.
	if ip == nil {
		return errors.Errorf(
			"invalid IP address: '%s'. Expected format: <ip> or <ip>/<cidr> (e.g., 10.0.0.1, 172.16.0.0/16, fe80::1, fe80::/10)",
			address,
		), false
	}

	// Check if the IP version matches the expected IP version.
	// For IPv4, ip.To4() will not be nil. For IPv6, ip.To4() will be nil.
	return nil, ipv6 == (ip.To4() == nil)
}

// parseUint16 parses a string representing a uint16 value and returns its
// uint16 representation.
func parseUint16(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid uint16 value: '%s'", port)
	}

	return uint16(parsedPort), nil
}

// hasLocalIPv6 checks if the local system has an active non-loopback IPv6
// address. It scans through all network interfaces to find any IPv6 address
// that is not a loopback address.
func hasLocalIPv6() (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() == nil {
			return true, nil
		}
	}

	return false, errors.New("no local IPv6 addresses detected")
}

// configureIPv6OutboundAddress sets up a dedicated IPv6 address (::6) on the
// loopback interface ("lo") for our transparent proxy functionality.
//
// Background:
//   - The default IPv6 configuration (prefix length 128) only allows binding to
//     the loopback address (::1).
//   - Our transparent proxy requires a distinct IPv6 address (::6 in this case)
//     to identify traffic processed by the kuma-dp sidecar.
//   - This identification allows for further processing and avoids redirection
//     loops.
//
// This function is equivalent to running the command:
// `ip -6 addr add "::6/128" dev lo`
func configureIPv6OutboundAddress() error {
	link, err := netlink.LinkByName("lo")
	if err != nil {
		return errors.Wrap(err, "failed to find loopback interface ('lo')")
	}

	// Equivalent to "::6/128"
	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP("::6"),
			Mask: net.CIDRMask(128, 128),
		},
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		// Address already exists, ignore error and continue
		if strings.Contains(strings.ToLower(err.Error()), "file exists") {
			return nil
		}

		return errors.Wrapf(err, "failed to add IPv6 address %s to loopback interface", addr.IPNet)
	}

	return nil
}
