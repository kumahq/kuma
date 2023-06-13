package v3_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("RetryConfigurer", func() {
	type testCase struct {
		listenerName     string
		listenerAddress  string
		listenerPort     uint32
		listenerProtocol core_xds.SocketAddressProtocol
		statsName        string
		service          string
		routes           envoy_common.Routes
		dpTags           mesh_proto.MultiValueTagSet
		protocol         core_mesh.Protocol
		retry            *core_mesh.RetryResource
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager(given.statsName, false)).
					Configure(HttpOutboundRoute(
						given.service,
						given.routes,
						given.dpTags,
					)).
					Configure(Retry(given.retry, given.protocol)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager with an outbound route"+
			" and simple http retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:17777",
			listenerAddress: "127.0.0.1",
			listenerPort:    17777,
			statsName:       "127.0.0.1:17777",
			service:         "backend",
			routes: envoy_common.Routes{
				{
					Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
						envoy_common.WithService("backend"),
						envoy_common.WithWeight(100),
					)},
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			protocol: "http",
			retry: &core_mesh.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Http: &mesh_proto.Retry_Conf_Http{
							NumRetries:       util_proto.UInt32(7),
							RetriableMethods: []mesh_proto.HttpMethod{mesh_proto.HttpMethod_GET, mesh_proto.HttpMethod_POST},
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
                      retryPolicy:
                        numRetries: 7
                        retriableRequestHeaders:
                        - stringMatch: 
                            exact: GET
                          name: :method
                        - stringMatch: 
                            exact: POST
                          name: :method
                        retryOn: gateway-error,connect-failure,refused-stream
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
		Entry("basic http_connection_manager with an outbound route"+
			" and more complex http retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			service:         "backend",
			routes: envoy_common.Routes{
				{
					Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
						envoy_common.WithService("backend"),
						envoy_common.WithWeight(100),
					)},
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			protocol: "http",
			retry: &core_mesh.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Http: &mesh_proto.Retry_Conf_Http{
							NumRetries:    util_proto.UInt32(3),
							PerTryTimeout: util_proto.Duration(time.Second * 1),
							BackOff: &mesh_proto.Retry_Conf_BackOff{
								BaseInterval: util_proto.Duration(time.Nanosecond * 200000000),
								MaxInterval:  util_proto.Duration(time.Nanosecond * 500000000),
							},
							RetriableStatusCodes: []uint32{500, 502},
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
                      retryPolicy:
                        numRetries: 3
                        perTryTimeout: 1s
                        retriableStatusCodes:
                        - 500
                        - 502
                        retryBackOff:
                          baseInterval: 0.200s
                          maxInterval: 0.500s
                        retryOn: gateway-error,connect-failure,refused-stream,retriable-status-codes
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
		Entry("basic http_connection_manager with an outbound route"+
			" and simple grpc retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:17777",
			listenerAddress: "127.0.0.1",
			listenerPort:    17777,
			statsName:       "127.0.0.1:17777",
			service:         "backend",
			routes: envoy_common.Routes{
				{
					Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
						envoy_common.WithService("backend"),
						envoy_common.WithWeight(100),
					)},
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			protocol: "grpc",
			retry: &core_mesh.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Grpc: &mesh_proto.Retry_Conf_Grpc{
							NumRetries: util_proto.UInt32(18),
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
                      retryPolicy:
                        numRetries: 18
                        retryOn: cancelled,connect-failure,gateway-error,refused-stream,reset,resource-exhausted,unavailable
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
		Entry("basic http_connection_manager with an outbound route"+
			" and more complex http retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			service:         "backend",
			routes: envoy_common.Routes{
				{
					Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
						envoy_common.WithService("backend"),
						envoy_common.WithWeight(100),
					)},
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			protocol: "grpc",
			retry: &core_mesh.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Grpc: &mesh_proto.Retry_Conf_Grpc{
							NumRetries:    util_proto.UInt32(2),
							PerTryTimeout: util_proto.Duration(time.Second * 2),
							BackOff: &mesh_proto.Retry_Conf_BackOff{
								BaseInterval: util_proto.Duration(time.Nanosecond * 400000000),
								MaxInterval:  util_proto.Duration(time.Second * 1),
							},
							RetryOn: []mesh_proto.Retry_Conf_Grpc_RetryOn{
								mesh_proto.Retry_Conf_Grpc_cancelled,
								mesh_proto.Retry_Conf_Grpc_resource_exhausted,
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
                      retryPolicy:
                        numRetries: 2
                        perTryTimeout: 2s
                        retryBackOff:
                          baseInterval: 0.400s
                          maxInterval: 1s
                        retryOn: cancelled,resource-exhausted
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
		Entry("basic http_connection_manager with an outbound route"+
			" and more complex http retry policy", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			service:         "backend",
			routes: envoy_common.Routes{
				{
					Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
						envoy_common.WithService("backend"),
						envoy_common.WithWeight(100),
					)},
				},
			},
			dpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
			protocol: "http",
			retry: &core_mesh.RetryResource{
				Spec: &mesh_proto.Retry{
					Conf: &mesh_proto.Retry_Conf{
						Http: &mesh_proto.Retry_Conf_Http{
							NumRetries:    util_proto.UInt32(3),
							PerTryTimeout: util_proto.Duration(time.Second * 1),
							BackOff: &mesh_proto.Retry_Conf_BackOff{
								BaseInterval: util_proto.Duration(time.Nanosecond * 200000000),
								MaxInterval:  util_proto.Duration(time.Nanosecond * 500000000),
							},
							RetriableStatusCodes: []uint32{410},
							RetryOn: []mesh_proto.HttpRetryOn{
								mesh_proto.HttpRetryOn_all_5xx,
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
                      retryPolicy:
                        numRetries: 3
                        perTryTimeout: 1s
                        retriableStatusCodes:
                        - 410
                        retryBackOff:
                          baseInterval: 0.200s
                          maxInterval: 0.500s
                        retryOn: 5xx,retriable-status-codes
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
