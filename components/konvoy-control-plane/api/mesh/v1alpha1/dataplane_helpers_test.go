package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	. "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
)

var _ = Describe("InboundInterface", func() {

	Describe("String()", func() {
		type testCase struct {
			iface    InboundInterface
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				Expect(given.iface.String()).To(Equal(given.expected))
			},
			Entry("all fields set", testCase{
				iface: InboundInterface{
					DataplaneIP:   "1.2.3.4",
					DataplanePort: 80,
					WorkloadPort:  8080,
				},
				expected: "1.2.3.4:80:8080",
			}),
		)
	})
})

var _ = Describe("ParseInboundInterface(..)", func() {

	Context("valid input values", func() {
		type testCase struct {
			input    string
			expected InboundInterface
		}

		DescribeTable("should parse valid input values",
			func(given testCase) {
				// when
				iface, err := ParseInboundInterface(given.input)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(iface).To(Equal(given.expected))
			},
			Entry("all fields set", testCase{
				input: "1.2.3.4:80:8080",
				expected: InboundInterface{
					DataplaneIP:   "1.2.3.4",
					DataplanePort: 80,
					WorkloadPort:  8080,
				},
			}),
		)
	})

	Context("invalid input values", func() {
		type testCase struct {
			input       string
			expectedErr gomega_types.GomegaMatcher
		}

		DescribeTable("should fail on invalid input values",
			func(given testCase) {
				// when
				iface, err := ParseInboundInterface(given.input)
				// then
				Expect(err.Error()).To(given.expectedErr)
				// and
				Expect(iface).To(BeZero())
			},
			Entry("dataplane IP address is missing", testCase{
				input:       ":80:8080",
				expectedErr: MatchRegexp(`invalid format: expected .*, got ":80:8080"`),
			}),
			Entry("dataplane IP address is not valid", testCase{
				input:       "localhost:80:65536",
				expectedErr: MatchRegexp(`invalid format: expected .*, got "localhost:80:65536"`),
			}),
			Entry("service port is missing", testCase{
				input:       "1.2.3.4::8080",
				expectedErr: MatchRegexp(`invalid format: expected .*, got "1.2.3.4::8080"`),
			}),
			Entry("service port is out of range", testCase{
				input:       "1.2.3.4:0:8080",
				expectedErr: Equal(`invalid <DATAPLANE_PORT> in "1.2.3.4:0:8080": port number must be in the range [1, 65535] but got 0`),
			}),
			Entry("application port is missing", testCase{
				input:       "1.2.3.4:80:",
				expectedErr: MatchRegexp(`invalid format: expected .*, got "1.2.3.4:80:"`),
			}),
			Entry("application port is out of range", testCase{
				input:       "1.2.3.4:80:65536",
				expectedErr: Equal(`invalid <WORKLOAD_PORT> in "1.2.3.4:80:65536": port number must be in the range [1, 65535] but got 65536`),
			}),
		)
	})
})

var _ = Describe("Dataplane_Networking", func() {

	Describe("GetInboundInterfaces()", func() {

		Context("valid input values", func() {
			type testCase struct {
				input    *Dataplane_Networking
				expected []InboundInterface
			}

			DescribeTable("should parse valid input values",
				func(given testCase) {
					// when
					ifaces, err := given.input.GetInboundInterfaces()
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(ifaces).To(ConsistOf(given.expected))
				},
				Entry("nil", testCase{
					input:    nil,
					expected: nil,
				}),
				Entry("empty", testCase{
					input:    &Dataplane_Networking{},
					expected: []InboundInterface{},
				}),
				Entry("2 inbound interfaces", testCase{
					input: &Dataplane_Networking{
						Inbound: []*Dataplane_Networking_Inbound{
							{Interface: "192.168.0.1:80:8080"},
							{Interface: "192.168.0.1:443:8443"},
						},
					},
					expected: []InboundInterface{
						{DataplaneIP: "192.168.0.1", DataplanePort: 80, WorkloadPort: 8080},
						{DataplaneIP: "192.168.0.1", DataplanePort: 443, WorkloadPort: 8443},
					},
				}),
			)
		})

		Context("invalid input values", func() {
			type testCase struct {
				input       *Dataplane_Networking
				expectedErr gomega_types.GomegaMatcher
			}

			DescribeTable("should fail on invalid input values",
				func(given testCase) {
					// when
					ifaces, err := given.input.GetInboundInterfaces()
					// then
					Expect(ifaces).To(BeNil())
					// and
					Expect(err.Error()).To(given.expectedErr)
				},
				Entry("dataplane IP address is missing", testCase{
					input: &Dataplane_Networking{
						Inbound: []*Dataplane_Networking_Inbound{
							{Interface: "192.168.0.1:80:8080"},
							{Interface: ":443:8443"},
						},
					},
					expectedErr: MatchRegexp(`invalid format: expected .*, got ":443:8443"`),
				}),
			)
		})
	})
})
