package listeners_test

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("RetryConfigurer", func() {
	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		service         string
		subsets         []envoy_common.ClusterSubset
		dpTags          mesh_proto.MultiValueTagSet
		retry           *mesh_core.RetryResource
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder().
				Configure(OutboundListener(
					given.listenerName,
					given.listenerAddress,
					given.listenerPort,
				)).
				Configure(FilterChain(NewFilterChainBuilder().
					Configure(HttpConnectionManager(given.statsName)).
					Configure(HttpOutboundRoute(
						given.service,
						given.subsets,
						given.dpTags,
					)).
					Configure(Retry(given.retry, "http")))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager with an outbound route and simple retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:17777",
			listenerAddress: "127.0.0.1",
			listenerPort:    17777,
			statsName:       "127.0.0.1:17777",
			service:         "backend",
			subsets: []envoy_common.ClusterSubset{
				{
					ClusterName: "backend",
					Weight:      100,
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			retry: &mesh_core.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Protocol: &mesh_proto.Retry_Conf_Http_{
							Http: &mesh_proto.Retry_Conf_Http{
								NumRetries: &wrappers.UInt32Value{
									Value: 7,
								},
							},
						},
					},
				},
			},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 17777
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
                      retryPolicy:
                        numRetries: 7
                        retryOn: 5XX
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_17777"
            name: outbound:127.0.0.1:17777
            trafficDirection: OUTBOUND`,
		}),
		Entry("basic http_connection_manager with an outbound route and more complex retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			service:         "backend",
			subsets: []envoy_common.ClusterSubset{
				{
					ClusterName: "backend",
					Weight:      100,
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			retry: &mesh_core.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Protocol: &mesh_proto.Retry_Conf_Http_{
							Http: &mesh_proto.Retry_Conf_Http{
								NumRetries: &wrappers.UInt32Value{
									Value: 3,
								},
								PerTryTimeout: &duration.Duration{
									Seconds: 1,
								},
								BackOff: &mesh_proto.Retry_Conf_BackOff{
									BaseInterval: &duration.Duration{
										Nanos: 200000000,
									},
									MaxInterval: &duration.Duration{
										Nanos: 500000000,
									},
								},
								RetriableStatusCodes: []uint32{500, 502},
							},
						},
					},
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
                      retryPolicy:
                        numRetries: 3
                        perTryTimeout: 1s
                        retriableStatusCodes:
                        - 500
                        - 502
                        retryBackOff:
                          baseInterval: 0.200s
                          maxInterval: 0.500s
                        retryOn: retriable-status-codes
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_18080"
            name: outbound:127.0.0.1:18080
            trafficDirection: OUTBOUND`,
		}),
	)
})
