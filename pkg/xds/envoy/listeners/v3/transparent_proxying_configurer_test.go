package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	tproxy_dp "github.com/kumahq/kuma/pkg/transparentproxy/config/dataplane"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("TransparentProxyingConfigurer", func() {
	type testCase struct {
		listenerProtocol    xds.SocketAddressProtocol
		listenerAddress     string
		listenerPort        uint32
		transparentProxying *tproxy_dp.DataplaneConfig
		expected            string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewInboundListenerBuilder(envoy.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol).
				Configure(TransparentProxying(given.transparentProxying)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic listener with transparent proxying", testCase{
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			transparentProxying: &tproxy_dp.DataplaneConfig{
				Redirect: tproxy_dp.DataplaneRedirect{
					Inbound:  tproxy_dp.DataplaneTrafficFlowFromPortLike(12345),
					Outbound: tproxy_dp.DataplaneTrafficFlowFromPortLike(12346),
				},
			},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
            bindToPort: false
`,
		}),
		Entry("basic listener without transparent proxying", testCase{
			listenerAddress:     "192.168.0.1",
			listenerPort:        8080,
			transparentProxying: &tproxy_dp.DataplaneConfig{},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
`,
		}),
	)
})
