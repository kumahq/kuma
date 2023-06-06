package v1alpha1_test

import (
	"path/filepath"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshRateLimit", func() {
	type sidecarTestCase struct {
		resources            []*core_xds.Resource
		fromRules            core_rules.FromRules
		inboundRateLimitsMap core_xds.InboundRateLimitsMap
		expectedListeners    []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			context := test_xds.CreateSampleMeshContext()
			proxy := core_xds.Proxy{
				Dataplane: builders.Dataplane().
					WithName("test").
					WithMesh("default").
					WithAddress("127.0.0.1").
					AddInbound(
						builders.Inbound().
							WithAddress("127.0.0.1").
							WithPort(17777).
							WithService("backend"),
					).
					AddInbound(
						builders.Inbound().
							WithAddress("127.0.0.1").
							WithPort(17778).
							WithService("frontend"),
					).
					Build(),
				Policies: core_xds.MatchedPolicies{
					RateLimitsInbound: given.inboundRateLimitsMap,
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						api.MeshRateLimitType: {
							Type:      api.MeshRateLimitType,
							FromRules: given.fromRules,
						},
					},
				},
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())

			// then
			for i, expected := range given.expectedListeners {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[i].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("basic listener: 2 inbounds one http and second tcp", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:   "inbound:127.0.0.1:17777",
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
											Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
												envoy_common.WithService("backend"),
												envoy_common.WithWeight(100),
											)},
										},
									},
								),
							),
						)).MustBuild(),
				},
				{
					Name:   "inbound:127.0.0.1:17778",
					Origin: generator.OriginInbound,
					Resource: NewListenerBuilder(envoy_common.APIV3).
						Configure(InboundListener("inbound:127.0.0.1:17778", "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
							Configure(TcpProxy("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
									OnRateLimit: &api.OnRateLimit{
										Status: pointer.To(uint32(444)),
										Headers: &api.HeaderModifier{
											Add: []api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header",
													Value: "test-value",
												},
												{
													Name:  "x-kuma-rate-limit",
													Value: "other-value",
												},
											},
											Set: []api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header-set",
													Value: "test-value",
												},
											},
										},
									},
								},
							},
						},
					}},
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
								TCP: &api.LocalTCP{
									ConnectionRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
			},
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"basic_listener_1.golden.yaml", "basic_listener_2.golden.yaml"},
		}),
		Entry("old policy defined and no changes from MeshRateLimit", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:   "inbound:127.0.0.1:17777",
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
											Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
												envoy_common.WithService("backend"),
												envoy_common.WithWeight(100),
											)},
											RateLimit: &mesh_proto.RateLimit{
												Sources: []*mesh_proto.Selector{
													{
														Match: map[string]string{
															mesh_proto.ServiceTag: "frontend",
														},
													},
												},
												Conf: &mesh_proto.RateLimit_Conf{
													Http: &mesh_proto.RateLimit_Conf_Http{
														Requests: 100,
														Interval: &durationpb.Duration{Seconds: 14},
													},
												},
											},
										},
									},
								),
							),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{
										Num:      100,
										Interval: *test.ParseDuration("10s"),
									},
									OnRateLimit: &api.OnRateLimit{
										Status: pointer.To(uint32(444)),
										Headers: &api.HeaderModifier{
											Add: []api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header",
													Value: "test-value",
												},
											},
											Set: []api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit",
													Value: "other-value",
												},
											},
										},
									},
								},
								TCP: &api.LocalTCP{
									ConnectionRate: &api.Rate{
										Num:      100,
										Interval: *test.ParseDuration("99s"),
									},
								},
							},
						},
					}},
				},
			},
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{
				mesh_proto.InboundInterface{
					DataplaneAdvertisedIP: "127.0.0.1",
					DataplaneIP:           "127.0.0.1",
					DataplanePort:         17777,
					WorkloadIP:            "127.0.0.1",
					WorkloadPort:          17777,
				}: []*core_mesh.RateLimitResource{
					{
						Spec: &mesh_proto.RateLimit{
							Sources: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "frontend",
									},
								},
							},
							Conf: &mesh_proto.RateLimit_Conf{
								Http: &mesh_proto.RateLimit_Conf_Http{
									Requests: 100,
									Interval: &durationpb.Duration{Seconds: 14},
								},
							},
						},
					},
				},
			},
			expectedListeners: []string{"old_policy.golden.yaml"},
		}),
		Entry("tcp rate limiter is disabled", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17778",
				Origin: generator.OriginInbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(InboundListener("inbound:127.0.0.1:17778", "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxy("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								TCP: &api.LocalTCP{
									Disabled:       pointer.To(true),
									ConnectionRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
			},
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"tcp_disabled.golden.yaml"},
		}),
		Entry("http rate limiter is disabled", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
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
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
									},
								},
							),
						),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									Disabled:    pointer.To(true),
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
			},
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"http_disabled.golden.yaml"},
		}),
		Entry("tcp rate limiter is not configured", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17778",
				Origin: generator.OriginInbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(InboundListener("inbound:127.0.0.1:17778", "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxy("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								TCP: &api.LocalTCP{
									ConnectionRate: nil,
								},
							},
						},
					}},
				},
			},
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"tcp_disabled.golden.yaml"},
		}),
		Entry("http rate limiter is not configured", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
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
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
									},
								},
							),
						),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: nil,
								},
							},
						},
					}},
				},
			},
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"http_disabled.golden.yaml"},
		}),
	)

	It("should generate proper Envoy config for MeshGateway Dataplanes", func() {
		// given
		fromRules := core_rules.FromRules{
			Rules: map[core_rules.InboundListener]core_rules.Rules{
				{Address: "192.168.0.1", Port: 8080}: {{
					Subset: core_rules.Subset{},
					Conf: api.Conf{
						Local: &api.Local{
							HTTP: &api.LocalHTTP{
								RequestRate: &api.Rate{
									Num:      100,
									Interval: v1.Duration{Duration: 10 * time.Second},
								},
								OnRateLimit: &api.OnRateLimit{
									Status: pointer.To(uint32(444)),
									Headers: &api.HeaderModifier{
										Add: []api.HeaderKeyValue{
											{
												Name:  "x-kuma-rate-limit-header",
												Value: "test-value",
											},
											{
												Name:  "x-kuma-rate-limit",
												Value: "other-value",
											},
										},
									},
								},
							},
						},
					},
				}},
			},
		}

		resources := xds_context.NewResources()
		resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
			Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
		}
		resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
			Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
		}

		context := test_xds.CreateSampleMeshContextWith(resources)
		proxy := xds.Proxy{
			APIVersion: "v3",
			Dataplane:  samples.GatewayDataplane(),
			Policies: core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					api.MeshRateLimitType: {
						Type:      api.MeshRateLimitType,
						FromRules: fromRules,
					},
				},
			},
		}
		gatewayGenerator := gatewayGenerator()
		generatedResources, err := gatewayGenerator.Generate(context, &proxy)
		Expect(err).NotTo(HaveOccurred())

		// when
		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

		// then
		Expect(plugin.Apply(generatedResources, context, &proxy)).To(Succeed())
		Expect(util_proto.ToYAML(generatedResources.ListOf(envoy_resource.RouteType)[0].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", "gateway_basic_routes.golden.yaml")))
		Expect(util_proto.ToYAML(generatedResources.ListOf(envoy_resource.ListenerType)[0].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", "gateway_basic_listener.golden.yaml")))
	})
})

func gatewayGenerator() gateway_plugin.Generator {
	return gateway_plugin.Generator{
		FilterChainGenerators: gateway_plugin.FilterChainGenerators{
			FilterChainGenerators: map[mesh_proto.MeshGateway_Listener_Protocol]gateway_plugin.FilterChainGenerator{
				mesh_proto.MeshGateway_Listener_HTTP:  &gateway_plugin.HTTPFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_HTTPS: &gateway_plugin.HTTPSFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_TCP:   &gateway_plugin.TCPFilterChainGenerator{},
			},
		},
		ClusterGenerator: gateway_plugin.ClusterGenerator{
			Zone: "test-zone",
		},
		Zone: "test-zone",
	}
}
