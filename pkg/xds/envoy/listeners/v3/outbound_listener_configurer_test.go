package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("OutboundListenerConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerAddress  string
		listenerPort     uint32
		listenerProtocol core_xds.SocketAddressProtocol
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic listener", testCase{
			listenerName:    "outbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			expected: `
            name: outbound:192.168.0.1:8080
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
`,
		}),
		Entry("basic UDP listener", testCase{
			listenerName:     "outbound:192.168.0.1:8080",
			listenerAddress:  "192.168.0.1",
			listenerPort:     8080,
			listenerProtocol: core_xds.SocketAddressProtocolUDP,
			expected: `
            name: outbound:192.168.0.1:8080
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                protocol: UDP
                address: 192.168.0.1
                portValue: 8080
`,
		}),
	)
})
