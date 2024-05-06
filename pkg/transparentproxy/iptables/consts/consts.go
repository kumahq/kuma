package consts

const (
	Long  = true
	Short = false
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

	// Table names
	TableNat    = "nat"
	TableRaw    = "raw"
	TableMangle = "mangle"
)

const (
	// commands
	FlagAppend   = "append"
	FlagInsert   = "insert"
	FlagNewChain = "new-chain"

	// parameters
	FlagJump = "jump"
)

var Flags = map[string]map[bool]string{
	// commands
	FlagAppend: {
		Long:  "--append",
		Short: "-A",
	},
	FlagInsert: {
		Long:  "--insert",
		Short: "-I",
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
