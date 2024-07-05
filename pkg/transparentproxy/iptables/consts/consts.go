package consts

const (
	Long  = true
	Short = false
)

<<<<<<< HEAD
=======
// IptablesCommandByFamily maps a boolean value indicating IPv4 (false) or IPv6
// (true) usage to the corresponding iptables command name. This allows for code
// to be written generically without duplicating logic for both IPv4 and IPv6.
var IptablesCommandByFamily = map[bool]string{
	false: Iptables,
	true:  Ip6tables,
}

// Default ports used for iptables redirection.
const (
	DefaultRedirectInbountPort     uint16 = 15006
	DefaultRedirectInbountPortIPv6 uint16 = 15010
	DefaultRedirectOutboundPort    uint16 = 15001
	DefaultRedirectDNSPort         uint16 = 15053
)

>>>>>>> ddef32cbe (refactor(transparent-proxy): put default ports in consts package (#10801))
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
