package config

import (
	"context"
	"encoding/json"
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

	core_config "github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
)

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

var _ json.Unmarshaler = (*Port)(nil)

type Port uint16

func (p *Port) String() string { return strconv.Itoa(int(*p)) }

func (p *Port) Type() string { return "uint16" }

func (p *Port) Set(s string) error {
	var err error

	if *p, err = parsePort(s); err != nil {
		return err
	}

	return nil
}

func (p *Port) UnmarshalJSON(bs []byte) error { return p.Set(string(bs)) }

var _ json.Unmarshaler = &Ports{}

type Ports []Port

func (p *Ports) String() string {
	var ports []string
	for _, port := range *p {
		ports = append(ports, strconv.Itoa(int(port)))
	}
	return strings.Join(ports, ",")
}

func (p *Ports) Type() string { return "uint16[,...]" }

func (p *Ports) Set(s string) error {
	*p = nil

	if s = strings.TrimSpace(s); s == "" {
		return nil
	}

	for _, port := range strings.Split(s, ",") {
		trimmedPort := strings.TrimSpace(port)
		if trimmedPort == "" {
			continue
		}

		parsedPort, err := parsePort(trimmedPort)
		if err != nil {
			return err
		}

		*p = append(*p, parsedPort)
	}

	return nil
}

func (p *Ports) UnmarshalJSON(bs []byte) error {
	var jsonValue interface{}

	if err := json.Unmarshal(bs, &jsonValue); err != nil {
		return err
	}

	switch typedValue := jsonValue.(type) {
	case []interface{}:
		var values []string
		for _, item := range typedValue {
			switch i := item.(type) {
			case string, float64:
				values = append(values, fmt.Sprint(i))
			}
		}
		return p.Set(strings.Join(values, ","))
	case string, float64:
		return p.Set(fmt.Sprint(typedValue))
	}

	return p.Set(string(bs))
}

type Exclusion struct {
	Protocol consts.ProtocolL4
	Address  string
	UIDs     ValueOrRangeList
	Ports    ValueOrRangeList
}

// TrafficFlow is a struct for Inbound/Outbound configuration
type TrafficFlow struct {
	Enabled                       bool     `json:"enabled"`                                                          // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_ENABLED, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_ENABLED
	Port                          Port     `json:"port"`                                                             // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_PORT, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_PORT
	ChainName                     string   `json:"-" split_words:"true"`                                             // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_CHAIN_NAME, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_CHAIN_NAME
	RedirectChainName             string   `json:"-" split_words:"true"`                                             // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_REDIRECT_CHAIN_NAME, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_REDIRECT_CHAIN_NAME
	IncludePorts                  Ports    `json:"includePorts,omitempty" split_words:"true"`                        // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_INCLUDE_PORTS, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_INCLUDE_PORTS
	ExcludePorts                  Ports    `json:"excludePorts,omitempty" split_words:"true"`                        // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_EXCLUDE_PORTS, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_EXCLUDE_PORTS
	ExcludePortsForUIDs           []string `json:"excludePortsForUIDs,omitempty" envconfig:"exclude_ports_for_uids"` // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_EXCLUDE_PORTS_FOR_UIDS, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_EXCLUDE_PORTS_FOR_UIDS
	ExcludePortsForIPs            []string `json:"excludePortsForIPs,omitempty" envconfig:"exclude_ports_for_ips"`   // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_EXCLUDE_PORTS_FOR_IPS, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_EXCLUDE_PORTS_FOR_IPS
	InsertRedirectInsteadOfAppend bool     `json:"insertRedirectInsteadOfAppend,omitempty" split_words:"true"`       // KUMA_TRANSPARENT_PROXY_REDIRECT_INBOUND_INSERT_REDIRECT_INSTEAD_OF_APPEND, KUMA_TRANSPARENT_PROXY_REDIRECT_OUTBOUND_INSERT_REDIRECT_INSTEAD_OF_APPEND
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
		return initialized, errors.Wrap(err, "parsing excluded outbound ports for uids failed")
	}

	excludePortsForIPs, err := parseExcludePortsForIPs(c.ExcludePortsForIPs, ipv6)
	if err != nil {
		return initialized, errors.Wrap(err, "parsing excluded outbound ports for IPs failed")
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
	Port              Port
	ChainName         string
	RedirectChainName string
}

