package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var _ = Describe("EndpointParameter", func() {
	Describe("Destination", func() {
		Describe("should build valid destination parameter with, provided address", func() {
			DescribeTable("when not negated",
				func(address string, want, wantVerbose []string) {
					// when
					got := Destination(address).Build(false)

					// then
					Expect(got).To(Equal(want))

					// and, when (verbose)
					got = Destination(address).Build(true)

					// then
					Expect(got).To(Equal(wantVerbose))
				},
				Entry("IPv4 IP address", "254.254.254.254",
					[]string{"-d", "254.254.254.254"},
					[]string{"--destination", "254.254.254.254"},
				),
				Entry("IPv4 IP address with CIDR mask", "127.0.0.1/32",
					[]string{"-d", "127.0.0.1/32"},
					[]string{"--destination", "127.0.0.1/32"},
				),
				Entry("IPv6 IP address", "::1",
					[]string{"-d", "::1"},
					[]string{"--destination", "::1"},
				),
				Entry("IPv6 IP address with CIDR mask", "::1/128",
					[]string{"-d", "::1/128"},
					[]string{"--destination", "::1/128"},
				),
			)

			DescribeTable("when negated",
				func(address string, want, wantVerbose []string) {
					// when
					got := Destination(address).Negate().Build(false)

					// then
					Expect(got).To(Equal(want))

					// and, when (verbose)
					got = Destination(address).Negate().Build(true)

					// then
					Expect(got).To(Equal(wantVerbose))
				},
				Entry("IPv4 IP address", "254.254.254.254",
					[]string{"!", "-d", "254.254.254.254"},
					[]string{"!", "--destination", "254.254.254.254"},
				),
				Entry("IPv4 IP address with CIDR mask", "127.0.0.1/32",
					[]string{"!", "-d", "127.0.0.1/32"},
					[]string{"!", "--destination", "127.0.0.1/32"},
				),
				Entry("IPv6 IP address", "::1",
					[]string{"!", "-d", "::1"},
					[]string{"!", "--destination", "::1"},
				),
				Entry("IPv6 IP address with CIDR mask", "::1/128",
					[]string{"!", "-d", "::1/128"},
					[]string{"!", "--destination", "::1/128"},
				),
			)
		})
	})

	Describe("NotDestination", func() {
		DescribeTable("should return the result of Destination(...).Negate()",
			func(address string) {
				// given
				want := Destination(address).Negate()

				// when
				got := NotDestination(address).Build(false)

				// then
				Expect(got).To(BeEquivalentTo(want.Build(false)))

				// and, when (verbose)
				got = NotDestination(address).Build(true)

				// then
				Expect(got).To(BeEquivalentTo(want.Build(true)))
			},
			Entry("IPv4 IP address", "254.254.254.254"),
			Entry("IPv4 IP address with CIDR mask", "127.0.0.1/32"),
			Entry("IPv6 IP address", "::1"),
			Entry("IPv6 IP address with CIDR mask", "::1/128"),
		)
	})

	Describe("Source", func() {
		DescribeTable("should build valid source parameter with the built, provided Source",
			func(address string, verbose bool, want []string) {
				// when
				got := Source(address).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("127.0.0.1/32",
				"127.0.0.1/32", false,
				[]string{"-s", "127.0.0.1/32"},
			),
			Entry("127.0.0.1/32 - verbose",
				"127.0.0.1/32", true,
				[]string{"--source", "127.0.0.1/32"},
			),
			Entry("254.254.254.254",
				"254.254.254.254", false,
				[]string{"-s", "254.254.254.254"},
			),
			Entry("254.254.254.254 - verbose",
				"254.254.254.254", true,
				[]string{"--source", "254.254.254.254"},
			),
		)

		DescribeTable("should build valid source parameter with the built, provided address when negated",
			func(address string, verbose bool, want []string) {
				// when
				got := Source(address).Negate().Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("127.0.0.1/32",
				"127.0.0.1/32", false,
				[]string{"!", "-s", "127.0.0.1/32"},
			),
			Entry("127.0.0.1/32 - verbose",
				"127.0.0.1/32", true,
				[]string{"!", "--source", "127.0.0.1/32"},
			),
			Entry("254.254.254.254",
				"254.254.254.254", false,
				[]string{"!", "-s", "254.254.254.254"},
			),
			Entry("254.254.254.254 - verbose",
				"254.254.254.254", true,
				[]string{"!", "--source", "254.254.254.254"},
			),
		)
	})
})
