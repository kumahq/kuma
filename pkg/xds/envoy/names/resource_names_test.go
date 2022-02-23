package names_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ = Describe("ListenerFilterChainConfigurer", func() {

	type testCase struct {
		address  string
		port     uint32
		expected string
	}

	DescribeTable("GetInboundListenerName",
		func(given testCase) {
			// when
			name := names.GetInboundListenerName(given.address, given.port)
			// then
			Expect(name).To(Equal(given.expected))
		},
		Entry("IPv4 address", testCase{
			address:  "192.168.0.1",
			port:     8080,
			expected: "inbound:192.168.0.1:8080",
		}),
		Entry("IPv6 address", testCase{
			address:  "fd00::1",
			port:     8080,
			expected: "inbound:[fd00::1]:8080",
		}),
	)

	DescribeTable("GetOutboundListenerName",
		func(given testCase) {
			// when
			name := names.GetOutboundListenerName(given.address, given.port)
			// then
			Expect(name).To(Equal(given.expected))
		},
		Entry("IPv4 address", testCase{
			address:  "192.168.0.1",
			port:     8080,
			expected: "outbound:192.168.0.1:8080",
		}),
		Entry("IPv6 address", testCase{
			address:  "fd00::1",
			port:     8080,
			expected: "outbound:[fd00::1]:8080",
		}),
	)

})
