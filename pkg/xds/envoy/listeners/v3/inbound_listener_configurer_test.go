package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
)

var _ = Describe("InboundListenerConfigurer", func() {
	type testCase struct {
		listenerProtocol  xds.SocketAddressProtocol
		listenerAddress   string
		listenerPort      uint32
		enableReusedPorts bool
		expected          string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewInboundListenerBuilder(envoy.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol, given.enableReusedPorts).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic TCP listener with reusable ports enabled", testCase{
			listenerProtocol:  xds.SocketAddressProtocolTCP,
			listenerAddress:   "192.168.0.1",
			listenerPort:      8080,
			enableReusedPorts: true,
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            enableReusePort: true
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
`,
		}),
		Entry("basic TCP listener with reusable ports disabled", testCase{
			listenerProtocol:  xds.SocketAddressProtocolTCP,
			listenerAddress:   "192.168.0.1",
			listenerPort:      8080,
			enableReusedPorts: false,
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            enableReusePort: false
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
`,
		}),
		Entry("basic UDP listener always enables reuse port", testCase{
			listenerProtocol:  xds.SocketAddressProtocolUDP,
			listenerAddress:   "192.168.0.1",
			listenerPort:      8080,
			enableReusedPorts: false,
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            enableReusePort: true
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
                protocol: UDP
`,
		}),
	)
})
