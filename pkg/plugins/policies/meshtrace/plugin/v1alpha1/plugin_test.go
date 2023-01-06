package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
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
		singleItemRules   core_xds.SingleItemRules
		expectedListeners []string
		expectedClusters  []string
	}
	inboundAndOutbound := []core_xds.Resource{
		{
			Name:   "inbound",
			Origin: generator.OriginInbound,
			Resource: NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener("inbound:127.0.0.1:17777", "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager("127.0.0.1:17777", false)),
				)).MustBuild(),
		}, {
			Name:   "outbound",
			Origin: generator.OriginOutbound,
			Resource: NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager("127.0.0.1:27777", false)),
				)).MustBuild(),
		},
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
			resources: inboundAndOutbound,
			singleItemRules: core_xds.SingleItemRules{
				Rules: []*core_xds.Rule{
					{
						Subset: []core_xds.Tag{},
						Conf: api.Conf{
							Sampling: api.Sampling{
								Random: pointer.To(uint32(50)),
							},
							Backends: []api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: pointer.To(true),
									ApiVersion:        "httpProto",
									TraceId128Bit:     true,
								},
							}},
						},
					},
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
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_17777"
                  tracing:
                      provider:
                          name: envoy.zipkin
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
                      provider:
                          name: envoy.zipkin
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
			expectedClusters: []string{`
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
`},
		}),
		Entry("inbound/outbound for datadog", testCase{
			resources: inboundAndOutbound,
			singleItemRules: core_xds.SingleItemRules{
				Rules: []*core_xds.Rule{
					{
						Subset: []core_xds.Tag{},
						Conf: api.Conf{
							Sampling: api.Sampling{
								Random: pointer.To(uint32(50)),
							},
							Backends: []api.Backend{{
								Datadog: &api.DatadogBackend{
									Url:          "http://ingest.datadog.eu:8126",
									SplitService: true,
								},
							}},
						},
					},
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
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  statPrefix: "127_0_0_1_17777"
                  tracing:
                      provider:
                          name: envoy.datadog
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
                          name: envoy.datadog
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.DatadogConfig
                              collectorCluster: meshtrace:datadog
                              serviceName: backend_OUTBOUND_other-service
                      randomSampling:
                          value: 50
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
			expectedClusters: []string{`
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
`},
		}),
	)
})
