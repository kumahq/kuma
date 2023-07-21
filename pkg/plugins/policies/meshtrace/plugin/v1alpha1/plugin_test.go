package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshTrace", func() {
	type testCase struct {
		resources         []core_xds.Resource
		singleItemRules   core_rules.SingleItemRules
		expectedListeners []string
		expectedClusters  []string
	}
	inboundAndOutbound := func() []core_xds.Resource {
		return []core_xds.Resource{
			{
				Name:   "inbound",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)),
					)).MustBuild(),
			}, {
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)),
					)).MustBuild(),
			},
		}
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			context := xds_context.Context{}
			proxy := xds.Proxy{
				APIVersion: envoy_common.APIV3,
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
						api.MeshTraceType: {
							Type:            api.MeshTraceType,
							SingleItemRules: given.singleItemRules,
						},
					},
				},
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, &proxy)).To(Succeed())
			policies_xds.ResourceArrayShouldEqual(resources.ListOf(envoy_resource.ListenerType), given.expectedListeners)
			policies_xds.ResourceArrayShouldEqual(resources.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("inbound/outbound for zipkin", testCase{
			resources: inboundAndOutbound(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "app", Literal: pointer.To("backend")},
								{Name: "app_code", Header: &api.HeaderTag{Name: "app_code"}},
								{Name: "client_id", Header: &api.HeaderTag{Name: "client_id", Default: pointer.To("none")}},
							},
							Sampling: &api.Sampling{
								Overall: pointer.To(intstr.FromInt(10)),
								Client:  pointer.To(intstr.FromInt(20)),
								Random:  pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: pointer.To(true),
									ApiVersion:        pointer.To("httpProto"),
									TraceId128Bit:     pointer.To(true),
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
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
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_17777"
                  tracing:
                      clientSampling:
                          value: 20
                      customTags:
                          - literal:
                              value: backend
                            tag: app
                          - requestHeader:
                              name: app_code
                            tag: app_code
                          - requestHeader:
                              defaultValue: none
                              name: client_id
                            tag: client_id
                      overallSampling:
                          value: 10
                      provider:
                          name: envoy.tracers.zipkin
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                              collectorCluster: meshtrace:zipkin
                              collectorEndpoint: /api/v2/spans
                              collectorEndpointVersion: HTTP_PROTO
                              collectorHostname: jaeger-collector.mesh-observability:9411
                              sharedSpanContext: true
                              traceId128bit: true
                      randomSampling:
                          value: 50
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`, `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_27777"
                  tracing:
                      clientSampling:
                          value: 20
                      customTags:
                          - literal:
                              value: backend
                            tag: app
                          - requestHeader:
                              name: app_code
                            tag: app_code
                          - requestHeader:
                              defaultValue: none
                              name: client_id
                            tag: client_id
                      overallSampling:
                          value: 10
                      provider:
                          name: envoy.tracers.zipkin
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                              collectorCluster: meshtrace:zipkin
                              collectorEndpoint: /api/v2/spans
                              collectorEndpointVersion: HTTP_PROTO
                              collectorHostname: jaeger-collector.mesh-observability:9411
                              sharedSpanContext: true
                              traceId128bit: true
                      randomSampling:
                          value: 50
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
			expectedClusters: []string{
				`
            altStatName: meshtrace_zipkin
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshtrace:zipkin
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: jaeger-collector.mesh-observability
                                    portValue: 9411
            name: meshtrace:zipkin
            type: STRICT_DNS
`,
			},
		}),
		Entry("inbound/outbound for opentelemetry", testCase{
			resources: inboundAndOutbound(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "app", Literal: pointer.To("backend")},
								{Name: "app_code", Header: &api.HeaderTag{Name: "app_code"}},
								{Name: "client_id", Header: &api.HeaderTag{Name: "client_id", Default: pointer.To("none")}},
							},
							Sampling: &api.Sampling{
								Overall: pointer.To(intstr.FromInt(10)),
								Client:  pointer.To(intstr.FromInt(20)),
								Random:  pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OpenTelemetryBackend{
									Endpoint: "jaeger-collector.mesh-observability:4317",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
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
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_17777"
                  tracing:
                      clientSampling:
                          value: 20
                      customTags:
                          - literal:
                              value: backend
                            tag: app
                          - requestHeader:
                              name: app_code
                            tag: app_code
                          - requestHeader:
                              defaultValue: none
                              name: client_id
                            tag: client_id
                      overallSampling:
                          value: 10
                      provider:
                          name: envoy.tracers.opentelemetry
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.OpenTelemetryConfig
                              grpcService: 
                                  envoyGrpc:
                                      clusterName: meshtrace:opentelemetry
                              serviceName: backend
                      randomSampling:
                          value: 50
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`, `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_27777"
                  tracing:
                      clientSampling:
                          value: 20
                      customTags:
                          - literal:
                              value: backend
                            tag: app
                          - requestHeader:
                              name: app_code
                            tag: app_code
                          - requestHeader:
                              defaultValue: none
                              name: client_id
                            tag: client_id
                      overallSampling:
                          value: 10
                      provider:
                          name: envoy.tracers.opentelemetry
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.OpenTelemetryConfig
                              grpcService: 
                                  envoyGrpc:
                                      clusterName: meshtrace:opentelemetry
                              serviceName: backend
                      randomSampling:
                          value: 50
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND
`,
			},
			expectedClusters: []string{
				`
            altStatName: meshtrace_opentelemetry
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshtrace:opentelemetry
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: jaeger-collector.mesh-observability
                                    portValue: 4317
            name: meshtrace:opentelemetry
            type: STRICT_DNS
            typedExtensionProtocolOptions:
                envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
                    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
                    explicitHttpConfig:
                        http2ProtocolOptions: {}
`,
			},
		}),
		Entry("inbound/outbound for datadog", testCase{
			resources: inboundAndOutbound(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Sampling: &api.Sampling{
								Random: pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Datadog: &api.DatadogBackend{
									Url:          "http://ingest.datadog.eu:8126",
									SplitService: pointer.To(true),
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
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
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_17777"
                  tracing:
                      provider:
                          name: envoy.tracers.datadog
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.DatadogConfig
                              collectorCluster: meshtrace:datadog
                              serviceName: backend_INBOUND
                      randomSampling:
                          value: 50
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`, `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 27777
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_27777"
                  tracing:
                      provider:
                          name: envoy.tracers.datadog
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.DatadogConfig
                              collectorCluster: meshtrace:datadog
                              serviceName: backend_OUTBOUND_other-service
                      randomSampling:
                          value: 50
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
			expectedClusters: []string{
				`
            altStatName: meshtrace_datadog
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshtrace:datadog
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: ingest.datadog.eu
                                    portValue: 8126
            name: meshtrace:datadog
            type: STRICT_DNS
`,
			},
		}),
		Entry("sampling is empty", testCase{
			resources: inboundAndOutbound(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: pointer.To(true),
									TraceId128Bit:     pointer.To(true),
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{
				`
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
                        httpFilters:
                            - name: envoy.filters.http.router
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                        statPrefix: "127_0_0_1_17777"
                        tracing:
                            provider:
                                name: envoy.tracers.zipkin
                                typedConfig:
                                    '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                                    collectorCluster: meshtrace:zipkin
                                    collectorEndpoint: /api/v2/spans
                                    collectorEndpointVersion: HTTP_JSON
                                    collectorHostname: jaeger-collector.mesh-observability:9411
                                    sharedSpanContext: true
                                    traceId128bit: true
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`,
				`
            address:
                socketAddress:
                    address: 127.0.0.1
                    portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.http_connection_manager
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                        httpFilters:
                            - name: envoy.filters.http.router
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                        statPrefix: "127_0_0_1_27777"
                        tracing:
                            provider:
                                name: envoy.tracers.zipkin
                                typedConfig:
                                    '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                                    collectorCluster: meshtrace:zipkin
                                    collectorEndpoint: /api/v2/spans
                                    collectorEndpointVersion: HTTP_JSON
                                    collectorHostname: jaeger-collector.mesh-observability:9411
                                    sharedSpanContext: true
                                    traceId128bit: true
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
			expectedClusters: []string{
				`
            altStatName: meshtrace_zipkin
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshtrace:zipkin
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: jaeger-collector.mesh-observability
                                    portValue: 9411
            name: meshtrace:zipkin
            type: STRICT_DNS
`,
			},
		}),
		Entry("backends list is empty", testCase{
			resources: inboundAndOutbound(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{},
						},
					},
				},
			},
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
                        httpFilters:
                            - name: envoy.filters.http.router
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                        statPrefix: "127_0_0_1_17777"
            name: inbound:127.0.0.1:17777
            trafficDirection: INBOUND`, `
            address:
                socketAddress:
                    address: 127.0.0.1
                    portValue: 27777
            filterChains:
                - filters:
                    - name: envoy.filters.network.http_connection_manager
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                        httpFilters:
                            - name: envoy.filters.http.router
                              typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                        statPrefix: "127_0_0_1_27777"
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`},
		}),
	)
})