type DNS struct {
	Enabled                bool   `json:"enabled"`                                   // KUMA_TRANSPARENT_PROXY_REDIRECT_DNS_ENABLED
	Port                   Port   `json:"port"`                                      // KUMA_TRANSPARENT_PROXY_REDIRECT_DNS_PORT
	CaptureAll             bool   `json:"captureAll" split_words:"true"`             // KUMA_TRANSPARENT_PROXY_REDIRECT_DNS_CAPTURE_ALL
	SkipConntrackZoneSplit bool   `json:"skipConntrackZoneSplit" split_words:"true"` // KUMA_TRANSPARENT_PROXY_REDIRECT_DNS_SKIP_CONNTRACK_ZONE_SPLIT
	ResolvConfigPath       string `json:"resolvConfigPath" split_words:"true"`       // KUMA_TRANSPARENT_PROXY_REDIRECT_DNS_RESOLV_CONFIG_PATH
	// The iptables chain where the upstream DNS requests should be directed to.
	// It is only applied for IP V4. Use with care. (default "RETURN")
	UpstreamTargetChain string `json:"-" ignored:"true"`
}

type InitializedDNS struct {
	DNS
	Servers            []string
	ConntrackZoneSplit bool
	Enabled            bool
}

// Initialize initializes the ServersIPv4 and ServersIPv6 fields by parsing
// the nameservers from the file specified in the ResolvConfigPath field of
// the input DNS struct
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

	if !c.SkipConntrackZoneSplit {
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
		return initialized, errors.Wrapf(err, "unable to read file %s", c.ResolvConfigPath)
	}

	// Loop through each DNS server address parsed from the resolv.conf file
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
			consts.IPTypeMap[ipv6],
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
	Networks []string `json:"networks,omitempty"` // KUMA_TRANSPARENT_PROXY_REDIRECT_VNET_NETWORKS
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
			return InitializedVNet{}, errors.Errorf("invalid virtual network definition: %s", network)
		}

		address, _, err := net.ParseCIDR(pair[1])
		if err != nil {
			return InitializedVNet{}, errors.Wrapf(err, "invalid CIDR definition for %s", pair[1])
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
	// virtual network configurations
	InterfaceCIDRs map[string]string
}

type Redirect struct {
	// NamePrefix is a prefix which will be used go generate chains name
	NamePrefix string      `json:"-" ignored:"true"`
	Inbound    TrafficFlow `json:"inbound"`
	Outbound   TrafficFlow `json:"outbound"`
	DNS        DNS         `json:"dns"`
	VNet       VNet        `json:"vnet"`
}

// Custom Marshal logic to omit the VNet field from the JSON output if it contains
// no networks. This approach ensures that empty fields are not rendered, since
// we're working with a value type (Redirect) and not pointers, and we only include
// VNet when it has meaningful data
func (c Redirect) MarshalJSON() ([]byte, error) {
	type ConfigAlias Redirect

	type ConfigAliasOmitEmpty struct {
		ConfigAlias
		VNet any `json:"vnet,omitempty"`
	}

	result := ConfigAliasOmitEmpty{
		ConfigAlias: ConfigAlias(c),
		VNet:        c.VNet,
	}

	if len(c.VNet.Networks) == 0 {
		result.VNet = nil
	}

	return json.Marshal(result)
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
	Enabled              bool   `json:"enabled"`                                                   // KUMA_TRANSPARENT_PROXY_EBPF_ENABLED
	InstanceIP           string `json:"instanceIP" envconfig:"instance_ip"`                        // KUMA_TRANSPARENT_PROXY_EBPF_INSTANCE_IP
	InstanceIPEnvVarName string `json:"instanceIPEnvVarName" envconfig:"instance_ip_env_var_name"` // KUMA_TRANSPARENT_PROXY_EBPF_INSTANCE_IP_ENV_VAR_NAME
	BPFFSPath            string `json:"bpffsPath" envconfig:"bpffs_path"`                          // KUMA_TRANSPARENT_PROXY_EBPF_BPFFS_PATH
	CgroupPath           string `json:"cgroupPath" split_words:"true"`                             // KUMA_TRANSPARENT_PROXY_EBPF_CGROUP_PATH
	ProgramsSourcePath   string `json:"programsSourcePath" split_words:"true"`                     // KUMA_TRANSPARENT_PROXY_EBPF_PROGRAM_SOURCE_PATH
	// The name of network interface which TC ebpf programs should bind to,
	// when not provided, we'll try to automatically determine it
	TCAttachIface string `json:"tcAttachIface" envconfig:"tc_attach_iface"` // KUMA_TRANSPARENT_PROXY_EBPF_TC_ATTACH_IFACE
}

