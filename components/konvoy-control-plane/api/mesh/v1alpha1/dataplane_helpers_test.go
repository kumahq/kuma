package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

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
					WorkloadAddress: "1.2.3.4",
					WorkloadPort:    8080,
					ServicePort:     80,
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
					WorkloadAddress: "1.2.3.4",
					WorkloadPort:    8080,
					ServicePort:     80,
				},
			}),
		)
	})

	Context("invalid input values", func() {
		type testCase struct {
			input       string
			expectedErr string
		}

		DescribeTable("should fail on invalid input values",
			func(given testCase) {
				// when
				iface, err := ParseInboundInterface(given.input)
				// then
				Expect(err.Error()).To(Equal(given.expectedErr))
				// and
				Expect(iface).To(BeZero())
			},
			Entry("dataplane IP address is missing", testCase{
				input:       ":80:8080",
				expectedErr: `invalid format: expected ^(?P<workload_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<service_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$, got ":80:8080"`,
			}),
			Entry("dataplane IP address is not valid", testCase{
				input:       "localhost:80:65536",
				expectedErr: `invalid format: expected ^(?P<workload_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<service_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$, got "localhost:80:65536"`,
			}),
			Entry("service port is missing", testCase{
				input:       "1.2.3.4::8080",
				expectedErr: `invalid format: expected ^(?P<workload_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<service_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$, got "1.2.3.4::8080"`,
			}),
			Entry("service port is out of range", testCase{
				input:       "1.2.3.4:0:8080",
				expectedErr: `invalid <SERVICE_PORT> in "1.2.3.4:0:8080": port number must be in the range [1, 65535] but got 0`,
			}),
			Entry("application port is missing", testCase{
				input:       "1.2.3.4:80:",
				expectedErr: `invalid format: expected ^(?P<workload_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<service_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$, got "1.2.3.4:80:"`,
			}),
			Entry("application port is out of range", testCase{
				input:       "1.2.3.4:80:65536",
				expectedErr: `invalid <WORKLOAD_PORT> in "1.2.3.4:80:65536": port number must be in the range [1, 65535] but got 65536`,
			}),
		)
	})
})
