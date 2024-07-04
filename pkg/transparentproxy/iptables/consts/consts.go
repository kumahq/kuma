package consts

import (
	"regexp"
)

const (
<<<<<<< HEAD
=======
	Iptables  = "iptables"
	Ip6tables = "ip6tables"
)

// IptablesCommandByFamily maps a boolean value indicating IPv4 (false) or IPv6
// (true) usage to the corresponding iptables command name. This allows for code
// to be written generically without duplicating logic for both IPv4 and IPv6.
var IptablesCommandByFamily = map[bool]string{
	false: Iptables,
	true:  Ip6tables,
}

const (
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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

<<<<<<< HEAD
var Flags = map[string]map[bool]string{
	// commands
	"append": {
		Long:  "--append",
		Short: "-A",
	},
	"new-chain": {
		Long:  "--new-chain",
		Short: "-N",
	},

	// parameters
	"jump": {
		Long:  "--jump",
		Short: "-j",
=======
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

	// iptables-restore
	FlagWait         = "--wait"
	FlagWaitInterval = "--wait-interval"
	FlagNoFlush      = "--noflush"
)

// FlagVariationsMap maps a flag name (e.g., "-t") to a map containing its long
// (true) and short (false) flag representations. This allows for code to easily
// look up the appropriate flag based on desired usage (short or long).
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
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
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
var IptablesModeRegex = regexp.MustCompile(`(?m)^ip6?tables(?:.*?\((.*?)\))?`)

// Map IptablesMode to the mode taken from the result of `iptables --version`
var IptablesModeMap = map[IptablesMode][]string{
	IptablesModeLegacy: {
		"legacy", // i.e. iptables v1.8.5 (legacy)
		"",       // i.e. iptables v1.6.1
	},
	IptablesModeNft: {"nf_tables"}, // i.e. iptables v1.8.9 (nf_tables)
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