func (c Ebpf) Initialize() (InitializedEbpf, error) {
	if !c.Enabled {
		return InitializedEbpf{}, nil
	}

	instanceIP := c.InstanceIP
	defaultInstanceIPEnvVarName := "INSTANCE_IP"
	switch {
	case c.InstanceIP != "":
		break
	case c.InstanceIPEnvVarName != "" && os.Getenv(c.InstanceIPEnvVarName) == "":
		return InitializedEbpf{}, errors.Errorf(
			"environment variable '%s' does not contain an instance IP",
			c.InstanceIPEnvVarName,
		)
	case c.InstanceIPEnvVarName == "" && os.Getenv(defaultInstanceIPEnvVarName) == "":
		return InitializedEbpf{}, errors.New(
			"no instance IP or environment variable containing instance IP specified",
		)
	case c.InstanceIPEnvVarName == "":
		instanceIP = os.Getenv(defaultInstanceIPEnvVarName)
	default:
		instanceIP = os.Getenv(c.InstanceIPEnvVarName)
	}

	return InitializedEbpf{
		Enabled:            c.Enabled,
		InstanceIP:         instanceIP,
		BPFFSPath:          c.BPFFSPath,
		CgroupPath:         c.CgroupPath,
		ProgramsSourcePath: c.ProgramsSourcePath,
		TCAttachIface:      c.TCAttachIface,
	}, nil
}

type InitializedEbpf struct {
	Enabled            bool
	InstanceIP         string
	BPFFSPath          string
	CgroupPath         string
	ProgramsSourcePath string
	TCAttachIface      string
}

type Log struct {
	// Enabled determines whether iptables rules logging is activated. When
	// true, each packet matching an iptables rule will have its details logged,
	// aiding in diagnostics and monitoring of packet flows.
	Enabled bool `json:"enabled"` // KUMA_TRANSPARENT_PROXY_LOG_ENABLED
	// Level specifies the log level for iptables logging as defined by
	// netfilter. This level controls the verbosity and detail of the log
	// entries for matching packets. Higher values increase the verbosity.
	// Commonly used levels are: 1 (alerts), 4 (warnings), 5 (notices),
	// 7 (debugging). The exact behavior can depend on the system's syslog
	// configuration
	Level uint16 `json:"level"` // KUMA_TRANSPARENT_PROXY_LOG_LEVEL
}

type Retry struct {
	// MaxRetries specifies the number of retries after the initial attempt.
	// A value of 0 means no retries, and only the initial attempt will be made.
	MaxRetries int `json:"maxRetries" split_words:"true"` // KUMA_TRANSPARENT_PROXY_RETRY_MAX_RETRIES
	// SleepBetweenRetries defines the duration to wait between retry attempts.
	// This delay helps in situations where immediate retries may not be
	// beneficial, allowing time for transient issues to resolve.
	SleepBetweenRetries config_types.Duration `json:"sleepBetweenRetries" split_words:"true"` // KUMA_TRANSPARENT_PROXY_RETRY_SLEEP_BETWEEN_RETRIES
}

// Comments struct contains the configuration for iptables rule comments.
// It includes an option to enable or disable comments.
type Comments struct {
	Disabled bool `json:"disabled"` // KUMA_TRANSPARENT_PROXY_COMMENTS_DISABLED
}

// InitializedComments struct contains the processed configuration for iptables
// rule comments. It indicates whether comments are enabled and the prefix to
// use for comment text
type InitializedComments struct {
	// Enabled indicates whether iptables rule comments are enabled based on
	// the initial configuration and system capabilities
	Enabled bool
	// Prefix defines the prefix to be used for comments on iptables rules,
	// aiding in identifying and organizing rules created by the transparent
	// proxy
	Prefix string
}

// Initialize processes the Comments configuration and determines whether
// iptables rule comments should be enabled. It checks the system's
// functionality to see if the comment module is available and returns
// an InitializedComments struct with the result
func (c Comments) Initialize(e InitializedExecutablesIPvX) InitializedComments {
	return InitializedComments{
		Enabled: !c.Disabled && e.Functionality.Modules.Comment,
		Prefix:  consts.IptablesRuleCommentPrefix,
	}
}

