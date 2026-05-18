package listeners_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
)

var _ = Describe("ListenerFilterChainConfigurer", func() {
	type testCase struct {
		listenerProtocol  xds.SocketAddressProtocol
		listenerAddress   string
		listenerPort      uint32
		enableReusedPorts bool
		expected          string
	}

	Context("V3", func() {
		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// when
				listener, err := NewInboundListenerBuilder(envoy.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol, given.enableReusedPorts).
					Configure(FilterChain(NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource))).
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
			Entry("basic TCP listener with an empty filter chain", testCase{
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
            filterChains:
            - {}
`,
			}),
			Entry("basic TCP listener with enable reusable port disabled", testCase{
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
            filterChains:
            - {}
`,
			}),
		)
	})
})
