package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

			context := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("backend").
						AddInbound(builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777)),
				).
				WithOutbounds(core_xds.Outbounds{
					{LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build()},
				}).
				WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, given.singleItemRules)).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, proxy)).To(Succeed())
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
                      spawnUpstreamSpan: false
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
                      spawnUpstreamSpan: false
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
            connectTimeout: 5s
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
                      spawnUpstreamSpan: false
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
                      spawnUpstreamSpan: false
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
            connectTimeout: 5s
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
                      spawnUpstreamSpan: false
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
                      spawnUpstreamSpan: false
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
            connectTimeout: 5s
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
                            spawnUpstreamSpan: false
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
                            spawnUpstreamSpan: false
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
            connectTimeout: 5s
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
	type gatewayTestCase struct {
		rules core_rules.SingleItemRules
	}
	DescribeTable("should generate proper Envoy config for gateways",
		func(given gatewayTestCase) {
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
			}

			xdsCtx := xds_samples.SampleContextWith(resources)

			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, given.rules)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]
			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.listeners.golden.yaml", name))))
		},
		Entry("simple-gateway", gatewayTestCase{
			rules: core_rules.SingleItemRules{
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
		}),
	)
})

func getResourceYaml(list core_xds.ResourceList) []byte {
	actualResource, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualResource
}
