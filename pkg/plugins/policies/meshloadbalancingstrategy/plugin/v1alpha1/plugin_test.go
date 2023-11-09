package v1alpha1_test

import (
	"context"
	"fmt"
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
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
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
		context   xds_context.Context
	}
	DescribeTable("Apply to sidecar Dataplanes",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resources, given.context, given.proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			Expect(getResource(resources, envoy_resource.ListenerType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
			Expect(getResource(resources, envoy_resource.ClusterType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
			Expect(getResource(resources, envoy_resource.EndpointType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
		},
		Entry("basic", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{}),
						createEndpointWith("zone-2", "192.168.1.2", map[string]string{}),
					}),
				},
				{
					Name:   "payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:     "backend",
					Origin:   generator.OriginOutbound,
					Resource: backendListener(),
				},
				{
					Name:     "payments",
					Origin:   generator.OriginOutbound,
					Resource: paymentsListener(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
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
							ToRules: core_rules.ToRules{
								Rules: []*core_rules.Rule{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
										},
									},
									{
										Subset: core_rules.MeshService("payment"),
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
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "mesh-1:eds-cluster").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "mesh-1:eds-cluster",
					Origin: egress.OriginEgress,
					Resource: endpoints.CreateClusterLoadAssignment("mesh-1:eds-cluster", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{}),
						createEndpointWith("zone-2", "192.168.1.2", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.3", map[string]string{}),
					}),
				},
				{
					Name:   "mesh-2:static-cluster",
					Origin: egress.OriginEgress,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "mesh-2:static-cluster").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					MeshResourcesList: []*core_xds.MeshResources{
						{
							Mesh: builders.Mesh().WithName("mesh-1").Build(),
							Dynamic: core_xds.ExternalServiceDynamicPolicies{
								"eds-cluster": {
									v1alpha1.MeshLoadBalancingStrategyType: core_xds.TypedMatchingPolicies{
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{}: {
													{Conf: v1alpha1.Conf{LocalityAwareness: &v1alpha1.LocalityAwareness{
														Disabled: pointer.To(false),
														CrossZone: &v1alpha1.CrossZone{
															Failover: []v1alpha1.Failover{
																{
																	From: &v1alpha1.FromZone{Zones: []string{"zone-1"}},
																	To: v1alpha1.ToZone{
																		Type:  v1alpha1.Only,
																		Zones: &[]string{"zone-2"},
																	},
																},
																{
																	From: &v1alpha1.FromZone{Zones: []string{"zone-1"}},
																	To: v1alpha1.ToZone{
																		Type:  v1alpha1.Only,
																		Zones: &[]string{"zone-3"},
																	},
																},
															},
														},
													}}},
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
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{}: {
													{Conf: v1alpha1.Conf{LocalityAwareness: &v1alpha1.LocalityAwareness{
														Disabled: pointer.To(false),
														CrossZone: &v1alpha1.CrossZone{
															Failover: []v1alpha1.Failover{
																{
																	To: v1alpha1.ToZone{
																		Type: v1alpha1.Any,
																	},
																},
															},
														},
													}}},
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
		Entry("egress_basic", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "mesh-1:eds-cluster",
					Origin: egress.OriginEgress,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "mesh-1:eds-cluster").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "mesh-1:eds-cluster",
					Origin: egress.OriginEgress,
					Resource: endpoints.CreateClusterLoadAssignment("mesh-1:eds-cluster", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{}),
						createEndpointWith("zone-2", "192.168.1.2", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.3", map[string]string{}),
					}),
				},
				{
					Name:   "mesh-2:static-cluster",
					Origin: egress.OriginEgress,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "mesh-2:static-cluster").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					MeshResourcesList: []*core_xds.MeshResources{
						{
							Mesh: builders.Mesh().WithName("mesh-1").Build(),
							Dynamic: core_xds.ExternalServiceDynamicPolicies{
								"eds-cluster": {
									v1alpha1.MeshLoadBalancingStrategyType: core_xds.TypedMatchingPolicies{
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{}: {
													{Conf: v1alpha1.Conf{LocalityAwareness: &v1alpha1.LocalityAwareness{
														Disabled: pointer.To(false),
													}}},
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
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{}: {
													{Conf: v1alpha1.Conf{LocalityAwareness: &v1alpha1.LocalityAwareness{
														Disabled: pointer.To(false),
													}}},
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
		Entry("locality_aware_basic", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-1", "192.168.1.3", map[string]string{"k8s.io/az": "test"}),
						createEndpointWith("zone-1", "192.168.1.4", map[string]string{"k8s.io/region": "test"}),
						createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:   "payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{"k8s.io/node": "node1"}),
							createEndpointWith("zone-1", "192.168.0.2", map[string]string{"k8s.io/node": "node2"}),
							createEndpointWith("zone-2", "192.168.0.3", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:     "backend",
					Origin:   generator.OriginOutbound,
					Resource: backendListener(),
				},
				{
					Name:     "payments",
					Origin:   generator.OriginOutbound,
					Resource: paymentsListener(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
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
							ToRules: core_rules.ToRules{
								Rules: []*core_rules.Rule{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												LocalZone: &v1alpha1.LocalZone{
													AffinityTags: &[]v1alpha1.AffinityTag{
														{
															Key:    "k8s.io/node",
															Weight: pointer.To[uint32](9000),
														},
														{
															Key:    "k8s.io/az",
															Weight: pointer.To[uint32](900),
														},
														{
															Key:    "k8s.io/region",
															Weight: pointer.To[uint32](90),
														},
													},
												},
												CrossZone: &v1alpha1.CrossZone{
													Failover: []v1alpha1.Failover{
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.AnyExcept,
																Zones: &[]string{"zone-3", "zone-4", "zone-5"},
															},
														},
														{
															From: &v1alpha1.FromZone{
																Zones: []string{"zone-1"},
															},
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-3"},
															},
														},
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-4"},
															},
														},
													},
												},
											},
										},
									},
									{
										Subset: core_rules.MeshService("payment"),
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
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												LocalZone: &v1alpha1.LocalZone{
													AffinityTags: &[]v1alpha1.AffinityTag{
														{
															Key:    "k8s.io/node",
															Weight: pointer.To[uint32](9000),
														},
														{
															Key:    "k8s.io/az",
															Weight: pointer.To[uint32](900),
														},
														{
															Key:    "k8s.io/region",
															Weight: pointer.To[uint32](90),
														},
													},
												},
												CrossZone: &v1alpha1.CrossZone{
													Failover: []v1alpha1.Failover{
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-2"},
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
		Entry("locality_aware_basic_egress_enabled", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-1", "192.168.1.3", map[string]string{"k8s.io/az": "test"}),
						createEndpointWith("zone-1", "192.168.1.4", map[string]string{"k8s.io/region": "test"}),
						createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
					}),
				},
				{
					Name:   "payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{"k8s.io/node": "node1"}),
							createEndpointWith("zone-1", "192.168.0.2", map[string]string{"k8s.io/node": "node2"}),
							createEndpointWith("zone-2", "192.168.0.5", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:     "backend",
					Origin:   generator.OriginOutbound,
					Resource: backendListener(),
				},
				{
					Name:     "payments",
					Origin:   generator.OriginOutbound,
					Resource: paymentsListener(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
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
							ToRules: core_rules.ToRules{
								Rules: []*core_rules.Rule{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												LocalZone: &v1alpha1.LocalZone{
													AffinityTags: &[]v1alpha1.AffinityTag{
														{
															Key:    "k8s.io/node",
															Weight: pointer.To[uint32](9000),
														},
														{
															Key:    "k8s.io/az",
															Weight: pointer.To[uint32](900),
														},
														{
															Key:    "k8s.io/region",
															Weight: pointer.To[uint32](90),
														},
													},
												},
												CrossZone: &v1alpha1.CrossZone{
													Failover: []v1alpha1.Failover{
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.AnyExcept,
																Zones: &[]string{"zone-3", "zone-4", "zone-5"},
															},
														},
														{
															From: &v1alpha1.FromZone{
																Zones: []string{"zone-1"},
															},
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-3"},
															},
														},
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-4"},
															},
														},
													},
												},
											},
										},
									},
									{
										Subset: core_rules.MeshService("payment"),
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
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												LocalZone: &v1alpha1.LocalZone{
													AffinityTags: &[]v1alpha1.AffinityTag{
														{
															Key:    "k8s.io/node",
															Weight: pointer.To[uint32](9000),
														},
														{
															Key:    "k8s.io/az",
															Weight: pointer.To[uint32](900),
														},
														{
															Key:    "k8s.io/region",
															Weight: pointer.To[uint32](90),
														},
													},
												},
												CrossZone: &v1alpha1.CrossZone{
													Failover: []v1alpha1.Failover{
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-2"},
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
			context: contextWithEgressEnabled(),
		}),
		Entry("locality_aware_no_cross_zone", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-1", "192.168.1.3", map[string]string{"k8s.io/az": "test"}),
						createEndpointWith("zone-1", "192.168.1.4", map[string]string{"k8s.io/region": "test"}),
						createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:   "payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:     "backend",
					Origin:   generator.OriginOutbound,
					Resource: backendListener(),
				},
				{
					Name:     "payments",
					Origin:   generator.OriginOutbound,
					Resource: paymentsListener(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
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
							ToRules: core_rules.ToRules{
								Rules: []*core_rules.Rule{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												LocalZone: &v1alpha1.LocalZone{
													AffinityTags: &[]v1alpha1.AffinityTag{
														{
															Key:    "k8s.io/node",
															Weight: pointer.To[uint32](9000),
														},
														{
															Key:    "k8s.io/az",
															Weight: pointer.To[uint32](900),
														},
														{
															Key:    "k8s.io/region",
															Weight: pointer.To[uint32](90),
														},
													},
												},
											},
										},
									},
									{
										Subset: core_rules.MeshService("payment"),
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
		Entry("locality_aware_cross_zone", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-1", "192.168.1.3", map[string]string{"k8s.io/az": "test"}),
						createEndpointWith("zone-1", "192.168.1.4", map[string]string{"k8s.io/region": "test"}),
						createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:   "payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:     "backend",
					Origin:   generator.OriginOutbound,
					Resource: backendListener(),
				},
				{
					Name:     "payments",
					Origin:   generator.OriginOutbound,
					Resource: paymentsListener(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
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
							ToRules: core_rules.ToRules{
								Rules: []*core_rules.Rule{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												CrossZone: &v1alpha1.CrossZone{
													Failover: []v1alpha1.Failover{
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.AnyExcept,
																Zones: &[]string{"zone-3", "zone-4", "zone-5"},
															},
														},
														{
															From: &v1alpha1.FromZone{
																Zones: []string{"zone-1"},
															},
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-3"},
															},
														},
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-4"},
															},
														},
													},
												},
											},
										},
									},
									{
										Subset: core_rules.MeshService("payment"),
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
		Entry("locality_aware_split", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "backend-bb38a94289f18fb9",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend-bb38a94289f18fb9").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend-c72efb5be46fae6b",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend-c72efb5be46fae6b").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:   "backend-bb38a94289f18fb9",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend-bb38a94289f18fb9", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-1", "192.168.1.3", map[string]string{"k8s.io/az": "test"}),
						createEndpointWith("zone-1", "192.168.1.4", map[string]string{"k8s.io/region": "test"}),
						createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:   "backend-c72efb5be46fae6b",
					Origin: generator.OriginOutbound,
					Resource: endpoints.CreateClusterLoadAssignment("backend-c72efb5be46fae6b", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:   "payment",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:   "backend",
					Origin: generator.OriginOutbound,
					Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:27777", false)).
							Configure(
								HttpOutboundRoute(
									"backend",
									envoy_common.Routes{{
										Clusters: []envoy_common.Cluster{
											envoy_common.NewCluster(
												envoy_common.WithService("backend-bb38a94289f18fb9"),
												envoy_common.WithWeight(90),
											),
											envoy_common.NewCluster(
												envoy_common.WithService("backend-c72efb5be46fae6b"),
												envoy_common.WithWeight(10),
											),
										},
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
					Name:     "payments",
					Origin:   generator.OriginOutbound,
					Resource: paymentsListener(),
				},
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
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
							ToRules: core_rules.ToRules{
								Rules: []*core_rules.Rule{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: v1alpha1.Conf{
											LoadBalancer: &v1alpha1.LoadBalancer{
												Type: v1alpha1.RandomType,
											},
											LocalityAwareness: &v1alpha1.LocalityAwareness{
												LocalZone: &v1alpha1.LocalZone{
													AffinityTags: &[]v1alpha1.AffinityTag{
														{
															Key: "k8s.io/node",
														},
														{
															Key: "k8s.io/az",
														},
														{
															Key: "k8s.io/region",
														},
													},
												},
												CrossZone: &v1alpha1.CrossZone{
													Failover: []v1alpha1.Failover{
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.AnyExcept,
																Zones: &[]string{"zone-3", "zone-4", "zone-5"},
															},
														},
														{
															From: &v1alpha1.FromZone{
																Zones: []string{"zone-1"},
															},
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-3"},
															},
														},
														{
															To: v1alpha1.ToZone{
																Type:  v1alpha1.Only,
																Zones: &[]string{"zone-4"},
															},
														},
													},
												},
											},
										},
									},
									{
										Subset: core_rules.MeshService("payment"),
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
	)
	type gatewayTestCase struct {
		name        string
		endpointMap core_xds.EndpointMap
		toRules     core_rules.ToRules
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

			xdsCtx := test_xds.CreateSampleMeshContextWithEndpoints(resources, given.endpointMap)
			proxy := core_xds.Proxy{
				APIVersion: "v3",
				Dataplane: builders.Dataplane().
					WithName("sample-gateway").
					WithAddress("192.168.0.1").
					WithBuiltInGateway("sample-gateway").
					AddBuiltInGatewayTags(map[string]string{
						"k8s.io/node":   "node1",
						"k8s.io/az":     "test",
						"k8s.io/region": "test",
					}).
					Build(),
				Policies: core_xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						v1alpha1.MeshLoadBalancingStrategyType: {
							Type:    v1alpha1.MeshLoadBalancingStrategyType,
							ToRules: given.toRules,
						},
					},
				},
				RuntimeExtensions: map[string]interface{}{},
				Zone:              "test-zone",
			}
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, &proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), xdsCtx, &proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, &proxy)).To(Succeed())

			getResourceYaml := func(list core_xds.ResourceList) []byte {
				actualResource, err := util_proto.ToYAML(list[0].Resource)
				Expect(err).ToNot(HaveOccurred())
				return actualResource
			}

			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_cluster.golden.yaml", given.name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.EndpointType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_endpoint.golden.yaml", given.name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_listener.golden.yaml", given.name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_route.golden.yaml", given.name))))
		},
		Entry("basic outbound cluster", gatewayTestCase{
			name: "basic",
			endpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {
					{
						Tags: map[string]string{
							"app": "backend",
						},
					},
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
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
		Entry("locality aware gateway", gatewayTestCase{
			name: "locality_aware",
			endpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {
					createEndpointWith("test-zone", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
					createEndpointWith("test-zone", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
					createEndpointWith("test-zone", "192.168.1.3", map[string]string{"k8s.io/az": "test"}),
					createEndpointWith("test-zone", "192.168.1.4", map[string]string{"k8s.io/region": "test"}),
					createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
					createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
					createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
					createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: v1alpha1.Conf{
							LocalityAwareness: &v1alpha1.LocalityAwareness{
								LocalZone: &v1alpha1.LocalZone{
									AffinityTags: &[]v1alpha1.AffinityTag{
										{
											Key:    "k8s.io/node",
											Weight: pointer.To[uint32](9000),
										},
										{
											Key:    "k8s.io/az",
											Weight: pointer.To[uint32](900),
										},
										{
											Key:    "k8s.io/region",
											Weight: pointer.To[uint32](90),
										},
									},
								},
								CrossZone: &v1alpha1.CrossZone{
									Failover: []v1alpha1.Failover{
										{
											To: v1alpha1.ToZone{
												Type:  v1alpha1.AnyExcept,
												Zones: &[]string{"zone-3", "zone-4", "zone-5"},
											},
										},
										{
											From: &v1alpha1.FromZone{
												Zones: []string{"zone-1"},
											},
											To: v1alpha1.ToZone{
												Type:  v1alpha1.Only,
												Zones: &[]string{"zone-3"},
											},
										},
										{
											To: v1alpha1.ToZone{
												Type:  v1alpha1.Only,
												Zones: &[]string{"zone-4"},
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
})

func createEndpointWith(zone string, ip string, extraTags map[string]string) core_xds.Endpoint {
	tags := map[string]string{
		mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
		mesh_proto.ZoneTag:     zone,
	}

	for k, v := range extraTags {
		tags[k] = v
	}

	return core_xds.Endpoint{
		Target:   ip,
		Port:     8080,
		Tags:     tags,
		Locality: &core_xds.Locality{Zone: zone, Priority: 0},
	}
}

func paymentsListener() envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27778, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
		)).MustBuild()
}

func backendListener() envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
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
		)).MustBuild()
}

func contextWithEgressEnabled() xds_context.Context {
	return xds_context.Context{
		Mesh: xds_context.MeshContext{
			Resource: &core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Routing: &mesh_proto.Routing{ZoneEgress: true},
				},
			},
		},
	}
}
