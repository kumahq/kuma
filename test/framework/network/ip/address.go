package ip

import (
	"math/rand"
	"net"

	"github.com/onsi/ginkgo/v2"
)

// #nosec G404 -- used just for tests
var r = rand.New(rand.NewSource(ginkgo.GinkgoRandomSeed()))

// ref. https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
// ref. https://en.wikipedia.org/wiki/Reserved_IP_addresses#IPv4
var reservedIPv4 = []*net.IPNet{
	// This network
	{IP: net.IPv4(0, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
	// This host on this network
	{IP: net.IPv4(0, 0, 0, 0), Mask: net.CIDRMask(32, 32)},
	// Private-Use
	{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
	// Shared Address Space
	{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)},
	// Loopback
	{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
	// Link Local
	{IP: net.IPv4(169, 254, 0, 0), Mask: net.CIDRMask(16, 32)},
	// Private-Use
	{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
	// IETF Protocol Assignments
	{IP: net.IPv4(192, 0, 0, 0), Mask: net.CIDRMask(24, 32)},
	// IPv4 Service Continuity Prefix
	{IP: net.IPv4(192, 0, 0, 0), Mask: net.CIDRMask(29, 32)},
	// IPv4 dummy address
	{IP: net.IPv4(192, 0, 0, 8), Mask: net.CIDRMask(32, 32)},
	// Port Control Protocol Anycast
	{IP: net.IPv4(192, 0, 0, 9), Mask: net.CIDRMask(32, 32)},
	// Traversal Using Relays around NAT Anycast
	{IP: net.IPv4(192, 0, 0, 10), Mask: net.CIDRMask(32, 32)},
	// NAT64/DNS64 Discovery
	{IP: net.IPv4(192, 0, 0, 170), Mask: net.CIDRMask(32, 32)},
	// NAT64/DNS64 Discovery
	{IP: net.IPv4(192, 0, 0, 171), Mask: net.CIDRMask(32, 32)},
	// Documentation (TEST-NET-1)
	{IP: net.IPv4(192, 0, 2, 0), Mask: net.CIDRMask(24, 32)},
	// AS112-v4
	{IP: net.IPv4(192, 31, 196, 0), Mask: net.CIDRMask(24, 32)},
	// AMT
	{IP: net.IPv4(192, 52, 193, 0), Mask: net.CIDRMask(24, 32)},
	// Deprecated (6to4 Relay Anycast)
	{IP: net.IPv4(192, 88, 99, 0), Mask: net.CIDRMask(24, 32)},
	// Direct Delegation AS112 Service
	{IP: net.IPv4(192, 175, 48, 0), Mask: net.CIDRMask(24, 32)},
	// Benchmarking
	{IP: net.IPv4(198, 18, 0, 0), Mask: net.CIDRMask(15, 32)},
	// Documentation (TEST-NET-2)
	{IP: net.IPv4(198, 51, 100, 0), Mask: net.CIDRMask(24, 32)},
	// Documentation (TEST-NET-3)
	{IP: net.IPv4(203, 0, 113, 0), Mask: net.CIDRMask(24, 32)},
	// IP multicast (Former Class D network)
	{IP: net.IPv4(224, 0, 0, 0), Mask: net.CIDRMask(4, 32)},
	// Reserved
	{IP: net.IPv4(240, 0, 0, 0), Mask: net.CIDRMask(4, 32)},
	// Limited Broadcast
	{IP: net.ParseIP("255.255.255.255"), Mask: net.CIDRMask(32, 32)},
}

// ref. https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
// ref. https://en.wikipedia.org/wiki/Reserved_IP_addresses#IPv6
var reservedIPv6 = []*net.IPNet{
	// Loopback Address
	{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},
	// Unspecified Address
	{IP: net.ParseIP("::"), Mask: net.CIDRMask(128, 128)},
	// IPv4-mapped Address
	{IP: net.ParseIP("::ffff:0:0"), Mask: net.CIDRMask(96, 128)},
	// IPv4-IPv6 Translation
	{IP: net.ParseIP("64:ff9b::"), Mask: net.CIDRMask(96, 128)},
	// IPv4-IPv6 Translation
	{IP: net.ParseIP("64:ff9b:1::"), Mask: net.CIDRMask(48, 128)},
	// Discard-Only Address Block
	{IP: net.ParseIP("100::"), Mask: net.CIDRMask(64, 128)},
	// IETF Protocol Assignments
	{IP: net.ParseIP("2001::"), Mask: net.CIDRMask(23, 128)},
	// TEREDO
	{IP: net.ParseIP("2001::"), Mask: net.CIDRMask(32, 128)},
	// Port Control Protocol Anycast
	{IP: net.ParseIP("2001:1::1"), Mask: net.CIDRMask(128, 128)},
	// Traversal Using Relays around NAT Anycast
	{IP: net.ParseIP("2001:1::2"), Mask: net.CIDRMask(128, 128)},
	// Benchmarking
	{IP: net.ParseIP("2001:2::"), Mask: net.CIDRMask(48, 128)},
	// AMT
	{IP: net.ParseIP("2001:3::"), Mask: net.CIDRMask(32, 128)},
	// AS112-v6
	{IP: net.ParseIP("2001:4:112::"), Mask: net.CIDRMask(48, 128)},
	// Deprecated (previously ORCHID)
	{IP: net.ParseIP("2001:10::"), Mask: net.CIDRMask(28, 128)},
	// ORCHIDv2
	{IP: net.ParseIP("2001:20::"), Mask: net.CIDRMask(28, 128)},
	// Documentation
	{IP: net.ParseIP("2001:db8::"), Mask: net.CIDRMask(32, 128)},
	// 6to4
	{IP: net.ParseIP("2002::"), Mask: net.CIDRMask(16, 128)},
	// Direct Delegation AS112 Service
	{IP: net.ParseIP("2620:4f:8000::"), Mask: net.CIDRMask(48, 128)},
	// Unique-Local
	{IP: net.ParseIP("fc00::"), Mask: net.CIDRMask(7, 128)},
	// Link-Local Unicast
	{IP: net.ParseIP("fe80::"), Mask: net.CIDRMask(10, 128)},
	// Multicast address
	{IP: net.ParseIP("ff00::"), Mask: net.CIDRMask(8, 128)},
}

func genRandomIP(size int, reserved []*net.IPNet) net.IP {
	ipBytes := make([]byte, size)

	for i := 0; i < size; i++ {
		ipBytes[i] = byte(r.Intn(256))
	}

	ip := net.IP(ipBytes)

	for _, ipNet := range reserved {
		if ipNet.Contains(ip) {
			return genRandomIP(size, reserved)
		}
	}

	return ip
}

// GenRandomIPv4 will return random, non-reserved IPv4 address
func GenRandomIPv4() net.IP {
	return genRandomIP(4, reservedIPv4).To4()
}

// GenRandomIPv6 will return random, non-reserved IPv6 address
func GenRandomIPv6() net.IP {
	return genRandomIP(16, reservedIPv6).To16()
}
