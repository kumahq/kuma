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
)

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
	},
}
