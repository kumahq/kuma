package v1alpha1_test

import (
	"fmt"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

func getResource(resourceSet *core_xds.ResourceSet, typ envoy_resource.Type) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshLoadBalancingStrategy", func() {
	type testCase struct {
		resources []core_xds.Resource
		proxy     *core_xds.Proxy
	}
	DescribeTable("Apply to sidecar Dataplanes",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			context := xds_context.Context{}

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resources, context, given.proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			Expect(getResource(resources, envoy_resource.ListenerType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
			Expect(getResource(resources, envoy_resource.ClusterType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
			Expect(getResource(resources, envoy_resource.EndpointType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
		},
		Entry("basic", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "cluster-backend",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
						Configure(clusters.EdsCluster("backend")).
						MustBuild(),
				},
				{
					Name:   "cluster-backend",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						{
							Target: "192.168.1.1",
							Port:   8080,
							Tags: map[string]string{
								mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
								mesh_proto.ZoneTag:     "zone-1",
							},
							Locality: &core_xds.Locality{Zone: "zone-1", Priority: 0},
						},
						{
							Target: "192.168.1.2",
							Port:   8080,
							Tags: map[string]string{
								mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
								mesh_proto.ZoneTag:     "zone-2",
							},
							Locality: &core_xds.Locality{Zone: "zone-2", Priority: 0},
						},
					}),
				},
				{
					Name:   "cluster-payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
						Configure(clusters.ProvidedEndpointCluster(
							"payment",
							false,
							core_xds.Endpoint{
								Target: "192.168.0.1",
								Port:   8080,
								Tags: map[string]string{
									mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
									mesh_proto.ZoneTag:     "zone-1",
								},
								Locality: &core_xds.Locality{Zone: "zone-1", Priority: 0},
							},
							core_xds.Endpoint{
								Target: "192.168.0.2",
								Port:   8080,
								Tags: map[string]string{
									mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
									mesh_proto.ZoneTag:     "zone-2",
								},
								Locality: &core_xds.Locality{Zone: "zone-2", Priority: 0},
							},
						)).MustBuild(),
				},
				{
					Name:   "listener-backend",
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
											"backend": true,
										},
									},
								),
							),
						)).MustBuild(),
				},
				{
					Name:   "listener-payments",
					Origin: generator.OriginOutbound,
					Resource: NewListenerBuilder(envoy_common.APIV3).
						Configure(OutboundListener("outbound:127.0.0.1:27778", "127.0.0.1", 27778, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
							Configure(HttpConnectionManager("127.0.0.1:27778", false)).
							Configure(
								HttpOutboundRoute(
									"backend",
									envoy_common.Routes{{
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("payment"),
											envoy_common.WithWeight(100),
										)},
									}},
									map[string]map[string]bool{
										"kuma.io/service": {
											"payment": true,
										},
									},
								),
							),
						)).MustBuild(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
					}).
					AddOutbound(
						builders.Outbound().WithAddress("127.0.0.1").WithPort(27777).WithTags(map[string]string{
							mesh_proto.ServiceTag:  "backend",
							mesh_proto.ProtocolTag: "http",
						}),
					).
					AddOutbound(
						builders.Outbound().WithAddress("127.0.0.1").WithPort(27778).WithTags(map[string]string{
							mesh_proto.ServiceTag:  "payment",
							mesh_proto.ProtocolTag: "http",
						}),
					).
					Build(),
				Policies: core_xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						v1alpha1.MeshLoadBalancingStrategyType: {
							Type: v1alpha1.MeshLoadBalancingStrategyType,
							ToRules: core_xds.ToRules{
								Rules: []*rules.Rule{
									{
										Subset: rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
										},
									},
									{
										Subset: rules.MeshService("payment"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RingHashType,
												RingHash: &v1alpha1.RingHash{
													MinRingSize:  pointer.To[uint32](100),
													MaxRingSize:  pointer.To[uint32](1000),
													HashFunction: pointer.To(v1alpha1.MurmurHash2Type),
													HashPolicies: &[]v1alpha1.HashPolicy{
														{
															Type: v1alpha1.QueryParameterType,
															QueryParameter: &v1alpha1.QueryParameter{
																Name: "queryparam",
															},
															Terminal: pointer.To(true),
														},
														{
															Type: v1alpha1.ConnectionType,
															Connection: &v1alpha1.Connection{
																SourceIP: pointer.To(true),
															},
															Terminal: pointer.To(false),
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
				Routing: core_xds.Routing{
					OutboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
						"backend": {
							{
								Tags: map[string]string{mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP},
							},
						},
						"payment": {
							{
								Tags: map[string]string{mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP},
							},
						},
					},
				},
			},
		}),
		Entry("egress", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "mesh-1:eds-cluster",
					Origin: egress.OriginEgress,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
						Configure(clusters.EdsCluster("mesh-1:eds-cluster")).
						MustBuild(),
				},
				{
					Name:   "mesh-1:cla-for-eds-cluster",
					Origin: egress.OriginEgress,
					Resource: endpoints.CreateClusterLoadAssignment("mesh-1:eds-cluster", []core_xds.Endpoint{
						{
							Target: "192.168.1.1",
							Port:   8080,
							Tags: map[string]string{
								mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
								mesh_proto.ZoneTag:     "zone-1",
							},
							Locality: &core_xds.Locality{Zone: "zone-1", Priority: 0},
						},
						{
							Target: "192.168.1.2",
							Port:   8080,
							Tags: map[string]string{
								mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
								mesh_proto.ZoneTag:     "zone-2",
							},
							Locality: &core_xds.Locality{Zone: "zone-2", Priority: 0},
						},
					}),
				},
				{
					Name:   "mesh-2:static-cluster",
					Origin: egress.OriginEgress,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
						Configure(clusters.ProvidedEndpointCluster(
							"mesh-2:static-cluster",
							false,
							core_xds.Endpoint{
								Target: "192.168.0.1",
								Port:   8080,
								Tags: map[string]string{
									mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
									mesh_proto.ZoneTag:     "zone-1",
								},
								Locality: &core_xds.Locality{Zone: "zone-1", Priority: 0},
							},
							core_xds.Endpoint{
								Target: "192.168.0.2",
								Port:   8080,
								Tags: map[string]string{
									mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
									mesh_proto.ZoneTag:     "zone-2",
								},
								Locality: &core_xds.Locality{Zone: "zone-2", Priority: 0},
							},
						)).MustBuild(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					ZoneEgressResource: &core_mesh.ZoneEgressResource{
						Spec: &mesh_proto.ZoneEgress{Zone: "zone-1"},
					},
					MeshResourcesList: []*core_xds.MeshResources{
						{
							Mesh: builders.Mesh().WithName("mesh-1").Build(),
							Dynamic: core_xds.ExternalServiceDynamicPolicies{
								"eds-cluster": {
									v1alpha1.MeshLoadBalancingStrategyType: core_xds.TypedMatchingPolicies{
										FromRules: core_xds.FromRules{
											Rules: map[core_xds.InboundListener]rules.Rules{
												{}: {
													{Conf: v1alpha1.Conf{LocalityAwareness: &v1alpha1.LocalityAwareness{Disabled: pointer.To(false)}}},
												},
											},
										},
									},
								},
							},
						},
						{
							Mesh: builders.Mesh().WithName("mesh-2").Build(),
							Dynamic: core_xds.ExternalServiceDynamicPolicies{
								"static-cluster": {
									v1alpha1.MeshLoadBalancingStrategyType: core_xds.TypedMatchingPolicies{
										FromRules: core_xds.FromRules{
											Rules: map[core_xds.InboundListener]rules.Rules{
												{}: {
													{Conf: v1alpha1.Conf{LocalityAwareness: &v1alpha1.LocalityAwareness{Disabled: pointer.To(false)}}},
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
		}),
	)
	type gatewayTestCase struct {
		name    string
		toRules core_xds.ToRules
	}
	DescribeTable("should generate proper Envoy config for MeshGateways",
		func(given gatewayTestCase) {
			Expect(given.name).ToNot(BeEmpty())
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
			}

			context := test_xds.CreateSampleMeshContextWith(resources)
			proxy := core_xds.Proxy{
				APIVersion: "v3",
				Dataplane:  samples.GatewayDataplane(),
				Policies: core_xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						v1alpha1.MeshLoadBalancingStrategyType: {
							Type:    v1alpha1.MeshLoadBalancingStrategyType,
							ToRules: given.toRules,
						},
					},
				},
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context, &proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, context, &proxy)).To(Succeed())

			getResourceYaml := func(list core_xds.ResourceList) []byte {
				actualResource, err := util_proto.ToYAML(list[0].Resource)
				Expect(err).ToNot(HaveOccurred())
				return actualResource
			}

			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_cluster.golden.yaml", given.name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_listener.golden.yaml", given.name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_route.golden.yaml", given.name))))
		},
		Entry("basic outbound cluster", gatewayTestCase{
			name: "basic",
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{},
						Conf: v1alpha1.Conf{
							LoadBalancer: &v1alpha1.LoadBalancer{
								Type: v1alpha1.RingHashType,
								RingHash: &v1alpha1.RingHash{
									MinRingSize:  pointer.To[uint32](100),
									MaxRingSize:  pointer.To[uint32](1000),
									HashFunction: pointer.To(v1alpha1.MurmurHash2Type),
									HashPolicies: &[]v1alpha1.HashPolicy{
										{
											Type: v1alpha1.QueryParameterType,
											QueryParameter: &v1alpha1.QueryParameter{
												Name: "queryparam",
											},
											Terminal: pointer.To(true),
										},
										{
											Type: v1alpha1.ConnectionType,
											Connection: &v1alpha1.Connection{
												SourceIP: pointer.To(true),
											},
											Terminal: pointer.To(false),
										},
									},
								},
							},
						},
					},
				},
			},
		}),
	)
})
