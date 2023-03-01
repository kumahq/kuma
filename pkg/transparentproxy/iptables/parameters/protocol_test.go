package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

func returnTrue() bool {
	return true
}

func returnFalse() bool {
	return false
}

var _ = Describe("ProtocolParameter", func() {
	Describe("Protocol", func() {
		DescribeTable("DestinationPort",
			func(port int, verbose bool, want string) {
				// when
				got := DestinationPort(uint16(port)).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("port 22",
				22, false,
				"--dport 22",
			),
			Entry("port 22 - verbose",
				22, true,
				"--destination-port 22",
			),
			Entry("port 7777",
				7777, false,
				"--dport 7777",
			),
			Entry("port 7777 - verbose",
				7777, true,
				"--destination-port 7777",
			),
		)

		DescribeTable("NotDestinationPort",
			func(port int, verbose bool, want string) {
				// when
				got := NotDestinationPort(uint16(port)).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("port 22",
				22, false,
				"! --dport 22",
			),
			Entry("port 22 - verbose",
				22, true,
				"! --destination-port 22",
			),
			Entry("port 7777",
				7777, false,
				"! --dport 7777",
			),
			Entry("port 7777 - verbose",
				7777, true,
				"! --destination-port 7777",
			),
		)

		Describe("NotDestinationPortIf", func() {
			DescribeTable("should return nil, when predicate returns false",
				func(port int) {
					Expect(NotDestinationPortIf(returnFalse, uint16(port))).To(BeNil())
				},
				Entry("port 22", 22),
				Entry("port 80", 80),
				Entry("port 8080", 8080),
				Entry("port 7777", 7777),
			)

			DescribeTable("should build a valid flag, when predicate returns true",
				func(port int, verbose bool, want string) {
					// when
					got := NotDestinationPortIf(returnTrue, uint16(port)).Build(verbose)

					// then
					Expect(got).To(Equal(want))
				},
				Entry("port 22",
					22, false,
					"! --dport 22",
				),
				Entry("port 22 - verbose",
					22, true,
					"! --destination-port 22",
				),
				Entry("port 80",
					80, false,
					"! --dport 80",
				),
				Entry("port 80 - verbose",
					80, true,
					"! --destination-port 80",
				),
				Entry("port 8080",
					8080, false,
					"! --dport 8080",
				),
				Entry("port 8080 - verbose",
					8080, true,
					"! --destination-port 8080",
				),
				Entry("port 7777",
					7777, false,
					"! --dport 7777",
				),
				Entry("port 7777 - verbose",
					7777, true,
					"! --destination-port 7777",
				),
			)
		})

		DescribeTable("Tcp",
			func(parameters []*TcpUdpParameter, verbose bool, want string) {
				// when
				got := Tcp(parameters...).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("no parameters",
				nil, false,
				"tcp",
			),
			Entry("no parameters - verbose",
				nil, true,
				"tcp",
			),
			Entry("1 parameter (DestinationPort(22))",
				[]*TcpUdpParameter{DestinationPort(22)}, false,
				"tcp --dport 22",
			),
			Entry("1 parameter (DestinationPort(22)) - verbose",
				[]*TcpUdpParameter{DestinationPort(22)}, true,
				"tcp --destination-port 22",
			),
			Entry("1 parameter (NotDestinationPort(22))",
				[]*TcpUdpParameter{NotDestinationPort(22)}, false,
				"tcp ! --dport 22",
			),
			Entry("1 parameter (NotDestinationPort(22)) - verbose",
				[]*TcpUdpParameter{NotDestinationPort(22)}, true,
				"tcp ! --destination-port 22",
			),
		)

		DescribeTable("Udp",
			func(parameters []*TcpUdpParameter, verbose bool, want string) {
				// when
				got := Udp(parameters...).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("no parameters",
				nil, false,
				"udp",
			),
			Entry("no parameters - verbose",
				nil, true,
				"udp",
			),
			Entry("1 parameter (DestinationPort(53))",
				[]*TcpUdpParameter{DestinationPort(53)}, false,
				"udp --dport 53",
			),
			Entry("1 parameter (DestinationPort(53)) - verbose",
				[]*TcpUdpParameter{DestinationPort(53)}, true,
				"udp --destination-port 53",
			),
			Entry("1 parameter (NotDestinationPort(53))",
				[]*TcpUdpParameter{NotDestinationPort(53)}, false,
				"udp ! --dport 53",
			),
			Entry("1 parameter (NotDestinationPort(53)) - verbose",
				[]*TcpUdpParameter{NotDestinationPort(53)}, true,
				"udp ! --destination-port 53",
			),
		)
	})
})
