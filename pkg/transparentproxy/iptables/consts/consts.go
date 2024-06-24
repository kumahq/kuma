package consts

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

// Debug log level for iptables LOG jump target
// ref. https://git.netfilter.org/iptables/tree/extensions/libebt_log.c#n27
const LogLevelDebug uint16 = 7
