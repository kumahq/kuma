package v1alpha1_test

import (
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/plugin/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshRateLimit", func() {
	type sidecarTestCase struct {
		resources         []core_xds.Resource
		toRules           core_xds.ToRules
		fromRules         core_xds.FromRules
		expectedListeners []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			context := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}
			proxy := xds.Proxy{
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "backend",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										mesh_proto.ServiceTag: "backend",
									},
									Address: "127.0.0.1",
									Port:    17777,
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Address: "127.0.0.1",
									Port:    27777,
									Tags: map[string]string{
										mesh_proto.ServiceTag: "other-service",
									},
								},
							},
						},
					},
				},
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshRateLimitType: {
							Type:      api.MeshRateLimitType,
							ToRules:   given.toRules,
							FromRules: given.fromRules,
						},
					},
				},
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ListenerType), given.expectedListeners)
		},

		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: generator.OriginInbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(InboundListener("inbound:127.0.0.1:17777", "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)).
						Configure(
							HttpInboundRoutes(
								"backend",
								envoy_common.Routes{
									{
										Match: &envoy_common.HttpMatch{
											Headers: map[string]*envoy_common.StringMatcher{
												v3.TagsHeaderName: {
													MatchType: envoy_common.Regex,
													Value: tags.MatchingRegex(map[string]string{
														"kuma.io/service":  "web",
														"kuma.io/protocol": "http",
													}),
												},
											},
										},
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
										Tags: []tags.Tags{
											{
												"kuma.io/service":  "web",
												"kuma.io/protocol": "http",
											},
										},
										RateLimit: &envoy_common.RateLimitConfiguration{
											Interval: time.Duration(1 * time.Second),
											Requests: 1222,
										},
									},
									{
										Match: &envoy_common.HttpMatch{
											Headers: map[string]*envoy_common.StringMatcher{
												v3.TagsHeaderName: {
													MatchType: envoy_common.Regex,
													Value:     "web",
												},
											},
										},
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
										Tags: []tags.Tags{
											{
												"kuma.io/service":  "web",
												"kuma.io/protocol": "http",
											},
										},
										RateLimit: &envoy_common.RateLimitConfiguration{
											Interval: time.Duration(1 * time.Second),
											Requests: 1222,
										},
									},
								},
							),
						),
					)).MustBuild(),
			}},
			fromRules: core_xds.FromRules{
				Rules: map[xds.InboundListener]xds.Rules{
					{Address: "127.0.0.1", Port: 17777}: {
						{
							Subset: core_xds.Subset{
								{
									Key:   "kuma.io/service",
									Value: "web",
								},
								{
									Key:   "kuma.io/protocol",
									Value: "http",
								},
							},
							Conf: api.Conf{
								Local: &api.Local{
									HTTP: &api.LocalHTTP{
										Requests: 100,
										Interval: v1.Duration{Duration: time.Hour * 1},
									},
								},
							},
						}},
				}},
			expectedListeners: []string{`
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 17777
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  accessLog:
                  - name: envoy.access_loggers.file
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                      logFormat:
                          textFormatSource:
                              inlineString: |
                                [%START_TIME%] default "%REQ(:method)% %REQ(x-envoy-original-path?:path)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(x-envoy-upstream-service-time)% "%REQ(x-forwarded-for)%" "%REQ(user-agent)%" "%REQ(x-b3-traceid?x-datadog-traceid)%" "%REQ(x-request-id)%" "%REQ(:authority)%" "unknown" "backend" "" "%UPSTREAM_HOST%"
                      path: /tmp/log
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
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_17777"
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`,
			}}),
	)
})
