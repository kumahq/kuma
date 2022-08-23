package v3_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

var _ = Describe("HttpInboundRouteConfigurer", func() {

	routeWithRateLimiter := func(rateLimit *v1alpha1.RateLimit) envoy_common.Route {
		route := envoy_common.NewRouteFromCluster(envoy_common.NewCluster(
			envoy_common.WithService("localhost:8080"),
			envoy_common.WithWeight(200),
		))
		route.RateLimit = rateLimit

		if len(rateLimit.GetSources()) > 0 {
			route.Match = &v1alpha1.TrafficRoute_Http_Match{
				Headers: make(map[string]*v1alpha1.TrafficRoute_Http_Match_StringMatcher),
			}

			var selectorRegexs []string
			for _, selector := range rateLimit.SourceTags() {
				selectorRegexs = append(selectorRegexs, tags.MatchingRegex(selector))
			}
			regexOR := tags.RegexOR(selectorRegexs...)

			route.Match.Headers[v3.TagsHeaderName] = &v1alpha1.TrafficRoute_Http_Match_StringMatcher{
				MatcherType: &v1alpha1.TrafficRoute_Http_Match_StringMatcher_Regex{
					Regex: regexOR,
				},
			}
		}

		return route
	}

	type testCase struct {
		listenerName     string
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
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager(given.statsName, true)).
					Configure(HttpInboundRoutes(given.service, given.routes)))).
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
			routes:          envoy_common.Routes{routeWithRateLimiter(nil)},
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
`,
		}),
		Entry("basic http_connection_manager with a single destination cluster and rate limiter", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			service:         "backend",
			routes: envoy_common.Routes{routeWithRateLimiter(&v1alpha1.RateLimit{
				Sources:      nil,
				Destinations: nil,
				Conf: &v1alpha1.RateLimit_Conf{
					Http: &v1alpha1.RateLimit_Conf_Http{
						Requests: 100,
						Interval: util_proto.Duration(time.Second * 3),
						OnRateLimit: &v1alpha1.RateLimit_Conf_Http_OnRateLimit{
							Status: util_proto.UInt32(404),
							Headers: []*v1alpha1.RateLimit_Conf_Http_OnRateLimit_HeaderValue{
								{
									Key:    "x-local-rate-limit",
									Value:  "true",
									Append: util_proto.Bool(false),
								},
							},
						},
					},
				},
			})},
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
                        typedPerFilterConfig:
                          envoy.filters.http.local_ratelimit:
                            '@type': type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
                            filterEnabled:
                              defaultValue:
                                numerator: 100
                              runtimeKey: local_rate_limit_enabled
                            filterEnforced:
                              defaultValue:
                                numerator: 100
                              runtimeKey: local_rate_limit_enforced
                            responseHeadersToAdd:
                            - append: false
                              header:
                                key: x-local-rate-limit
                                value: "true"
                            statPrefix: rate_limit
                            status:
                              code: NotFound
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
			routes: envoy_common.Routes{routeWithRateLimiter(&v1alpha1.RateLimit{
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
						Requests: 100,
						Interval: util_proto.Duration(time.Second * 3),
						OnRateLimit: &v1alpha1.RateLimit_Conf_Http_OnRateLimit{
							Status: util_proto.UInt32(404),
							Headers: []*v1alpha1.RateLimit_Conf_Http_OnRateLimit_HeaderValue{
								{
									Key:    "x-local-rate-limit",
									Value:  "true",
									Append: util_proto.Bool(false),
								},
							},
						},
					},
				},
			})},
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
                            '@type': type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
                            filterEnabled:
                              defaultValue:
                                numerator: 100
                              runtimeKey: local_rate_limit_enabled
                            filterEnforced:
                              defaultValue:
                                numerator: 100
                              runtimeKey: local_rate_limit_enforced
                            responseHeadersToAdd:
                            - append: false
                              header:
                                key: x-local-rate-limit
                                value: "true"
                            statPrefix: rate_limit
                            status:
                              code: NotFound
                            tokenBucket:
                              fillInterval: 3s
                              maxTokens: 100
                              tokensPerFill: 100
                  statPrefix: localhost_8080
`,
		}),
	)
})
