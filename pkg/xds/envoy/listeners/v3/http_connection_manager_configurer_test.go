package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("HttpConnectionManagerConfigurer", func() {
	type testCase struct {
		listenerProtocol  xds.SocketAddressProtocol
		listenerAddress   string
		listenerPort      uint32
		statsName         string
		internalAddresses []xds.InternalAddress
		expected          string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewInboundListenerBuilder(envoy.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol).
				Configure(FilterChain(NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(HttpConnectionManager(given.statsName, true, given.internalAddresses)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager", testCase{
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			internalAddresses: []xds.InternalAddress{
				{AddressPrefix: "192.168.0.0", PrefixLen: 16},
			},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  forwardClientCertDetails: SANITIZE_SET
                  setCurrentClientCertDetails:
                    uri: true
                  statPrefix: localhost_8080
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 192.168.0.0
                        prefixLen: 16`,
		}),
	)
})
