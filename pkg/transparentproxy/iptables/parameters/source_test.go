package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var _ = Describe("SourceParameter", func() {
	DescribeTable("Address",
		func(address string, verbose bool, want string) {
			// when
			got := Address(address).Build(verbose)

			// then
			Expect(got).To(Equal(want))
		},
		Entry("CIDR IPv4 localhost",
			"127.0.0.1/32", false,
			"127.0.0.1/32",
		),
		Entry("CIDR IPv4 localhost - verbose",
			"127.0.0.1/32", true,
			"127.0.0.1/32",
		),
		Entry("IPv4 address (no CIDR) - localhost",
			"127.0.0.1", false,
			"127.0.0.1",
		),
		Entry("IPv4 address (no CIDR) - localhost - verbose",
			"127.0.0.1", true,
			"127.0.0.1",
		),
		Entry("CIDR IPv4 address",
			"254.254.254.254/32", false,
			"254.254.254.254/32",
		),
		Entry("CIDR IPv4 address - verbose",
			"254.254.254.254/32", true,
			"254.254.254.254/32",
		),
		Entry("IPv4 address (no CIDR)",
			"254.254.254.254", false,
			"254.254.254.254",
		),
		Entry("IPv4 address (no CIDR) - verbose",
			"254.254.254.254", true,
			"254.254.254.254",
		),
	)
	DescribeTable("negated Address should return the same values as not negated one, as it "+
		"shouldn't be possible to negate address (value shouldn't change)",
		func(address string, verbose bool, want string) {
			// when
			got := Address(address).Negate().Build(verbose)

			// then
			Expect(got).To(Equal(want))
		},
		Entry("CIDR IPv4 localhost - negated",
			"127.0.0.1/32", false,
			"127.0.0.1/32",
		),
		Entry("CIDR IPv4 localhost - verbose - negated",
			"127.0.0.1/32", true,
			"127.0.0.1/32",
		),
		Entry("IPv4 address (no CIDR) - localhost - negated",
			"127.0.0.1", false,
			"127.0.0.1",
		),
		Entry("IPv4 address (no CIDR) - localhost - verbose - negated",
			"127.0.0.1", true,
			"127.0.0.1",
		),
		Entry("CIDR IPv4 address - negated",
			"254.254.254.254/32", false,
			"254.254.254.254/32",
		),
		Entry("CIDR IPv4 address - verbose - negated",
			"254.254.254.254/32", true,
			"254.254.254.254/32",
		),
		Entry("IPv4 address (no CIDR) - negated",
			"254.254.254.254", false,
			"254.254.254.254",
		),
		Entry("IPv4 address (no CIDR) - verbose - negated",
			"254.254.254.254", true,
			"254.254.254.254",
		),
	)

	Describe("Source", func() {
		DescribeTable("should build valid source parameter with the built, provided "+
			"*SourceParameter",
			func(parameter *SourceParameter, verbose bool, want string) {
				// when
				got := Source(parameter).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("Address('127.0.0.1/32')",
				Address("127.0.0.1/32"), false,
				"-s 127.0.0.1/32",
			),
			Entry("Address('127.0.0.1/32') - verbose",
				Address("127.0.0.1/32"), true,
				"--source 127.0.0.1/32",
			),
			Entry("Address('254.254.254.254')",
				Address("254.254.254.254"), false,
				"-s 254.254.254.254",
			),
			Entry("Address('254.254.254.254') - verbose",
				Address("254.254.254.254"), true,
				"--source 254.254.254.254",
			),
		)

		DescribeTable("should build valid source parameter with the built, provided "+
			"*SourceParameter when negated",
			func(parameter *SourceParameter, verbose bool, want string) {
				// when
				got := Source(parameter).Negate().Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("Address('127.0.0.1/32')",
				Address("127.0.0.1/32"), false,
				"! -s 127.0.0.1/32",
			),
			Entry("Address('127.0.0.1/32') - verbose",
				Address("127.0.0.1/32"), true,
				"! --source 127.0.0.1/32",
			),
			Entry("Address('254.254.254.254')",
				Address("254.254.254.254"), false,
				"! -s 254.254.254.254",
			),
			Entry("Address('254.254.254.254') - verbose",
				Address("254.254.254.254"), true,
				"! --source 254.254.254.254",
			),
		)
	})
})
