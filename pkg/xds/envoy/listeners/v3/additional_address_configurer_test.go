package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("AdditionalAddressConfigurer", func() {
	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32

		additionalAddress     string
		additionalAddressPort uint32

		serviceName string
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			tcpSplit := plugins_xds.NewSplitBuilder().
				WithClusterName(given.serviceName).
				WithWeight(uint32(100)).
				Build()
			oface := []mesh_proto.OutboundInterface{
				{
					DataplaneIP:   given.additionalAddress,
					DataplanePort: given.additionalAddressPort,
				},
			}
			// when
			listener, err := NewOutboundListenerBuilder(envoy_common.APIV3, given.listenerAddress, given.listenerPort, xds.SocketAddressProtocolTCP).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).Configure(TCPProxy(given.serviceName, tcpSplit)))).
				Configure(AdditionalAddresses(oface)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("generate listener with additional addresses", testCase{
			listenerName:          "outbound:192.168.24.58:8080",
			listenerAddress:       "192.168.24.58",
			listenerPort:          8080,
			additionalAddress:     "240.0.0.1",
			additionalAddressPort: 80,
			serviceName:           "httpbin_app-ns_svc_8080",
			expected: `
              name: outbound:192.168.24.58:8080
              trafficDirection: OUTBOUND
              address:
                socketAddress:
                  address: 192.168.24.58
                  portValue: 8080
              additionalAddresses:
              - address:
                  socketAddress:
                    address: 240.0.0.1
                    portValue: 80
              filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: httpbin_app-ns_svc_8080
                      statPrefix: httpbin_app-ns_svc_8080
`,
		}),
	)
})
