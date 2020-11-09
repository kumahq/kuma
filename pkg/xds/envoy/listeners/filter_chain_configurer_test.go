package listeners_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ListenerFilterChainConfigurer", func() {

	type testCase struct {
		listenerName    string
		protocol        mesh_core.Protocol
		listenerAddress string
		listenerPort    uint32
		expected        string
	}

	Context("V2", func() {
		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// when
				listener, err := NewListenerBuilder(envoy.APIV2).
					Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
					Configure(FilterChain(NewFilterChainBuilder(envoy.APIV2))).
					Build()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := util_proto.ToYAML(listener)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("basic listener with an empty filter chain", testCase{
				listenerName:    "inbound:192.168.0.1:8080",
				listenerAddress: "192.168.0.1",
				listenerPort:    8080,
				expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - {}
`,
			}),
		)
	})

	Context("V3", func() {
		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// when
				listener, err := NewListenerBuilder(envoy.APIV3).
					Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
					Configure(FilterChain(NewFilterChainBuilder(envoy.APIV3))).
					Build()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := util_proto.ToYAML(listener)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("basic listener with an empty filter chain", testCase{
				listenerName:    "inbound:192.168.0.1:8080",
				listenerAddress: "192.168.0.1",
				listenerPort:    8080,
				expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - {}
`,
			}),
		)
	})
})
