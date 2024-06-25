package consts

import (
	"regexp"
)

const (
	Long  = true
	Short = false
)

const (
	Iptables  = "iptables"
	Ip6tables = "ip6tables"
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
	UDP                                     = "udp"
	TCP                                     = "tcp"
)

type TableName string

const (
	TableNat    TableName = "nat"
	TableRaw    TableName = "raw"
	TableMangle TableName = "mangle"
)

const (
	ChainPrerouting  = "PREROUTING"
	ChainInput       = "INPUT"
	ChainForward     = "FORWARD"
	ChainOutput      = "OUTPUT"
	ChainPostrouting = "POSTROUTING"
)

// DockerOutputChainRegex is aregular expression used to identify the presence
// of a custom chain named "DOCKER_OUTPUT" in iptables rules.
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
)

var FlagsShortLongMap = map[string]map[bool]string{
	FlagAppend: {
		Long:  "--append",
		Short: FlagAppend,
	},
	FlagInsert: {
		Long:  "--insert",
		Short: FlagInsert,
	},
	FlagNewChain: {
		Long:  "--new-chain",
		Short: FlagNewChain,
	},
}

const (
	ModuleOwner     = "owner"
	ModuleTcp       = "tcp"
	ModuleUdp       = "udp"
	ModuleComment   = "comment"
	ModuleConntrack = "conntrack"
)

type IptablesMode string

const (
	IptablesModeNft    IptablesMode = "nft"
	IptablesModeLegacy IptablesMode = "legacy"
)

// Regexp used to parse the result of `iptables --version` then used to map to
// with IptablesMode
var IptablesModeRegex = regexp.MustCompile(`(?m)^ip6?tables.*?\((.*?)\)`)

// Map IptablesMode to the mode taken from the result of `iptables --version`
var IptablesModeMap = map[IptablesMode]string{
	IptablesModeLegacy: "legacy",    // i.e. iptables v1.8.5 (legacy)
	IptablesModeNft:    "nf_tables", // i.e. iptables v1.8.9 (nf_tables)
}

// FallbackExecutablesSearchLocations is a list of directories to search for
// the iptables executables if it cannot be found in the user's PATH environment
// variable. This allows for some flexibility in environments where iptables
// may be installed in non-standard locations.
var FallbackExecutablesSearchLocations = []string{
	"/usr/sbin",
	"/sbin",
	"/usr/bin",
	"/bin",
}

// Debug log level for iptables LOG jump target
// ref. https://git.netfilter.org/iptables/tree/extensions/libebt_log.c#n27
const LogLevelDebug uint16 = 7
