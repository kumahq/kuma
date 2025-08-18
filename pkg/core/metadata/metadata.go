package metadata

import (
	"net"
)

// Service that indicates L4 pass through cluster
const PassThroughServiceName = "pass_through"

var (
	LoopbackIPv4 = net.IPv4(127, 0, 0, 1)
	LoopbackIPv6 = net.IPv6loopback
)