var _ core_config.Config = Config{}

type Config struct {
	core_config.BaseConfig

	KumaDPUser string   `json:"kumaDPUser" envconfig:"kuma_dp_user"` // KUMA_TRANSPARENT_PROXY_KUMA_DP_USER
	Redirect   Redirect `json:"redirect"`
	Ebpf       Ebpf     `json:"ebpf"`
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
	DropInvalidPackets bool `json:"dropInvalidPackets,omitempty" split_words:"true"` // KUMA_TRANSPARENT_PROXY_DROP_INVALID_PACKETS
	// RuntimeStdout is the place where Any debugging, runtime information
	// will be placed (os.Stdout by default)
	RuntimeStdout io.Writer `json:"-" ignored:"true"`
	// RuntimeStderr is the place where error, runtime information will be
	// placed (os.Stderr by default)
	RuntimeStderr io.Writer `json:"-" ignored:"true"`
	// Verbose when set will generate iptables configuration with longer
	// argument/flag names, additional comments etc
	Verbose bool `json:"verbose,omitempty"` // KUMA_TRANSPARENT_PROXY_VERBOSE
	// DryRun when set will not execute, but just display instructions which
	// otherwise would have served to install transparent proxy
	DryRun bool `json:"dryRun,omitempty" split_words:"true"` // KUMA_TRANSPARENT_PROXY_DRY_RUN
	// Log configures logging for iptables rules using the LOG chain. When
	// enabled, this setting causes the kernel to log details about packets that
	// match the iptables rules, including IP/IPv6 headers. The logs are useful
	// for debugging and can be accessed via tools like dmesg or syslog. The
	// logging behavior is defined by the nested Log struct
	Log Log `json:"log"`
	// Wait is the amount of time, in seconds, that the application should wait
	// for the xtables exclusive lock before exiting. If the lock is not
	// available within the specified time, the application will exit with
	// an error. Default value *(0) means wait forever. To disable this behavior
	// and exit immediately if the xtables lock is not available, set this to nil
	Wait uint `json:"wait"` // KUMA_TRANSPARENT_PROXY_WAIT
	// WaitInterval is the amount of time, in microseconds, that iptables should
	// wait between each iteration of the lock acquisition loop. This can be
	// useful if the xtables lock is being held by another application for
	// a long time, and you want to reduce the amount of CPU that iptables uses
	// while waiting for the lock
	WaitInterval uint `json:"waitInterval" split_words:"true"` // KUMA_TRANSPARENT_PROXY_WAIT_INTERVAL
	// Retry allows you to configure the number of times that the system should
	// retry an installation if it fails
	Retry Retry `json:"retry"`
	// StoreFirewalld when set, configures firewalld to store the generated
	// iptables rules
	StoreFirewalld bool `json:"storeFirewalld,omitempty" split_words:"true"` // KUMA_TRANSPARENT_PROXY_STORE_FIREWALLD
	// Executables field holds configuration for the executables used to
	// interact with iptables (or ip6tables). It can handle both nft (nftables)
	// and legacy iptables modes, and supports IPv4 and IPv6 versions
	Executables Executables `json:"iptablesExecutables" envconfig:"iptables_executables"` // KUMA_TRANSPARENT_PROXY_IPTABLES_EXECUTABLES
	// Comments configures the prefix and enable/disable status for iptables rule
	// comments. This setting helps in identifying and organizing iptables rules
	// created by the transparent proxy, making them easier to manage and debug
	Comments Comments `json:"comments"`
	// IPFamilyMode specifies the IP family mode to be used by the
	// configuration. It determines whether the system operates in dualstack
	// mode (supporting both IPv4 and IPv6) or IPv4-only mode. This setting is
	// crucial for environments where both IP families are in use, ensuring that
	// the correct iptables rules are applied for the specified IP family
	IPFamilyMode IPFamilyMode `json:"ipFamilyMode" envconfig:"ip_family_mode"` // KUMA_TRANSPARENT_PROXY_IP_FAMILY_MODE
	CNIMode      bool         `json:"cniMode,omitempty" envconfig:"cni_mode"`  // KUMA_TRANSPARENT_PROXY_CNI_MODE
}

func (c Config) WithStdout(stdout io.Writer) Config {
	c.RuntimeStdout = stdout
	return c
}

