package v1alpha1_test

import (
	"context"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
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
											Limit:      "111mbps",
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
											Limit:      "333mbps",
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
										Limit:      "111mbps",
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

	It("should generate proper Envoy config for MeshGateway Dataplanes", func() {
		// given
		rules := core_rules.GatewayRules{
			ToRules: map[core_rules.InboundListener]core_rules.Rules{
				{Address: "192.168.0.1", Port: 8080}: {{
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
									Limit:      "111mbps",
									Percentage: intstr.FromString("62.9"),
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
