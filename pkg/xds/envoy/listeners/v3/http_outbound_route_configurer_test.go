package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
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
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager(given.statsName, false)).
					Configure(HttpOutboundRoute(given.service, given.routes, given.dpTags)))).
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
                          timeout: 0s
                          weightedClusters:
                            clusters:
                            - name: backend-0
                              weight: 20
                            - name: backend-1
                              weight: 80
                            totalWeight: 100
                  statPrefix: "127_0_0_1_18080"
            name: outbound:127.0.0.1:18080
            trafficDirection: OUTBOUND`,
		}),
		Entry("http_connection_manager with matching routes", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			service:         "backend",
			routes: envoy_common.Routes{
				{
					Match: &mesh_proto.TrafficRoute_Http_Match{
						Method: &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
							MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact{
								Exact: "GET",
							},
						},
						Path: &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
							MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix{
								Prefix: "/asd",
							},
						},
						Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
							"x-custom-header-a": {
								MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix{
									Prefix: "prefix",
								},
							},
							"x-custom-header-b": {
								MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact{
									Exact: "exact",
								},
							},
							"x-custom-header-c": {
								MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex{
									Regex: "^regex$",
								},
							},
						},
					},
					Modify: &mesh_proto.TrafficRoute_Http_Modify{
						Path: &mesh_proto.TrafficRoute_Http_Modify_Path{
							Type: &mesh_proto.TrafficRoute_Http_Modify_Path_RewritePrefix{
								RewritePrefix: "/another",
							},
						},
						Host: &mesh_proto.TrafficRoute_Http_Modify_Host{
							Type: &mesh_proto.TrafficRoute_Http_Modify_Host_Value{
								Value: "test",
							},
						},
						RequestHeaders: &mesh_proto.TrafficRoute_Http_Modify_Headers{
							Add: []*mesh_proto.TrafficRoute_Http_Modify_Headers_Add{
								{
									Name:   "test-add",
									Value:  "abc",
									Append: false,
								},
							},
							Remove: []*mesh_proto.TrafficRoute_Http_Modify_Headers_Remove{
								{
									Name: "test-remove",
								},
							},
						},
						ResponseHeaders: &mesh_proto.TrafficRoute_Http_Modify_Headers{
							Add: []*mesh_proto.TrafficRoute_Http_Modify_Headers_Add{
								{
									Name:   "test-add",
									Value:  "abc",
									Append: false,
								},
							},
							Remove: []*mesh_proto.TrafficRoute_Http_Modify_Headers_Remove{
								{
									Name: "test-remove",
								},
							},
						},
					},
					Clusters: []envoy_common.Cluster{
						envoy_common.NewCluster(
							envoy_common.WithName("backend-0"),
							envoy_common.WithWeight(20),
							envoy_common.WithTags(map[string]string{"version": "v1"}),
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
                    requestHeadersToAdd:
                    - header:
                        key: x-kuma-tags
                        value: '&kuma.io/service=web&'
                    validateClusters: false
                    virtualHosts:
                    - domains:
                      - '*'
                      name: backend
                      routes:
                      - match:
                          headers:
                          - name: x-custom-header-a
                            prefixMatch: prefix
                          - stringMatch:
                              exact: exact
                            name: x-custom-header-b
                          - name: x-custom-header-c
                            safeRegexMatch:
                              googleRe2: {}
                              regex: ^regex$
                          - stringMatch:
                              exact: GET
                            name: :method
                          prefix: /asd
                        requestHeadersToAdd:
                        - append: false
                          header:
                            key: test-add
                            value: abc
                        requestHeadersToRemove:
                        - test-remove
                        responseHeadersToAdd:
                        - append: false
                          header:
                            key: test-add
                            value: abc
                        responseHeadersToRemove:
                        - test-remove
                        route:
                          cluster: backend-0
                          hostRewriteLiteral: test
                          prefixRewrite: /another
                          timeout: 0s
                  statPrefix: "127_0_0_1_18080"
            name: outbound:127.0.0.1:18080
            trafficDirection: OUTBOUND`,
		}),
	)
})
