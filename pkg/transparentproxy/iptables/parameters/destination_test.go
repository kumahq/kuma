package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var _ = Describe("DestinationParameter", func() {
	Describe("Destination", func() {
		Describe("should build valid destination parameter with, provided address", func() {
			DescribeTable("when not negated",
				func(address, want, wantVerbose string) {
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
					"-d 254.254.254.254",
					"--destination 254.254.254.254",
				),
				Entry("IPv4 IP address with CIDR mask", "127.0.0.1/32",
					"-d 127.0.0.1/32",
					"--destination 127.0.0.1/32",
				),
				Entry("IPv6 IP address", "::1",
					"-d ::1",
					"--destination ::1",
				),
				Entry("IPv6 IP address with CIDR mask", "::1/128",
					"-d ::1/128",
					"--destination ::1/128",
				),
			)

			DescribeTable("when negated",
				func(address, want, wantVerbose string) {
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
					"! -d 254.254.254.254",
					"! --destination 254.254.254.254",
				),
				Entry("IPv4 IP address with CIDR mask", "127.0.0.1/32",
					"! -d 127.0.0.1/32",
					"! --destination 127.0.0.1/32",
				),
				Entry("IPv6 IP address", "::1",
					"! -d ::1",
					"! --destination ::1",
				),
				Entry("IPv6 IP address with CIDR mask", "::1/128",
					"! -d ::1/128",
					"! --destination ::1/128",
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
})
