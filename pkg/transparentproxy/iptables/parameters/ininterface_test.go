package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var _ = Describe("InInterfaceParameter", func() {
	DescribeTable("should return ",
		func(iface string, verbose bool, want string) {
			// when
			got := InInterface(iface).Build(verbose)

			// then
			Expect(got).To(Equal(want))
		},
		Entry("interface localhost",
			"localhost", false,
			"-i localhost",
		),
		Entry("interface localhost - verbose",
			"localhost", true,
			"--in-interface localhost",
		),
	)
	DescribeTable("should return negation ",
		func(iface string, verbose bool, want string) {
			// when
			got := InInterface(iface).Negate().Build(verbose)

			// then
			Expect(got).To(Equal(want))
		},
		Entry("interface localhost",
			"localhost", false,
			"! -i localhost",
		),
		Entry("interface localhost - verbose",
			"localhost", true,
			"! --in-interface localhost",
		),
	)
})
