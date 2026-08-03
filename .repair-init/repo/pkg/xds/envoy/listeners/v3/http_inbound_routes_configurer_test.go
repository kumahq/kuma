package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
)

var _ = Describe("HttpInboundRouteConfigurer", func() {
	type testCase struct {
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		statsName        string
		service          string
		routes           envoy_common.Routes
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewInboundListenerBuilder(envoy_common.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol, true).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(HttpConnectionManager(given.statsName, true, nil, true)).
					Configure(HttpInboundRoutes(envoy_names.GetInboundRouteName(given.service), given.service, given.routes)))).
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
			statsName:       "localhost:8080",
			service:         "backend",
			routes: envoy_common.Routes{envoy_common.NewRouteFromCluster(envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
			))},
			expected: `
            name: inbound:192.168.0.1:8080
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
                    name: inbound:backend
                    validateClusters: false
                    requestHeadersToRemove:
                    - x-kuma-tags
                    virtualHosts:
                    - domains:
                      - '*'
                      name: backend
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: localhost:8080
                          timeout: 0s
                  statPrefix: localhost_8080
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 127.0.0.1
                        prefixLen: 32
                      - addressPrefix: ::1
                        prefixLen: 128`,
		}),
	)
})
