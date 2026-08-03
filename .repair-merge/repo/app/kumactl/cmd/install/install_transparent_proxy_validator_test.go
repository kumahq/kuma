package install

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("shouldSkipValidation",
	func(ipv6, hasLocalIPv6Addr, validateOnlyIPv4, expected bool) {
		Expect(shouldSkipValidation(ipv6, hasLocalIPv6Addr, validateOnlyIPv4)).To(Equal(expected))
	},
	// IPv4 validation (ipv6=false) must never be skipped regardless of mode or IPv6 availability
	Entry("ipv4 call, dualstack mode, no IPv6 addr", false, false, false, false),
	Entry("ipv4 call, ipv4-only mode, no IPv6 addr", false, false, true, false),
	Entry("ipv4 call, dualstack mode, has IPv6 addr", false, true, false, false),
	Entry("ipv4 call, ipv4-only mode, has IPv6 addr", false, true, true, false),
	// IPv6 validation (ipv6=true) is skipped when no local IPv6 address or ipv4-only mode
	Entry("ipv6 call, dualstack mode, no IPv6 addr", true, false, false, true),
	Entry("ipv6 call, ipv4-only mode, no IPv6 addr", true, false, true, true),
	Entry("ipv6 call, ipv4-only mode, has IPv6 addr", true, true, true, true),
	// IPv6 validation runs only in dualstack mode with a local IPv6 address present
	Entry("ipv6 call, dualstack mode, has IPv6 addr", true, true, false, false),
)
