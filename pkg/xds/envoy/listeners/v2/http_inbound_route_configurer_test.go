package v2_test

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"

	"github.com/kumahq/kuma/pkg/core/xds"

	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("HttpInboundRouteConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		statsName        string
		service          string
		route            envoy_common.Route
		ratelimit        *v1alpha1.RateLimit
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV2).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV2).
					Configure(HttpConnectionManager(given.statsName, true)).
					Configure(HttpInboundRoute(given.service, given.route, given.ratelimit)))).
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
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			service:         "backend",
			route: envoy_common.NewRouteFromCluster(envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
			)),
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  forwardClientCertDetails: SANITIZE_SET
                  setCurrentClientCertDetails:
                    uri: true
                  httpFilters:
                  - name: envoy.filters.http.router
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
`,
		}),
		Entry("basic http_connection_manager with a single destination cluster and rate limiter", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			service:         "backend",
			route: envoy_common.NewRouteFromCluster(envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
			)),
			ratelimit: &v1alpha1.RateLimit{
				Sources:      nil,
				Destinations: nil,
				Conf: &v1alpha1.RateLimit_Conf{
					Http: &v1alpha1.RateLimit_Conf_Http{
						Connections: &wrappers.UInt32Value{
							Value: 100,
						},
						Interval: &duration.Duration{
							Seconds: 3,
						},
					},
				},
			},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager                  
                  forwardClientCertDetails: SANITIZE_SET
                  setCurrentClientCertDetails:
                    uri: true
                  httpFilters:
                  - name: envoy.filters.http.router
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
                        typedPerFilterConfig:
                          envoy.filters.http.local_ratelimit:
                            '@type': type.googleapis.com/envoy.config.filter.network.local_rate_limit.v2alpha.LocalRateLimit
                            statPrefix: rate_limit
                            tokenBucket:
                              fillInterval: 3s
                              maxTokens: 100
                              tokensPerFill: 100
                  statPrefix: localhost_8080
`,
		}),
		Entry("basic http_connection_manager with a single destination cluster and rate limiter with sources", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			service:         "backend",
			route: envoy_common.NewRouteFromCluster(envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
			)),
			ratelimit: &v1alpha1.RateLimit{
				Sources: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web1",
							"version": "1.0",
						},
					},
				},
				Destinations: nil,
				Conf: &v1alpha1.RateLimit_Conf{
					Http: &v1alpha1.RateLimit_Conf_Http{
						Connections: &wrappers.UInt32Value{
							Value: 100,
						},
						Interval: &duration.Duration{
							Seconds: 3,
						},
					},
				},
			},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  forwardClientCertDetails: SANITIZE_SET
                  setCurrentClientCertDetails:
                    uri: true
                  httpFilters:
                  - name: envoy.filters.http.router
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
                          headers:
                          - name: x-kuma-tags
                            safeRegexMatch:
                              googleRe2: {}
                              regex: .*&service=[^&]*web1[,&].*&version=[^&]*1\.0[,&].*
                          prefix: /
                        route:
                          cluster: localhost:8080
                          timeout: 0s
                        typedPerFilterConfig:
                          envoy.filters.http.local_ratelimit:
                            '@type': type.googleapis.com/envoy.config.filter.network.local_rate_limit.v2alpha.LocalRateLimit
                            statPrefix: rate_limit
                            tokenBucket:
                              fillInterval: 3s
                              maxTokens: 100
                              tokensPerFill: 100
                  statPrefix: localhost_8080
`,
		}),
	)
})
