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
			func(port int, verbose bool, want []string) {
				// when
				got := DestinationPort(uint16(port)).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("port 22",
				22, false,
				[]string{"--dport", "22"},
			),
			Entry("port 22 - verbose",
				22, true,
				[]string{"--destination-port", "22"},
			),
			Entry("port 7777",
				7777, false,
				[]string{"--dport", "7777"},
			),
			Entry("port 7777 - verbose",
				7777, true,
				[]string{"--destination-port", "7777"},
			),
		)

		DescribeTable("NotDestinationPort",
			func(port int, verbose bool, want []string) {
				// when
				got := NotDestinationPort(uint16(port)).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("port 22",
				22, false,
				[]string{"!", "--dport", "22"},
			),
			Entry("port 22 - verbose",
				22, true,
				[]string{"!", "--destination-port", "22"},
			),
			Entry("port 7777",
				7777, false,
				[]string{"!", "--dport", "7777"},
			),
			Entry("port 7777 - verbose",
				7777, true,
				[]string{"!", "--destination-port", "7777"},
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
				func(port int, verbose bool, want []string) {
					// when
					got := NotDestinationPortIf(returnTrue, uint16(port)).Build(verbose)

					// then
					Expect(got).To(Equal(want))
				},
				Entry("port 22",
					22, false,
					[]string{"!", "--dport", "22"},
				),
				Entry("port 22 - verbose",
					22, true,
					[]string{"!", "--destination-port", "22"},
				),
				Entry("port 80",
					80, false,
					[]string{"!", "--dport", "80"},
				),
				Entry("port 80 - verbose",
					80, true,
					[]string{"!", "--destination-port", "80"},
				),
				Entry("port 8080",
					8080, false,
					[]string{"!", "--dport", "8080"},
				),
				Entry("port 8080 - verbose",
					8080, true,
					[]string{"!", "--destination-port", "8080"},
				),
				Entry("port 7777",
					7777, false,
					[]string{"!", "--dport", "7777"},
				),
				Entry("port 7777 - verbose",
					7777, true,
					[]string{"!", "--destination-port", "7777"},
				),
			)
		})

		DescribeTable("Tcp",
			func(parameters []*TcpUdpParameter, verbose bool, want []string) {
				// when
				got := Tcp(parameters...).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("no parameters",
				nil, false,
				[]string{"tcp"},
			),
			Entry("no parameters - verbose",
				nil, true,
				[]string{"tcp"},
			),
			Entry("1 parameter (DestinationPort(22))",
				[]*TcpUdpParameter{DestinationPort(uint16(22))}, false,
				[]string{"tcp", "--dport", "22"},
			),
			Entry("1 parameter (DestinationPort(22)) - verbose",
				[]*TcpUdpParameter{DestinationPort(uint16(22))}, true,
				[]string{"tcp", "--destination-port", "22"},
			),
			Entry("1 parameter (NotDestinationPort(22))",
				[]*TcpUdpParameter{NotDestinationPort(uint16(22))}, false,
				[]string{"tcp", "!", "--dport", "22"},
			),
			Entry("1 parameter (NotDestinationPort(22)) - verbose",
				[]*TcpUdpParameter{NotDestinationPort(uint16(22))}, true,
				[]string{"tcp", "!", "--destination-port", "22"},
			),
		)

		DescribeTable("Udp",
			func(parameters []*TcpUdpParameter, verbose bool, want []string) {
				// when
				got := Udp(parameters...).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("no parameters",
				nil, false,
				[]string{"udp"},
			),
			Entry("no parameters - verbose",
				nil, true,
				[]string{"udp"},
			),
			Entry("1 parameter (DestinationPort(53))",
				[]*TcpUdpParameter{DestinationPort(uint16(53))}, false,
				[]string{"udp", "--dport", "53"},
			),
			Entry("1 parameter (DestinationPort(53)) - verbose",
				[]*TcpUdpParameter{DestinationPort(uint16(53))}, true,
				[]string{"udp", "--destination-port", "53"},
			),
			Entry("1 parameter (NotDestinationPort(53))",
				[]*TcpUdpParameter{NotDestinationPort(uint16(53))}, false,
				[]string{"udp", "!", "--dport", "53"},
			),
			Entry("1 parameter (NotDestinationPort(53)) - verbose",
				[]*TcpUdpParameter{NotDestinationPort(uint16(53))}, true,
				[]string{"udp", "!", "--destination-port", "53"},
			),
		)
	})
})
