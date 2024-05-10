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
	FlagTable = "table"

	// commands
	FlagAppend   = "append"
	FlagInsert   = "insert"
	FlagCheck    = "check"
	FlagNewChain = "new-chain"

	// parameters
	FlagJump = "jump"
)

var Flags = map[string]map[bool]string{
	FlagTable: {
		Long:  "--table",
		Short: "-t",
	},

	// commands
	FlagAppend: {
		Long:  "--append",
		Short: "-A",
	},
	FlagInsert: {
		Long:  "--insert",
		Short: "-I",
	},
	FlagCheck: {
		Long:  "--check",
		Short: "-C",
	},
	FlagNewChain: {
		Long:  "--new-chain",
		Short: "-N",
	},

	// parameters
	FlagJump: {
		Long:  "--jump",
		Short: "-j",
	},
}
