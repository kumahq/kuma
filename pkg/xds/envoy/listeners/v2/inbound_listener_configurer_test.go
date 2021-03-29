package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("InboundListenerConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy.APIV2).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
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
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                ipv4Compat: true
                portValue: 8080
`,
		}),
		Entry("basic UDP listener", testCase{
			listenerName:     "inbound:192.168.0.1:8080",
			listenerProtocol: xds.SocketAddressProtocolUDP,
			listenerAddress:  "192.168.0.1",
			listenerPort:     8080,
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                protocol: UDP
                address: 192.168.0.1
                ipv4Compat: true
                portValue: 8080
`,
		}),
	)
})