// Custom Marshal logic to avoid rendering empty values for specific fields.
// Since we're working with a value type (Config) rather than pointers, this approach
// ensures that unnecessary fields (like ebpf, comments, log, and executables) are omitted
// from the JSON output when they are not enabled or contain no meaningful data
func (c Config) MarshalJSON() ([]byte, error) {
	type ConfigAlias Config

	type ConfigAliasOmitEmpty struct {
		ConfigAlias
		Ebpf        any `json:"ebpf,omitempty"`
		Comments    any `json:"comments,omitempty"`
		Log         any `json:"log,omitempty"`
		Executables any `json:"iptablesExecutables,omitempty"`
	}

	result := ConfigAliasOmitEmpty{
		ConfigAlias: ConfigAlias(c),
		Ebpf:        c.Ebpf,
		Comments:    c.Comments,
		Log:         c.Log,
		Executables: c.Executables,
	}

	if !c.Ebpf.Enabled {
		result.Ebpf = nil
	}

	if !c.Comments.Disabled {
		result.Comments = nil
	}

	if !c.Log.Enabled {
		result.Log = nil
	}

	if len(getNonEmptyPaths(&c.Executables)) == 0 {
		result.Executables = nil
	}

	return json.Marshal(result)
}

func (c Config) InitializeKumaDPUser() (string, error) {
	switch {
	case c.CNIMode && c.KumaDPUser != "":
		return c.KumaDPUser, nil
	case c.CNIMode:
		return consts.OwnerDefaultUID, nil
	case c.KumaDPUser != "":
		if v, ok := findUserUID(c.KumaDPUser); ok {
			return v, nil
		}

		return "", errors.Errorf(
			"the specified UID or username ('%s') does not refer to a valid user on the host",
			c.KumaDPUser,
		)
	}

	if v, ok := findUserUID(consts.OwnerDefaultUID); ok {
		return v, nil
	}

	if v, ok := findUserUID(consts.OwnerDefaultUsername); ok {
		return v, nil
	}

	return "", errors.Errorf(
		"no UID or username provided, and user with the default UID ('%s') or username ('%s') could not be found",
		consts.OwnerDefaultUID,
		consts.OwnerDefaultUsername,
	)
}

type IPFamilyMode string

func (e *IPFamilyMode) UnmarshalJSON(bs []byte) error {
	var value string

	if err := json.Unmarshal(bs, &value); err != nil {
		return errors.Wrapf(err, "value '%s' is not a valid IPFamilyMode", bs)
	}

	if err := e.Set(value); err != nil {
		return errors.Wrapf(err, "value '%s' is not a valid IPFamilyMode", value)
	}

	return nil
}

const (
	IPFamilyModeDualStack IPFamilyMode = "dualstack"
	IPFamilyModeIPv4      IPFamilyMode = "ipv4"
)

// String returns the string representation of the IPFamilyMode.
// This is used both by fmt.Print and by Cobra in help text
func (e *IPFamilyMode) String() string {
	return string(*e)
}

// Type returns the type of the IPFamilyMode.
// This is only used in help text by Cobra
func (e *IPFamilyMode) Type() string {
	return "string"
}

// Set assigns the IPFamilyMode based on the provided value. It validates the
// input and sets the appropriate mode or returns an error if the input is invalid
func (e *IPFamilyMode) Set(v string) error {
	switch strings.ToLower(v) {
	case "": // Default value is "dualstack"
		*e = IPFamilyModeDualStack
	case string(IPFamilyModeDualStack), string(IPFamilyModeIPv4):
		*e = IPFamilyMode(v)
	default:
		return errors.Errorf("must be one of %s", AllowedIPFamilyModes())
	}

	return nil
}

func AllowedIPFamilyModes() string {
	return fmt.Sprintf("'%s' or '%s'", IPFamilyModeDualStack, IPFamilyModeIPv4)
}

