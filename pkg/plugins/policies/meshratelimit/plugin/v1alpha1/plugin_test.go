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
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/plugin/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshRateLimit", func() {
	type sidecarTestCase struct {
		resources            []*core_xds.Resource
		fromRules            core_xds.FromRules
		inboundRateLimitsMap core_xds.InboundRateLimitsMap
		expectedListeners    []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			context := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}
			proxy := core_xds.Proxy{
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "test",
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
								{
									Tags: map[string]string{
										mesh_proto.ServiceTag: "frontend",
									},
									Address: "127.0.0.1",
									Port:    17778,
								},
							},
						},
					},
				},
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

			Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())
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
				}},
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Local: api.Local{
								HTTP: &api.LocalHTTP{
									Requests: 100,
									Interval: v1.Duration{Duration: 10 * time.Second},
									OnRateLimit: &api.OnRateLimit{
										Status: policies_xds.PointerOf(uint32(444)),
										Headers: []api.HeaderValue{
											{
												Key:    "x-kuma-rate-limit-header",
												Value:  "test-value",
												Append: policies_xds.PointerOf(true),
											},
											{
												Key:   "x-kuma-rate-limit",
												Value: "-value",
											},
										},
									},
								},
							},
						},
					}},
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Local: api.Local{
								HTTP: &api.LocalHTTP{
									Requests: 100,
									Interval: v1.Duration{Duration: 10 * time.Second},
								},
								TCP: &api.LocalTCP{
									Connections: 100,
									Interval:    v1.Duration{Duration: 100 * time.Millisecond},
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
				}},
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Local: api.Local{
								HTTP: &api.LocalHTTP{
									Requests: 100,
									Interval: v1.Duration{Duration: 10 * time.Second},
									OnRateLimit: &api.OnRateLimit{
										Status: policies_xds.PointerOf(uint32(444)),
										Headers: []api.HeaderValue{
											{
												Key:    "x-kuma-rate-limit-header",
												Value:  "test-value",
												Append: policies_xds.PointerOf(true),
											},
											{
												Key:   "x-kuma-rate-limit",
												Value: "other-value",
											},
										},
									},
								},
								TCP: &api.LocalTCP{
									Connections: 100,
									Interval:    v1.Duration{Duration: 99 * time.Second},
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
	)

	It("should generate proper Envoy config for MeshGateway Dataplanes",
		func() {
			fromRules := core_xds.FromRules{
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{Address: "127.0.0.1", Port: 8080}: {{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Local: api.Local{
								HTTP: &api.LocalHTTP{
									Requests: 100,
									Interval: v1.Duration{Duration: 10 * time.Second},
									OnRateLimit: &api.OnRateLimit{
										Status: policies_xds.PointerOf(uint32(444)),
										Headers: []api.HeaderValue{
											{
												Key:    "x-kuma-rate-limit-header",
												Value:  "test-value",
												Append: policies_xds.PointerOf(true),
											},
											{
												Key:   "x-kuma-rate-limit",
												Value: "other-value",
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
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = GatewayResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = GatewayRoutes()

			context := createSimpleMeshContextWith(resources)
			proxy := core_xds.Proxy{
				APIVersion: "v3",
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "gateway",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "127.0.0.1",
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{
								Tags: map[string]string{
									mesh_proto.ServiceTag: "gateway",
								},
								Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
							},
						},
					},
				},
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
			if err != nil {
				return
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(generatedResources, context, &proxy)).To(Succeed())
			Expect(util_proto.ToYAML(generatedResources.ListOf(envoy_resource.RouteType)[0].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", "gateway_basic_routes.golden.yaml")))
			Expect(util_proto.ToYAML(generatedResources.ListOf(envoy_resource.ListenerType)[0].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", "gateway_basic_listener.golden.yaml")))
		})
})

func GatewayResources() *core_mesh.MeshGatewayResourceList {
	return &core_mesh.MeshGatewayResourceList{
		Items: []*core_mesh.MeshGatewayResource{{
			Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
			Spec: &mesh_proto.MeshGateway{
				Selectors: []*mesh_proto.Selector{{
					Match: map[string]string{
						mesh_proto.ServiceTag: "gateway",
					}},
				},
				Conf: &mesh_proto.MeshGateway_Conf{
					Listeners: []*mesh_proto.MeshGateway_Listener{
						{
							Protocol: mesh_proto.MeshGateway_Listener_HTTP,
							Port:     8080,
						},
					},
				},
			},
		}},
	}
}

func GatewayRoutes() *core_mesh.MeshGatewayRouteResourceList {
	return &core_mesh.MeshGatewayRouteResourceList{
		Items: []*core_mesh.MeshGatewayRouteResource{
			{
				Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
				Spec: &mesh_proto.MeshGatewayRoute{
					Selectors: []*mesh_proto.Selector{{
						Match: map[string]string{
							mesh_proto.ServiceTag: "gateway",
						}},
					},
					Conf: &mesh_proto.MeshGatewayRoute_Conf{
						Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
							Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
								Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
									{
										Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
											{Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
												Match: 0,
												Value: "/",
											}},
										},
										Backends: []*mesh_proto.MeshGatewayRoute_Backend{
											{
												Weight: 1,
												Destination: map[string]string{
													"kuma.io/service": "some-service",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func gatewayGenerator() gateway_plugin.Generator {
	return gateway_plugin.Generator{
		FilterChainGenerators: gateway_plugin.FilterChainGenerators{
			FilterChainGenerators: map[mesh_proto.MeshGateway_Listener_Protocol]gateway_plugin.FilterChainGenerator{
				mesh_proto.MeshGateway_Listener_HTTP:  &gateway_plugin.HTTPFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_HTTPS: &gateway_plugin.HTTPSFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_TCP:   &gateway_plugin.TCPFilterChainGenerator{},
			}},
		ClusterGenerator: gateway_plugin.ClusterGenerator{
			Zone: "test-zone",
		},
		Zone: "test-zone",
	}
}

func createSimpleMeshContextWith(resources xds_context.Resources) xds_context.Context {
	return xds_context.Context{
		Mesh: xds_context.MeshContext{
			Resource: &core_mesh.MeshResource{
				Meta: &test_model.ResourceMeta{
					Name: "default",
				},
			},
			Resources: resources,
			EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
				"some-service": {
					{
						Tags: map[string]string{
							"app": "some-service",
						},
					},
				},
			},
		},
		ControlPlane: &xds_context.ControlPlaneContext{CLACache: &test_xds.DummyCLACache{}, Zone: "test-zone"},
	}
}
