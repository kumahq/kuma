package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
)

var _ = Describe("HttpInboundRouteConfigurer", func() {
	type testCase struct {
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		cluster          envoy_common.Cluster
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given: the inbound is named the way InboundProxyGenerator names it
			inboundName := naming.MustContextualInboundName(core_mesh.NewDataplaneResource(), given.listenerPort)

			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3, inboundName).
				Configure(InboundListener(given.listenerAddress, given.listenerPort, given.listenerProtocol, true)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(HttpConnectionManager(inboundName, true, nil, true)).
					Configure(HttpInboundRoute(inboundName, inboundName, given.cluster)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager with a single destination cluster", testCase{
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			cluster:         plugins_xds.NewClusterBuilder().WithName("self_inbound_dp_8080").Build(),
			expected: `
            name: self_inbound_dp_8080
            trafficDirection: INBOUND
            enableReusePort: true
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  forwardClientCertDetails: SANITIZE_SET
                  setCurrentClientCertDetails:
                    uri: true
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  routeConfig:
                    name: self_inbound_dp_8080
                    validateClusters: false
                    requestHeadersToRemove:
                    - x-kuma-tags
                    virtualHosts:
                    - domains:
                      - '*'
                      name: self_inbound_dp_8080
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: self_inbound_dp_8080
                          timeout: 0s
                  statPrefix: self_inbound_dp_8080
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 127.0.0.1
                        prefixLen: 32
                      - addressPrefix: ::1
                        prefixLen: 128`,
		}),
	)
})
