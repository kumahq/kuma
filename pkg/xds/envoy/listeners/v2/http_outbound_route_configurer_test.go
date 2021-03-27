package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("HttpOutboundRouteConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerAddress  string
		listenerPort     uint32
		listenerProtocol core_xds.SocketAddressProtocol
		statsName        string
		service          string
		subsets          []envoy_common.ClusterSubset
		dpTags           mesh_proto.MultiValueTagSet
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV2).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV2).
					Configure(HttpConnectionManager(given.statsName)).
					Configure(HttpOutboundRoute(given.service, given.subsets, given.dpTags)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager with an outbound route", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			service:         "backend",
			subsets: []envoy_common.ClusterSubset{
				{
					ClusterName: "backend",
					Weight:      20,
					Tags: map[string]string{
						"version": "v1",
					},
				},
				{
					ClusterName: "backend",
					Weight:      80,
					Tags: map[string]string{
						"version": "v2",
					},
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 18080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                  routeConfig:
                    name: outbound:backend
                    validateClusters: false
                    requestHeadersToAdd:
                    - header:
                        key: x-kuma-tags
                        value: '&kuma.io/service=web&'
                    virtualHosts:
                    - domains:
                      - '*'
                      name: backend
                      routes:
                      - match:
                          prefix: /
                        route:
                          weightedClusters:
                            clusters:
                            - metadataMatch:
                                filterMetadata:
                                  envoy.lb:
                                    version: v1
                              name: backend
                              weight: 20
                            - metadataMatch:
                                filterMetadata:
                                  envoy.lb:
                                    version: v2
                              name: backend
                              weight: 80
                            totalWeight: 100
                          timeout: 0s
                  statPrefix: "127_0_0_1_18080"
            name: outbound:127.0.0.1:18080
            trafficDirection: OUTBOUND`,
		}),
	)
})
