package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
)

var _ = Describe("HttpOutboundRouteConfigurer", func() {
	type testCase struct {
		listenerName     string
		listenerAddress  string
		listenerPort     uint32
		listenerProtocol core_xds.SocketAddressProtocol
		statsName        string
		service          string
		routes           envoy_common.Routes
		dpTags           mesh_proto.MultiValueTagSet
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewOutboundListenerBuilder(envoy_common.APIV3, given.listenerAddress, given.listenerPort, given.listenerProtocol).
				WithOverwriteName(given.listenerName).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(HttpConnectionManager(given.statsName, false, nil, true)).
					Configure(HttpOutboundRoute(envoy_names.GetOutboundRouteName(given.service), given.service, given.routes, given.dpTags)))).
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
			routes: envoy_common.Routes{
				{
					Clusters: []envoy_common.Cluster{
						envoy_common.NewCluster(
							envoy_common.WithName("backend-0"),
							envoy_common.WithWeight(20),
							envoy_common.WithTags(map[string]string{"version": "v1"}),
						),
						envoy_common.NewCluster(
							envoy_common.WithName("backend-1"),
							envoy_common.WithWeight(80),
							envoy_common.WithTags(map[string]string{"version": "v2"}),
						),
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
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
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
                            - name: backend-0
                              weight: 20
                            - name: backend-1
                              weight: 80
                  statPrefix: "127_0_0_1_18080"
                  internalAddressConfig:
                    cidrRanges:
                      - addressPrefix: 127.0.0.1
                        prefixLen: 32
                      - addressPrefix: ::1
                        prefixLen: 128
            name: outbound:127.0.0.1:18080
            trafficDirection: OUTBOUND`,
		}),
	)
})
