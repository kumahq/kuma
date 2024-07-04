package consts

import (
	"regexp"
)

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

// Regexp used to parse the result of `iptables --version` then used to map to
// with IptablesMode
var IptablesModeRegex = regexp.MustCompile(`(?m)^ip6?tables(?:.*?\((.*?)\))?`)

// Map IptablesMode to the mode taken from the result of `iptables --version`
var IptablesModeMap = map[string][]string{
	"legacy": {
		"legacy", // i.e. iptables v1.8.5 (legacy)
		"",       // i.e. iptables v1.6.1
	},
	"nft": {"nf_tables"}, // i.e. iptables v1.8.9 (nf_tables)
}
