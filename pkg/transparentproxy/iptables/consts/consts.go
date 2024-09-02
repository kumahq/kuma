package consts

import (
	"regexp"
	"strings"

	k8s_version "k8s.io/apimachinery/pkg/util/version"
)

const (
	Iptables  = "iptables"
	Ip6tables = "ip6tables"
)

// IptablesCommandByFamily maps a boolean value indicating IPv4 (false) or IPv6
// (true) usage to the corresponding iptables command name. This allows for code
// to be written generically without duplicating logic for both IPv4 and IPv6
var IptablesCommandByFamily = map[bool]string{
	false: Iptables,
	true:  Ip6tables,
}

// IPTypeMap is a map that translates a boolean value to a string representing
// the type of IP address (IPv4 or IPv6). The key is a boolean where 'false'
// corresponds to "IPv4" and 'true' corresponds to "IPv6"
var IPTypeMap = map[bool]string{
	false: "IPv4",
	true:  "IPv6",
}

// Default ports used for iptables redirection
const (
	DefaultRedirectInbountPort  uint16 = 15006
	DefaultRedirectOutboundPort uint16 = 15001
	DefaultRedirectDNSPort      uint16 = 15053
)

const (
	DNSPort           uint16 = 53
	LocalhostIPv4            = "127.0.0.1"
	LocalhostCIDRIPv4        = "127.0.0.1/32"
	LocalhostIPv6            = "[::1]"
	LocalhostCIDRIPv6        = "::1/128"
	// InboundPassthroughSourceAddressCIDRIPv4
	// TODO (bartsmykla): add some description
	InboundPassthroughSourceAddressCIDRIPv4 = "127.0.0.6/32"
	InboundPassthroughSourceAddressCIDRIPv6 = "::6/128"
	OutputLogPrefix                         = "OUTPUT:"
	PreroutingLogPrefix                     = "PREROUTING:"
)

type ProtocolL4 string

const (
	ProtocolUDP ProtocolL4 = "udp"
	ProtocolTCP ProtocolL4 = "tcp"
	// ProtocolUndefined represents an undefined or unsupported protocol
	ProtocolUndefined ProtocolL4 = ""
)

// ParseProtocolL4 parses a string and returns the corresponding ProtocolL4
// constant. If the input string is not "udp" or "tcp", it returns
// ProtocolUndefined
func ParseProtocolL4(s string) ProtocolL4 {
	switch s := strings.ToLower(strings.TrimSpace(s)); s {
	case "udp", "tcp":
		return ProtocolL4(s)
	default:
		return ProtocolUndefined
	}
}

type TableName string

const (
	TableNat    TableName = "nat"
	TableRaw    TableName = "raw"
	TableMangle TableName = "mangle"
)

const (
	ChainPrerouting   = "PREROUTING"
	ChainInput        = "INPUT"
	ChainForward      = "FORWARD"
	ChainOutput       = "OUTPUT"
	ChainPostrouting  = "POSTROUTING"
	ChainDockerOutput = "DOCKER_OUTPUT"
)

// DockerOutputChainRegex is a regular expression used to identify the presence
// of a custom chain named "DOCKER_OUTPUT" in iptables rules
var DockerOutputChainRegex = regexp.MustCompile(`(?m)^:DOCKER_OUTPUT`)

const (
	FlagTable   = "-t"
	FlagMatch   = "-m"
	FlagHelp    = "-h"
	FlagVersion = "--version" // there is no short version of this flag

	// commands
	FlagAppend   = "-A"
	FlagInsert   = "-I"
	FlagNewChain = "-N"

	// iptables-restore
	FlagWait         = "--wait"
	FlagWaitInterval = "--wait-interval"
	FlagNoFlush      = "--noflush"
	FlagTest         = "--test"
)

// FlagVariationsMap maps a flag name (e.g., "-t") to a map containing its long
// (true) and short (false) flag representations. This allows for code to easily
// look up the appropriate flag based on desired usage (short or long)
var FlagVariationsMap = map[string]map[bool]string{
	FlagAppend: {
		true:  "--append",
		false: FlagAppend,
	},
	FlagInsert: {
		true:  "--insert",
		false: FlagInsert,
	},
	FlagNewChain: {
		true:  "--new-chain",
		false: FlagNewChain,
	},
}

const (
	ModuleOwner     = "owner"
	ModuleTcp       = "tcp"
	ModuleUdp       = "udp"
	ModuleComment   = "comment"
	ModuleConntrack = "conntrack"
	ModuleMultiport = "multiport"
)

type IptablesMode string

const (
	IptablesModeNft    IptablesMode = "nft"
	IptablesModeLegacy IptablesMode = "legacy"
)

// Regexp used to parse the result of `iptables --version` then used to map to
// with IptablesMode
var IptablesModeRegex = regexp.MustCompile(`(?m)^ip6?tables[\w-]*? v([0-9]+(?:\.[0-9]+)+)(?: \((.*?)\))?`)

// IptablesVersionWithLockfileEnv represents the iptables version (1.8.6) where
// the XTABLES_LOCKFILE environment variable was introduced
var IptablesVersionWithLockfileEnv = k8s_version.MustParseGeneric("1.8.6")

// Map IptablesMode to the mode taken from the result of `iptables --version`
var IptablesModeMap = map[string]IptablesMode{
	"":          IptablesModeLegacy, // i.e. iptables v1.6.1
	"legacy":    IptablesModeLegacy, // i.e. iptables v1.8.5 (legacy)
	"nf_tables": IptablesModeNft,    // i.e. iptables v1.8.9 (nf_tables)
}

// FallbackExecutablesSearchLocations is a list of directories to search for
// the iptables executables if it cannot be found in the user's PATH environment
// variable. This allows for some flexibility in environments where iptables
// may be installed in non-standard locations
var FallbackExecutablesSearchLocations = []string{
	"/usr/sbin",
	"/sbin",
	"/usr/bin",
	"/bin",
}

// Debug log level for iptables LOG jump target
// ref. https://git.netfilter.org/iptables/tree/extensions/libebt_log.c#n27
const LogLevelDebug uint16 = 7

// IptablesRuleCommentPrefix defines a consistent prefix used in iptables rule
// comments. This prefix helps identify and distinguish rules that were
// specifically generated by our transparent proxy
const IptablesRuleCommentPrefix = "kuma/mesh/transparent/proxy"

// IptablesChainsPrefix is the prefix used for naming custom iptables chains
// created for the transparent proxy. This prefix helps to clearly identify
// and differentiate the chains managed by the Kuma mesh from other iptables
// chains. The chains named with this prefix are used to apply specific rules
// necessary for the operation of the transparent proxy
const IptablesChainsPrefix = "KUMA_MESH"

// Default user identification constants used for running kuma-dp. These defaults
// are utilized when no specific user is provided
const (
	OwnerDefaultUID      = "5678"
	OwnerDefaultUsername = "kuma-dp"
)

// EnvVarXtablesLockfile is the name of the environment variable that specifies
// the path to the lock file used by iptables-legacy version 1.8.6 and later
const EnvVarXtablesLockfile = "XTABLES_LOCKFILE"

const (
	// Before iptables-legacy v1.8.6 iptables were using hardcoded path to lock
	// file which was /run/xtables.lock
	PathLegacyXtablesLock = "/run/xtables.lock"
	PathDevNull           = "/dev/null"
	PathNSSwitchConf      = "/etc/nsswitch.conf"
)
