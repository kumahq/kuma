package v1alpha1_test

import (
	"context"
	"path"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

var _ = Describe("MeshFaultInjection", func() {
	type sidecarTestCase struct {
		resources         []*core_xds.Resource
		fromRules         core_rules.FromRules
		expectedListeners []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			context := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						AddInbound(
							builders.Inbound().
								WithTags(map[string]string{mesh_proto.ProtocolTag: "http"}).
								WithAddress("127.0.0.1").
								WithPort(17777).
								WithService("backend"),
						).
						AddInbound(
							builders.Inbound().
								WithTags(map[string]string{mesh_proto.ProtocolTag: "tcp"}).
								WithAddress("127.0.0.1").
								WithPort(17778).
								WithService("frontend"),
						),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshFaultInjectionType, given.fromRules)).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

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
					Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
					Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {
						{
							Subset: core_rules.Subset{
								{
									Key:   "kuma.io/service",
									Value: "demo-client",
								},
							},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: int32(444),
											Percentage: intstr.FromString("12"),
										},
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("55s"),
											Percentage: intstr.FromString("55"),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "111Mbps",
											Percentage: intstr.FromString("62.9"),
										},
									},
								},
							},
						},
						{
							Subset: core_rules.Subset{
								{
									Key:   "kuma.io/service",
									Value: "demo-client",
									Not:   true,
								},
							},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: 111,
											Percentage: intstr.FromInt32(11),
										},
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("22s"),
											Percentage: intstr.FromInt32(22),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "333Mbps",
											Percentage: intstr.FromString("33.3"),
										},
									},
								},
							},
						},
					},
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Http: &[]api.FaultInjectionConf{
								{
									Abort: &api.AbortConf{
										HttpStatus: int32(444),
										Percentage: intstr.FromString("12.1"),
									},
									Delay: &api.DelayConf{
										Value:      *test.ParseDuration("55s"),
										Percentage: intstr.FromInt(55),
									},
									ResponseBandwidth: &api.ResponseBandwidthConf{
										Limit:      "111Mbps",
										Percentage: intstr.FromString("62.9"),
									},
								},
							},
						},
					}},
				},
			},
			expectedListeners: []string{"basic_listener_1.golden.yaml", "basic_listener_2.golden.yaml"},
		}),
	)

	It("should generate proper Envoy config for Egress", func() {
		// given
		rs := core_xds.NewResourceSet()

		// listener that matches
		listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 10002, core_xds.SocketAddressProtocolTCP).
			WithOverwriteName("test_listener").
			Configure(
				listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-1_mesh-1").Configure(
					listeners.MatchTransportProtocol("tls"),
					listeners.MatchServerNames("external-service-1{mesh=mesh-1}"),
					listeners.HttpConnectionManager("external-service-1", false),
				)),
				listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-2_mesh-1").Configure(
					listeners.MatchTransportProtocol("tls"),
					listeners.MatchServerNames("external-service-2{mesh=mesh-1}"),
					listeners.TCPProxy("external-service-2"),
				)),
				listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-1_mesh-2").Configure(
					listeners.MatchTransportProtocol("tls"),
					listeners.MatchServerNames("external-service-1{mesh=mesh-2}"),
					listeners.TCPProxy("external-service-1"),
				)),
				listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-2_mesh-2").Configure(
					listeners.MatchTransportProtocol("tls"),
					listeners.MatchServerNames("external-service-2{mesh=mesh-2}"),
					listeners.HttpConnectionManager("external-service-2", false),
				)),
				listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "internal-service-1_mesh-1").Configure(
					listeners.MatchTransportProtocol("tls"),
					listeners.MatchServerNames("internal-service-1{mesh=mesh-1}"),
					listeners.TCPProxy("internal-service-1"),
				)),
			).
			Build()
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   egress.OriginEgress,
			Resource: listener,
		})

		// mesh with enabled mTLS and egress
		ctxMesh1 := xds_builders.Context().
			WithMeshBuilder(builders.Mesh().
				WithName("mesh-1").
				WithBuiltinMTLSBackend("builtin-1").
				WithEnabledMTLSBackend("builtin-1").
				WithEgressRoutingEnabled()).
			Build()

		proxy := &core_xds.Proxy{
			APIVersion: envoy.APIV3,
			ZoneEgressProxy: &core_xds.ZoneEgressProxy{
				ZoneEgressResource: &mesh.ZoneEgressResource{
					Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "mesh-1"},
					Spec: &mesh_proto.ZoneEgress{
						Networking: &mesh_proto.ZoneEgress_Networking{
							Address: "192.168.0.1",
							Port:    10002,
						},
					},
				},
				ZoneIngresses: []*mesh.ZoneIngressResource{},
				MeshResourcesList: []*core_xds.MeshResources{
					{
						Mesh: builders.Mesh().WithName("mesh-1").WithEnabledMTLSBackend("ca-1").WithBuiltinMTLSBackend("ca-1").Build(),
						ExternalServices: []*mesh.ExternalServiceResource{
							{
								Meta: &test_model.ResourceMeta{
									Mesh: "mesh-1",
									Name: "es-1",
								},
								Spec: &mesh_proto.ExternalService{
									Tags: map[string]string{
										"kuma.io/service":  "external-service-1",
										"kuma.io/protocol": "http",
									},
									Networking: &mesh_proto.ExternalService_Networking{
										Address: "externalservice-1.org",
									},
								},
							},
						},
						Dynamic: core_xds.ExternalServiceDynamicPolicies{
							"external-service-1": {
								api.MeshFaultInjectionType: core_xds.TypedMatchingPolicies{
									FromRules: core_rules.FromRules{
										Rules: map[core_rules.InboundListener]core_rules.Rules{
											{
												Address: "192.168.0.1", Port: 10002,
											}: {
												{
													Subset: core_rules.MeshService("frontend"),
													Conf: api.Conf{
														Http: &[]api.FaultInjectionConf{
															{
																Abort: &api.AbortConf{
																	HttpStatus: int32(444),
																	Percentage: intstr.FromInt(12),
																},
																Delay: &api.DelayConf{
																	Value:      *test.ParseDuration("55s"),
																	Percentage: intstr.FromString("55"),
																},
																ResponseBandwidth: &api.ResponseBandwidthConf{
																	Limit:      "111Mbps",
																	Percentage: intstr.FromString("62.9"),
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
					},
					{
						Mesh: builders.Mesh().WithName("mesh-2").WithEnabledMTLSBackend("ca-2").WithBuiltinMTLSBackend("ca-2").Build(),
						ExternalServices: []*mesh.ExternalServiceResource{
							{
								Meta: &test_model.ResourceMeta{
									Mesh: "mesh-2",
									Name: "es-1",
								},
								Spec: &mesh_proto.ExternalService{
									Tags: map[string]string{
										"kuma.io/service": "external-service-1",
									},
									Networking: &mesh_proto.ExternalService_Networking{
										Address: "externalservice-1.org",
									},
								},
							},
							{
								Meta: &test_model.ResourceMeta{
									Mesh: "mesh-2",
									Name: "es-2",
								},
								Spec: &mesh_proto.ExternalService{
									Tags: map[string]string{
										"kuma.io/service":  "external-service-2",
										"kuma.io/protocol": "http",
									},
									Networking: &mesh_proto.ExternalService_Networking{
										Address: "externalservice-2.org",
									},
								},
							},
						},
						Dynamic: core_xds.ExternalServiceDynamicPolicies{
							"external-service-1": {
								api.MeshFaultInjectionType: core_xds.TypedMatchingPolicies{
									FromRules: core_rules.FromRules{
										Rules: map[core_rules.InboundListener]core_rules.Rules{
											{
												Address: "192.168.0.1", Port: 10002,
											}: {
												{
													Subset: core_rules.MeshSubset(),
													Conf: api.Conf{
														Http: &[]api.FaultInjectionConf{
															{
																Abort: &api.AbortConf{
																	HttpStatus: int32(444),
																	Percentage: intstr.FromInt(12),
																},
																Delay: &api.DelayConf{
																	Value:      *test.ParseDuration("55s"),
																	Percentage: intstr.FromString("55"),
																},
																ResponseBandwidth: &api.ResponseBandwidthConf{
																	Limit:      "111Mbps",
																	Percentage: intstr.FromString("62.9"),
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
							"external-service-2": {
								api.MeshFaultInjectionType: core_xds.TypedMatchingPolicies{
									FromRules: core_rules.FromRules{
										Rules: map[core_rules.InboundListener]core_rules.Rules{
											{
												Address: "192.168.0.1", Port: 10002,
											}: {
												{
													Subset: core_rules.MeshSubset(),
													Conf: api.Conf{
														Http: &[]api.FaultInjectionConf{
															{
																Abort: &api.AbortConf{
																	HttpStatus: int32(444),
																	Percentage: intstr.FromInt(12),
																},
																Delay: &api.DelayConf{
																	Value:      *test.ParseDuration("55s"),
																	Percentage: intstr.FromString("55"),
																},
																ResponseBandwidth: &api.ResponseBandwidthConf{
																	Limit:      "111Mbps",
																	Percentage: intstr.FromString("62.9"),
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
					},
				},
			},
		}

		// when
		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		err = p.Apply(rs, *ctxMesh1, proxy)
		Expect(err).ToNot(HaveOccurred())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())
		bytes, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "egress_basic_listener.golden.yaml")))
	})

	It("should generate proper Envoy config for MeshGateway Dataplanes", func() {
		// given
		rules := core_rules.GatewayRules{
			ToRules: core_rules.GatewayToRules{
				ByListener: map[core_rules.InboundListener]core_rules.ToRules{
					{Address: "192.168.0.1", Port: 8080}: {
						Rules: core_rules.Rules{{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: int32(444),
											Percentage: intstr.FromInt(12),
										},
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("55s"),
											Percentage: intstr.FromString("55"),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "111Mbps",
											Percentage: intstr.FromString("62.9"),
										},
									},
								},
							},
						}},
					},
				},
			},
		}

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
			WithPolicies(xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshFaultInjectionType, rules)).
			Build()
		for n, p := range core_plugins.Plugins().ProxyPlugins() {
			Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
		}
		gatewayGenerator := gatewayGenerator()
		generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
		Expect(err).NotTo(HaveOccurred())

		// when
		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

		// then
		Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())
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
