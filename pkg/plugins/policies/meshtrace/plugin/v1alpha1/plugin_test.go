package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

			for i, r := range resources.ListOf(envoy_resource.ListenerType) {
				actual, err := util_proto.ToYAML(r.Resource)
				Expect(err).ToNot(HaveOccurred())

				Expect(actual).To(MatchYAML(given.expectedListeners[i]))
			}

			for i, r := range resources.ListOf(envoy_resource.ClusterType) {
				actual, err := util_proto.ToYAML(r.Resource)
				Expect(err).ToNot(HaveOccurred())

				Expect(actual).To(MatchYAML(given.expectedClusters[i]))
			}
		},
		Entry("basic config", testCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
								"backend",
								envoy_common.Routes{{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								}},
								map[string]map[string]bool{
									"kuma.io/service": {
										"web": true,
									},
								},
							),
						),
					)).MustBuild(),
			}},
			singleItemRules: core_xds.SingleItemRules{
				Rules: []*core_xds.Rule{
					{
						Subset: []core_xds.Tag{},
						Conf: &api.MeshTrace_Conf{
							Backends: []*api.MeshTrace_Backend{{
								Zipkin: &api.MeshTrace_ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: wrapperspb.Bool(true),
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
                          cluster: backend
                          timeout: 0s
                  statPrefix: "127_0_0_1_27777"
                  tracing:
                      provider:
                          name: envoy.zipkin
                          typedConfig:
                              '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                              collectorCluster: meshTrace
                              collectorEndpoint: /api/v2/spans
                              collectorEndpointVersion: HTTP_PROTO
                              collectorHostname: jaeger-collector.mesh-observability:9411
                              sharedSpanContext: true
                              traceId128bit: true
            name: outbound:127.0.0.1:27777
            trafficDirection: OUTBOUND`,
			},
			expectedClusters: []string{`
            connectTimeout: 10s
            dnsLookupFamily: V4_ONLY
            loadAssignment:
                clusterName: meshTrace
                endpoints:
                    - lbEndpoints:
                        - endpoint:
                            address:
                                socketAddress:
                                    address: jaeger-collector.mesh-observability
                                    portValue: 9411
            name: meshTrace
            type: STRICT_DNS
`},
		}),
	)
})