// InitializedConfigIPvX extends the Config struct by adding fields that require
// additional logic to retrieve their values. These values typically involve
// interacting with the system or external resources
type InitializedConfigIPvX struct {
	Config
	// Logger is utilized for detailed logging throughout the lifecycle of the
	// InitializedConfigIPvX. This includes specific logging for iptables
	// operations such as rule setup, modification, and restoration. The Logger
	// in this struct ensures detailed, step-by-step logs are available for
	// operations related to the corresponding IP version (IPv4 or IPv6), aiding
	// in diagnostics and debugging
	Logger Logger
	// Redirect is an InitializedRedirect struct containing the initialized
	// redirection configuration. If DNS redirection is enabled this includes
	// the DNS servers retrieved from the specified resolv.conf file
	// (/etc/resolv.conf by default)
	Redirect InitializedRedirect
	// Executables field holds the initialized version of Config.Executables.
	// It attempts to locate the actual executable paths on the system based on
	// the provided configuration and verifies their functionality
	Executables InitializedExecutablesIPvX
	// DropInvalidPackets when enabled, kuma-dp will configure iptables to drop
	// packets that are considered invalid. This is useful in scenarios where
	// out-of-order packets bypass DNAT by iptables and reach the application
	// directly, causing connection resets. This field is set during
	// configuration initialization and considers whether the mangle table is
	// available for the corresponding IP version (IPv4 or IPv6)
	DropInvalidPackets bool
	// LoopbackInterfaceName represents the name of the loopback interface which
	// will be used to construct outbound iptable rules for outbound (i.e.
	// -A KUMA_MESH_OUTBOUND -s 127.0.0.6/32 -o lo -j RETURN)
	LoopbackInterfaceName string
	// LocalhostCIDR is a string representing the CIDR notation of the localhost
	// address for the given IP version (IPv4 or IPv6). This is used to
	// construct rules related to the loopback interface
	LocalhostCIDR string
	// InboundPassthroughCIDR is a string representing the CIDR notation of the
	// address used for inbound passthrough traffic. This is used to construct
	// rules allowing specific traffic to bypass normal proxying
	InboundPassthroughCIDR string
	// Comments holds the processed configuration for iptables rule comments,
	// indicating whether comments are enabled and the prefix to use for comment
	// text. This helps in identifying and organizing iptables rules created by
	// the transparent proxy, making them easier to manage and debug
	Comments   InitializedComments
	Ebpf       InitializedEbpf
	KumaDPUser string

	enabled bool
}

// Enabled returns the state of the 'enabled' field, indicating whether the
// IP version-specific configuration is enabled
func (c InitializedConfigIPvX) Enabled() bool {
	return c.enabled
}

type InitializedConfig struct {
	// Logger is utilized for recording general logs during the lifecycle of the
	// InitializedConfig, including the initialization and finalization phases
	// of the transparent proxy installation process. This logger is used to log
	// high-level information and statuses, while more specific logging related
	// to iptables operations is handled by the Logger in InitializedConfigIPvX
	Logger Logger
	// DryRun when set will not execute, but just display instructions which
	// otherwise would have served to install transparent proxy
	DryRun bool
	// IPv4 contains the initialized configuration specific to IPv4. This
	// includes all settings, executables, and rules relevant to IPv4 iptables
	// management
	IPv4 InitializedConfigIPvX
	// IPv6 contains the initialized configuration specific to IPv6. This
	// includes all settings, executables, and rules relevant to IPv6 ip6tables
	// management
	IPv6 InitializedConfigIPvX
}

