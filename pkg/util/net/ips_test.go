package net_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/net"
)

var _ = DescribeTable("ToV6",
	func(given string, expected string) {
		Expect(net.ToV6(given)).To(Equal(expected))
	},
	Entry("v6 already", "2001:db8::ff00:42:8329", "2001:db8::ff00:42:8329"),
	Entry("v6 not compacted", "2001:0db8:0000:0000:0000:ff00:0042:8329", "2001:0db8:0000:0000:0000:ff00:0042:8329"),
	Entry("v4 adds prefix", "240.0.0.0", "::ffff:f000:0"),
	Entry("v4 adds prefix", "240.0.255.0", "::ffff:f000:ff00"),
)