func (c Config) Initialize(ctx context.Context) (InitializedConfig, error) {
	var kumaDPUser string
	var err error

	if kumaDPUser, err = c.InitializeKumaDPUser(); err != nil {
		return InitializedConfig{}, err
	}

	l := Logger{
		stdout: c.RuntimeStdout,
		stderr: c.RuntimeStderr,
		maxTry: c.Retry.MaxRetries + 1,
	}

	loggerIPv4 := l.WithPrefix(consts.IptablesCommandByFamily[false])
	loggerIPv6 := l.WithPrefix(consts.IptablesCommandByFamily[true])

	loopbackInterfaceName, err := getLoopbackInterfaceName()
	if err != nil {
		return InitializedConfig{}, errors.Wrap(err, "unable to initialize loopback interface name")
	}

	executablesIPv4, err := c.Executables.InitializeIPv4(ctx, loggerIPv4, c)
	if err != nil {
		return InitializedConfig{}, errors.Wrap(err, "unable to initialize IPv4 executables")
	}

	redirectIPv4, err := c.Redirect.Initialize(loggerIPv4, executablesIPv4, false)
	if err != nil {
		return InitializedConfig{}, errors.Wrap(err, "unable to initialize IPv4 redirect configuration")
	}

	initialized := InitializedConfig{
		Logger: l,
		DryRun: c.DryRun,
		IPv4: InitializedConfigIPvX{
			Config:                 c,
			Logger:                 loggerIPv4,
			Executables:            executablesIPv4,
			LoopbackInterfaceName:  loopbackInterfaceName,
			LocalhostCIDR:          consts.LocalhostCIDRIPv4,
			InboundPassthroughCIDR: consts.InboundPassthroughSourceAddressCIDRIPv4,
			Comments:               c.Comments.Initialize(executablesIPv4),
			DropInvalidPackets:     c.DropInvalidPackets && executablesIPv4.Functionality.Tables.Mangle,
			KumaDPUser:             kumaDPUser,
			Redirect:               redirectIPv4,
			enabled:                true,
		},
	}

	// IPv6 initialization is optional; failures will log warnings and continue with IPv4 only

	if c.IPFamilyMode == IPFamilyModeIPv4 {
		return initialized, nil
	}

	if ok, err := HasLocalIPv6(); !ok || err != nil {
		if c.Verbose {
			loggerIPv6.Warn("IPv6 initialization skipped due to missing or faulty IPv6 support:", err)
		}
		return initialized, nil
	}

	if err := configureIPv6OutboundAddress(); err != nil {
		if c.Verbose {
			loggerIPv6.Warn("failed to configure IPv6 outbound address; IPv6 rules will be skipped:", err)
		}
		return initialized, nil
	}

	executablesIPv6, err := c.Executables.InitializeIPv6(ctx, loggerIPv6, c, executablesIPv4.mode)
	if err != nil {
		loggerIPv6.Warn("failed to initialize IPv6 executables:", err)
		return initialized, nil
	}

	redirectIPv6, err := c.Redirect.Initialize(loggerIPv6, executablesIPv6, true)
	if err != nil {
		loggerIPv6.Warn("failed to initialize IPv6 redirect configuration:", err)
		return initialized, nil
	}

	initialized.IPv6 = InitializedConfigIPvX{
		Config:                 c,
		Logger:                 loggerIPv6,
		Executables:            executablesIPv6,
		LoopbackInterfaceName:  loopbackInterfaceName,
		LocalhostCIDR:          consts.LocalhostCIDRIPv6,
		InboundPassthroughCIDR: consts.InboundPassthroughSourceAddressCIDRIPv6,
		Comments:               c.Comments.Initialize(executablesIPv6),
		DropInvalidPackets:     c.DropInvalidPackets && executablesIPv6.Functionality.Tables.Mangle,
		KumaDPUser:             kumaDPUser,
		Redirect:               redirectIPv6,
		enabled:                true,
	}

	return initialized, nil
}

func DefaultConfig() Config {
	return Config{
		KumaDPUser: "",
		Redirect: Redirect{
			NamePrefix: consts.IptablesChainsPrefix,
			Inbound: TrafficFlow{
				Enabled:           true,
				Port:              Port(consts.DefaultRedirectInbountPort),
				ChainName:         "INBOUND",
				RedirectChainName: "INBOUND_REDIRECT",
				ExcludePorts:      Ports{},
				IncludePorts:      Ports{},
			},
			Outbound: TrafficFlow{
				Enabled:           true,
				Port:              Port(consts.DefaultRedirectOutboundPort),
				ChainName:         "OUTBOUND",
				RedirectChainName: "OUTBOUND_REDIRECT",
				ExcludePorts:      Ports{},
				IncludePorts:      Ports{},
			},
			DNS: DNS{
				Port:                   Port(consts.DefaultRedirectDNSPort),
				Enabled:                false,
				CaptureAll:             false,
				SkipConntrackZoneSplit: false,
				ResolvConfigPath:       "/etc/resolv.conf",
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
		RuntimeStdout:      os.Stdout,
		RuntimeStderr:      os.Stderr,
		Verbose:            false,
		DryRun:             false,
		Log: Log{
			Enabled: false,
			Level:   consts.LogLevelDebug,
		},
		Wait:         5,
		WaitInterval: 0,
		Retry: Retry{
			// Specifies the number of retries after the initial attempt,
			// totaling 5 tries
			MaxRetries:          4,
			SleepBetweenRetries: config_types.Duration{Duration: 2 * time.Second},
		},
		Executables: NewExecutables(),
		Comments: Comments{
			Disabled: false,
		},
		IPFamilyMode:   IPFamilyModeDualStack,
		StoreFirewalld: false,
		CNIMode:        false,
	}
}
