package v1alpha1_test

import (
	"path/filepath"
	"strconv"
	"strings"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	meshroute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshloadbalancingstrategy/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v3/pkg/test/xds/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/endpoints/v3"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	gateway_metadata "github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var _ = Describe("MeshLoadBalancingStrategy", func() {
	backendKRI := kri.MustFromString("kri_msvc_default_zone-1_ns-1_backend_")
	paymentKRI := kri.MustFromString("kri_msvc_default_zone-1_ns-1_payment_")
	frontendKRI := kri.MustFromString("kri_msvc_default_zone-1_ns-1_frontend_")

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

			resource, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
			resource, err = util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
			resource, err = util_yaml.GetResourcesToYaml(resources, envoy_resource.EndpointType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
		},
		Entry("basic", testCase{
			resources: []core_xds.Resource{
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{}),
						createEndpointWith("zone-2", "192.168.1.2", map[string]string{}),
					}),
				},
				{
					Name:           "payment",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "frontend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: frontendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "frontend").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.2.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.2.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource:       backendListener(),
				},
				{
					Name:           "payments",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource:       paymentsListener(),
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
					Build(),
				Policies: *xds_builders.MatchedPolicies().
					WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
						ResourceRules: outbound.ResourceRules{
							backendKRI: {Conf: []any{api.Conf{
								LoadBalancer: &api.LoadBalancer{
									Type: api.RandomType,
								},
							}}},
							frontendKRI: {Conf: []any{api.Conf{
								LoadBalancer: &api.LoadBalancer{
									Type: api.LeastRequestType,
									LeastRequest: &api.LeastRequest{
										ActiveRequestBias: &intstr.IntOrString{Type: intstr.String, StrVal: "10.1"},
									},
								},
							}}},
							paymentKRI: {Conf: []any{api.Conf{
								HashPolicies: &[]api.HashPolicy{
									{
										Type: api.QueryParameterType,
										QueryParameter: &api.QueryParameter{
											Name: "queryparam",
										},
										Terminal: pointer.To(true),
									},
									{
										Type: api.ConnectionType,
										Connection: &api.Connection{
											SourceIP: pointer.To(true),
										},
										Terminal: pointer.To(false),
									},
								},
								LoadBalancer: &api.LoadBalancer{
									Type: api.RingHashType,
									RingHash: &api.RingHash{
										MinRingSize:  pointer.To[uint32](100),
										MaxRingSize:  pointer.To[uint32](1000),
										HashFunction: pointer.To(api.MurmurHash2Type),
									},
								},
							}}},
						},
					}).
					Build(),
				Routing: *paymentsAndBackendRouting().Build(),
			},
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
		Entry("locality_aware_basic", testCase{
			resources: []core_xds.Resource{
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
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
					Name:           "payment",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{"k8s.io/node": "node1"}),
							createEndpointWith("zone-1", "192.168.0.2", map[string]string{"k8s.io/node": "node2"}),
							createEndpointWith("zone-2", "192.168.0.3", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource:       backendListener(),
				},
				{
					Name:           "payments",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource:       paymentsListener(),
				},
			},
			proxy: xds_builders.Proxy().
				WithZone("zone-1").
				WithDataplane(
					builders.Dataplane().
						AddInboundOfTagsMap(map[string]string{
							mesh_proto.ServiceTag: "backend",
							mesh_proto.ZoneTag:    "zone-1",
							"k8s.io/node":         "node1",
							"k8s.io/az":           "test",
							"k8s.io/region":       "test",
						}),
				).
				WithRouting(paymentsAndBackendRouting()).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
							ResourceRules: outbound.ResourceRules{
								backendKRI: {Conf: []any{api.Conf{
									LoadBalancer: &api.LoadBalancer{
										Type: api.RandomType,
									},
									LocalityAwareness: &api.LocalityAwareness{
										LocalZone: &api.LocalZone{
											AffinityTags: &[]api.AffinityTag{
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
										CrossZone: &api.CrossZone{
											FailoverThreshold: &api.FailoverThreshold{Percentage: intstr.FromString("99")},
											Failover: &[]api.Failover{
												{
													To: api.ToZone{
														Type:  api.AnyExcept,
														Zones: &[]string{"zone-3", "zone-4", "zone-5"},
													},
												},
												{
													From: &api.FromZone{
														Zones: []string{"zone-1"},
													},
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-3"},
													},
												},
												{
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-4"},
													},
												},
											},
										},
									},
								}}},
								paymentKRI: {Conf: []any{api.Conf{
									HashPolicies: &[]api.HashPolicy{
										{
											Type: api.QueryParameterType,
											QueryParameter: &api.QueryParameter{
												Name: "queryparam",
											},
											Terminal: pointer.To(true),
										},
										{
											Type: api.ConnectionType,
											Connection: &api.Connection{
												SourceIP: pointer.To(true),
											},
											Terminal: pointer.To(false),
										},
									},
									LoadBalancer: &api.LoadBalancer{
										Type: api.RingHashType,
										RingHash: &api.RingHash{
											MinRingSize:  pointer.To[uint32](100),
											MaxRingSize:  pointer.To[uint32](1000),
											HashFunction: pointer.To(api.MurmurHash2Type),
										},
									},
									LocalityAwareness: &api.LocalityAwareness{
										LocalZone: &api.LocalZone{
											AffinityTags: &[]api.AffinityTag{
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
										CrossZone: &api.CrossZone{
											Failover: &[]api.Failover{
												{
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-2"},
													},
												},
											},
										},
									},
								}}},
							},
						}),
				).
				Build(),
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
		Entry("locality_aware_no_cross_zone", testCase{
			resources: []core_xds.Resource{
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
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
					Name:           "payment",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource:       backendListener(),
				},
				{
					Name:           "payments",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource:       paymentsListener(),
				},
			},
			proxy: xds_builders.Proxy().
				WithZone("zone-1").
				WithDataplane(builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
					}),
				).
				WithRouting(paymentsAndBackendRouting()).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
							ResourceRules: outbound.ResourceRules{
								backendKRI: {Conf: []any{api.Conf{
									LoadBalancer: &api.LoadBalancer{
										Type: api.RandomType,
									},
									LocalityAwareness: &api.LocalityAwareness{
										LocalZone: &api.LocalZone{
											AffinityTags: &[]api.AffinityTag{
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
								}}},
								paymentKRI: {Conf: []any{api.Conf{
									HashPolicies: &[]api.HashPolicy{
										{
											Type: api.QueryParameterType,
											QueryParameter: &api.QueryParameter{
												Name: "queryparam",
											},
											Terminal: pointer.To(true),
										},
										{
											Type: api.ConnectionType,
											Connection: &api.Connection{
												SourceIP: pointer.To(true),
											},
											Terminal: pointer.To(false),
										},
									},
									LoadBalancer: &api.LoadBalancer{
										Type: api.RingHashType,
										RingHash: &api.RingHash{
											MinRingSize:  pointer.To[uint32](100),
											MaxRingSize:  pointer.To[uint32](1000),
											HashFunction: pointer.To(api.MurmurHash2Type),
										},
									},
								}}},
							},
						}),
				).
				Build(),
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
		Entry("locality_aware_cross_zone", testCase{
			resources: []core_xds.Resource{
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
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
					Name:           "payment",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource:       backendListener(),
				},
				{
					Name:           "payments",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource:       paymentsListener(),
				},
			},
			proxy: xds_builders.Proxy().
				WithZone("zone-1").
				WithDataplane(builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
					}),
				).
				WithRouting(paymentsAndBackendRouting()).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
							ResourceRules: outbound.ResourceRules{
								backendKRI: {Conf: []any{api.Conf{
									LoadBalancer: &api.LoadBalancer{
										Type: api.RandomType,
									},
									LocalityAwareness: &api.LocalityAwareness{
										CrossZone: &api.CrossZone{
											Failover: &[]api.Failover{
												{
													To: api.ToZone{
														Type:  api.AnyExcept,
														Zones: &[]string{"zone-3", "zone-4", "zone-5"},
													},
												},
												{
													From: &api.FromZone{
														Zones: []string{"zone-1"},
													},
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-3"},
													},
												},
												{
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-4"},
													},
												},
											},
										},
									},
								}}},
								paymentKRI: {Conf: []any{api.Conf{
									HashPolicies: &[]api.HashPolicy{
										{
											Type: api.QueryParameterType,
											QueryParameter: &api.QueryParameter{
												Name: "queryparam",
											},
											Terminal: pointer.To(true),
										},
										{
											Type: api.ConnectionType,
											Connection: &api.Connection{
												SourceIP: pointer.To(true),
											},
											Terminal: pointer.To(false),
										},
									},
									LoadBalancer: &api.LoadBalancer{
										Type: api.RingHashType,
										RingHash: &api.RingHash{
											MinRingSize:  pointer.To[uint32](100),
											MaxRingSize:  pointer.To[uint32](1000),
											HashFunction: pointer.To(api.MurmurHash2Type),
										},
									},
								}}},
							},
						}),
				).
				Build(),
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
		Entry("locality_aware_split", testCase{
			resources: []core_xds.Resource{
				{
					Name:           "backend-bb38a94289f18fb9",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend-bb38a94289f18fb9").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend-c72efb5be46fae6b",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend-c72efb5be46fae6b").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend-bb38a94289f18fb9",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
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
					Name:           "backend-c72efb5be46fae6b",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: endpoints.CreateClusterLoadAssignment("backend-c72efb5be46fae6b", []core_xds.Endpoint{
						createEndpointWith("zone-1", "192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWith("zone-1", "192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:           "payment",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{}),
							createEndpointWith("zone-2", "192.168.0.2", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:27777", false, nil, true)).
							Configure(AddFilterChainConfigurer(outboundRoute(
								"backend",
								xds.NewSplitBuilder().WithClusterName("backend-bb38a94289f18fb9").WithWeight(90).Build(),
								xds.NewSplitBuilder().WithClusterName("backend-c72efb5be46fae6b").WithWeight(10).Build(),
							))),
						)).MustBuild(),
				},
				{
					Name:           "payments",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource:       paymentsListener(),
				},
			},
			proxy: xds_builders.Proxy().
				WithZone("zone-1").
				WithDataplane(builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
						"k8s.io/node":         "node1",
						"k8s.io/az":           "test",
						"k8s.io/region":       "test",
					}),
				).
				WithRouting(paymentsAndBackendRouting()).
				WithPolicies(
					xds_builders.MatchedPolicies().WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
						ResourceRules: outbound.ResourceRules{
							backendKRI: {Conf: []any{api.Conf{
								LoadBalancer: &api.LoadBalancer{
									Type: api.RandomType,
								},
								LocalityAwareness: &api.LocalityAwareness{
									LocalZone: &api.LocalZone{
										AffinityTags: &[]api.AffinityTag{
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
									CrossZone: &api.CrossZone{
										Failover: &[]api.Failover{
											{
												To: api.ToZone{
													Type:  api.AnyExcept,
													Zones: &[]string{"zone-3", "zone-4", "zone-5"},
												},
											},
											{
												From: &api.FromZone{
													Zones: []string{"zone-1"},
												},
												To: api.ToZone{
													Type:  api.Only,
													Zones: &[]string{"zone-3"},
												},
											},
											{
												To: api.ToZone{
													Type:  api.Only,
													Zones: &[]string{"zone-4"},
												},
											},
										},
									},
								},
							}}},
							paymentKRI: {Conf: []any{api.Conf{
								HashPolicies: &[]api.HashPolicy{
									{
										Type: api.QueryParameterType,
										QueryParameter: &api.QueryParameter{
											Name: "queryparam",
										},
										Terminal: pointer.To(true),
									},
									{
										Type: api.ConnectionType,
										Connection: &api.Connection{
											SourceIP: pointer.To(true),
										},
										Terminal: pointer.To(false),
									},
								},
								LoadBalancer: &api.LoadBalancer{
									Type: api.RingHashType,
									RingHash: &api.RingHash{
										MinRingSize:  pointer.To[uint32](100),
										MaxRingSize:  pointer.To[uint32](1000),
										HashFunction: pointer.To(api.MurmurHash2Type),
									},
								},
							}}},
						},
					}),
				).
				Build(),
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
		Entry("route", testCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(kri.MustFromString("kri_msvc_default_zone-1_ns-1_ms-1_"), 27777, []meshhttproute_xds.OutboundRoute{
					{
						Name: kri.MustFromString("kri_mhttpr_default_zone-1_ns-1_route-1_").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-1"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(kri.WithSectionName(kri.MustFromString("kri_msvc_default_zone-1_ns-1_ms-1_"), uint32(27777)).String()).Build(),
						},
					},
					{
						Name: kri.MustFromString("kri_mhttpr_default_zone-1_ns-1_route-2_").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-2"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(kri.WithSectionName(kri.MustFromString("kri_msvc_default_zone-1_ns-1_ms-1_"), uint32(27777)).String()).Build(),
						},
					},
					{
						Name: kri.MustFromString("kri_mhttpr_default_zone-1_ns-1_route-3_").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-3"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(kri.WithSectionName(kri.MustFromString("kri_msvc_default_zone-1_ns-1_ms-1_"), uint32(27777)).String()).Build(),
						},
					},
				}),
			},
			proxy: &core_xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Zone:       "zone-1",
				Dataplane: builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "zone-1",
					}).
					Build(),
				Policies: *xds_builders.MatchedPolicies().
					WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
						ResourceRules: outbound.ResourceRules{
							kri.MustFromString("kri_msvc_default_zone-1_ns-1_ms-1_"): outbound.ResourceRule{
								Conf: []any{
									api.Conf{
										HashPolicies: &[]api.HashPolicy{
											{
												Type: api.HeaderType,
												Header: &api.Header{
													Name: "x-per-meshservice-header",
												},
											},
										},
									},
								},
							},
							kri.MustFromString("kri_mhttpr_default_zone-1_ns-1_route-3_"): outbound.ResourceRule{
								Conf: []any{
									api.Conf{
										HashPolicies: &[]api.HashPolicy{
											{
												Type: api.HeaderType,
												Header: &api.Header{
													Name: "x-per-meshhttproute-route-3",
												},
											},
										},
									},
								},
							},
						},
					}).
					Build(),
			},
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
		Entry("locality_aware_tag_free_endpoints", testCase{
			resources: []core_xds.Resource{
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
						Configure(clusters.EdsCluster()).
						MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource: endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
						createEndpointWithLabels("192.168.1.1", map[string]string{"k8s.io/node": "node1"}),
						createEndpointWithLabels("192.168.1.2", map[string]string{"k8s.io/node": "node2"}),
						createEndpointWithLabels("192.168.1.3", map[string]string{"k8s.io/az": "test"}),
						createEndpointWithLabels("192.168.1.4", map[string]string{"k8s.io/region": "test"}),
						createEndpointWith("zone-2", "192.168.1.5", map[string]string{}),
						createEndpointWith("zone-3", "192.168.1.6", map[string]string{}),
						createEndpointWith("zone-4", "192.168.1.7", map[string]string{}),
						createEndpointWith("zone-5", "192.168.1.8", map[string]string{}),
					}),
				},
				{
					Name:           "payment",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "payment").
						Configure(clusters.ProvidedEndpointCluster(
							false,
							createEndpointWith("zone-1", "192.168.0.1", map[string]string{"k8s.io/node": "node1"}),
							createEndpointWith("zone-1", "192.168.0.2", map[string]string{"k8s.io/node": "node2"}),
							createEndpointWith("zone-2", "192.168.0.3", map[string]string{}),
						)).MustBuild(),
				},
				{
					Name:           "backend",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: backendKRI,
					Resource:       backendListener(),
				},
				{
					Name:           "payments",
					Origin:         metadata.OriginOutbound,
					ResourceOrigin: paymentKRI,
					Resource:       paymentsListener(),
				},
			},
			proxy: xds_builders.Proxy().
				WithZone("zone-1").
				WithDataplane(
					builders.Dataplane().
						WithLabels(map[string]string{
							"k8s.io/node":   "node1",
							"k8s.io/az":     "test",
							"k8s.io/region": "test",
						}),
				).
				WithRouting(paymentsAndBackendRouting()).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithToPolicy(api.MeshLoadBalancingStrategyType, core_rules.ToRules{
							ResourceRules: outbound.ResourceRules{
								backendKRI: {Conf: []any{api.Conf{
									LoadBalancer: &api.LoadBalancer{
										Type: api.RandomType,
									},
									LocalityAwareness: &api.LocalityAwareness{
										LocalZone: &api.LocalZone{
											AffinityTags: &[]api.AffinityTag{
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
										CrossZone: &api.CrossZone{
											FailoverThreshold: &api.FailoverThreshold{Percentage: intstr.FromString("99")},
											Failover: &[]api.Failover{
												{
													To: api.ToZone{
														Type:  api.AnyExcept,
														Zones: &[]string{"zone-3", "zone-4", "zone-5"},
													},
												},
												{
													From: &api.FromZone{
														Zones: []string{"zone-1"},
													},
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-3"},
													},
												},
												{
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-4"},
													},
												},
											},
										},
									},
								}}},
								paymentKRI: {Conf: []any{api.Conf{
									HashPolicies: &[]api.HashPolicy{
										{
											Type: api.QueryParameterType,
											QueryParameter: &api.QueryParameter{
												Name: "queryparam",
											},
											Terminal: pointer.To(true),
										},
										{
											Type: api.ConnectionType,
											Connection: &api.Connection{
												SourceIP: pointer.To(true),
											},
											Terminal: pointer.To(false),
										},
									},
									LoadBalancer: &api.LoadBalancer{
										Type: api.RingHashType,
										RingHash: &api.RingHash{
											MinRingSize:  pointer.To[uint32](100),
											MaxRingSize:  pointer.To[uint32](1000),
											HashFunction: pointer.To(api.MurmurHash2Type),
										},
									},
									LocalityAwareness: &api.LocalityAwareness{
										LocalZone: &api.LocalZone{
											AffinityTags: &[]api.AffinityTag{
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
										CrossZone: &api.CrossZone{
											Failover: &[]api.Failover{
												{
													To: api.ToZone{
														Type:  api.Only,
														Zones: &[]string{"zone-2"},
													},
												},
											},
										},
									},
								}}},
							},
						}),
				).
				Build(),
			context: *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				Build(),
		}),
	)

	It("applies gateway rules to gateway-origin resources", func() {
		gatewayListener := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(HttpConnectionManager("192.168.0.1:8080", false, nil, true)).
				Configure(HttpDynamicRoute("gateway-route")),
			)).
			MustBuild()
		gatewayCluster := clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
			Configure(clusters.EdsCluster()).
			MustBuild()
		gatewayEndpoints := endpoints.CreateClusterLoadAssignment("backend", []core_xds.Endpoint{
			createEndpointWith("zone-1", "192.168.1.1", map[string]string{}),
			createEndpointWith("zone-2", "192.168.1.2", map[string]string{}),
		})
		gatewayRoute := &envoy_route.RouteConfiguration{
			Name: "gateway-route",
			VirtualHosts: []*envoy_route.VirtualHost{{
				Name:    "*",
				Domains: []string{"*"},
				Routes: []*envoy_route.Route{{
					Match: &envoy_route.RouteMatch{
						PathSpecifier: &envoy_route.RouteMatch_Prefix{Prefix: "/"},
					},
					Action: &envoy_route.Route_Route{
						Route: &envoy_route.RouteAction{
							ClusterSpecifier: &envoy_route.RouteAction_Cluster{Cluster: "backend"},
						},
					},
				}},
			}},
		}

		resources := core_xds.NewResourceSet()
		resources.Add(&core_xds.Resource{
			Name:     "gateway-listener",
			Origin:   gateway_metadata.OriginGateway,
			Resource: gatewayListener,
		})
		resources.Add(&core_xds.Resource{
			Name:     "backend",
			Origin:   gateway_metadata.OriginGateway,
			Resource: gatewayCluster,
		})
		resources.Add(&core_xds.Resource{
			Name:     "backend",
			Origin:   gateway_metadata.OriginGateway,
			Resource: gatewayEndpoints,
		})
		resources.Add(&core_xds.Resource{
			Name:     "gateway-route",
			Origin:   gateway_metadata.OriginGateway,
			Resource: gatewayRoute,
		})

		proxy := &core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
			Zone:       "zone-1",
			Dataplane: builders.Dataplane().
				WithName("sample-gateway").
				WithAddress("192.168.0.1").
				WithDelegatedGateway("sample-gateway").
				Build(),
			Policies: *xds_builders.MatchedPolicies().
				WithGatewayPolicy(api.MeshLoadBalancingStrategyType, core_rules.GatewayRules{
					ToRules: core_rules.GatewayToRules{
						ByListener: map[core_rules.InboundListener]core_rules.ToRules{
							{Address: "192.168.0.1", Port: 8080}: {
								Rules: core_rules.Rules{{
									Subset: subsetutils.MeshService("backend"),
									Conf: api.Conf{
										HashPolicies: &[]api.HashPolicy{{
											Type: api.QueryParameterType,
											QueryParameter: &api.QueryParameter{
												Name: "queryparam",
											},
											Terminal: pointer.To(true),
										}},
										LoadBalancer: &api.LoadBalancer{
											Type: api.RingHashType,
											RingHash: &api.RingHash{
												MinRingSize:  pointer.To[uint32](100),
												MaxRingSize:  pointer.To[uint32](1000),
												HashFunction: pointer.To(api.MurmurHash2Type),
											},
										},
									},
								}},
							},
						},
					},
				}).
				Build(),
		}

		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resources, *xds_builders.Context().WithMeshBuilder(samples.MeshMTLSBuilder()).Build(), proxy)).To(Succeed())

		cluster := gatewayCluster.(*envoy_cluster.Cluster)
		Expect(cluster.LbPolicy).To(Equal(envoy_cluster.Cluster_RING_HASH))
		Expect(cluster.GetRingHashLbConfig().GetMinimumRingSize().GetValue()).To(Equal(uint64(100)))
		Expect(cluster.GetRingHashLbConfig().GetMaximumRingSize().GetValue()).To(Equal(uint64(1000)))
		Expect(cluster.GetRingHashLbConfig().GetHashFunction()).To(Equal(envoy_cluster.Cluster_RingHashLbConfig_MURMUR_HASH_2))

		Expect(gatewayEndpoints.Endpoints).To(HaveLen(2))
		Expect(gatewayEndpoints.Endpoints[1].Priority).To(Equal(uint32(1)))

		routeAction := gatewayRoute.VirtualHosts[0].Routes[0].GetRoute()
		Expect(routeAction.HashPolicy).To(HaveLen(1))
		Expect(routeAction.HashPolicy[0].GetQueryParameter().GetName()).To(Equal("queryparam"))
		Expect(routeAction.HashPolicy[0].GetTerminal()).To(BeTrue())
	})

	It("applies gateway hash policies only to routes for the targeted service", func() {
		gatewayListener := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(HttpConnectionManager("192.168.0.1:8080", false, nil, true)).
				Configure(HttpDynamicRoute("gateway-route")),
			)).
			MustBuild()
		backendCluster := clusters.NewClusterBuilder(envoy_common.APIV3, "backend").
			Configure(clusters.EdsCluster()).
			MustBuild()
		paymentsCluster := clusters.NewClusterBuilder(envoy_common.APIV3, "payments").
			Configure(clusters.EdsCluster()).
			MustBuild()
		backendRoute := &envoy_route.Route{
			Match: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{Prefix: "/backend"},
			},
			Action: &envoy_route.Route_Route{
				Route: &envoy_route.RouteAction{
					ClusterSpecifier: &envoy_route.RouteAction_Cluster{Cluster: "backend"},
				},
			},
		}
		paymentsRoute := &envoy_route.Route{
			Match: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{Prefix: "/payments"},
			},
			Action: &envoy_route.Route_Route{
				Route: &envoy_route.RouteAction{
					ClusterSpecifier: &envoy_route.RouteAction_Cluster{Cluster: "payments"},
				},
			},
		}
		gatewayRoute := &envoy_route.RouteConfiguration{
			Name: "gateway-route",
			VirtualHosts: []*envoy_route.VirtualHost{{
				Name:    "*",
				Domains: []string{"*"},
				Routes:  []*envoy_route.Route{backendRoute, paymentsRoute},
			}},
		}

		resources := core_xds.NewResourceSet()
		resources.Add(&core_xds.Resource{
			Name:     "gateway-listener",
			Origin:   gateway_metadata.OriginGateway,
			Resource: gatewayListener,
		})
		resources.Add(&core_xds.Resource{
			Name:     "backend",
			Origin:   gateway_metadata.OriginGateway,
			Resource: backendCluster,
		})
		resources.Add(&core_xds.Resource{
			Name:     "payments",
			Origin:   gateway_metadata.OriginGateway,
			Resource: paymentsCluster,
		})
		resources.Add(&core_xds.Resource{
			Name:     "gateway-route",
			Origin:   gateway_metadata.OriginGateway,
			Resource: gatewayRoute,
		})

		proxy := &core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
			Dataplane: builders.Dataplane().
				WithName("sample-gateway").
				WithAddress("192.168.0.1").
				WithDelegatedGateway("sample-gateway").
				Build(),
			Policies: *xds_builders.MatchedPolicies().
				WithGatewayPolicy(api.MeshLoadBalancingStrategyType, core_rules.GatewayRules{
					ToRules: core_rules.GatewayToRules{
						ByListener: map[core_rules.InboundListener]core_rules.ToRules{
							{Address: "192.168.0.1", Port: 8080}: {
								Rules: core_rules.Rules{
									{
										Subset: subsetutils.MeshService("backend"),
										Conf: api.Conf{
											HashPolicies: &[]api.HashPolicy{{
												Type: api.HeaderType,
												Header: &api.Header{
													Name: "x-backend",
												},
											}},
										},
									},
									{
										Subset: subsetutils.MeshService("payments"),
										Conf: api.Conf{
											HashPolicies: &[]api.HashPolicy{{
												Type: api.QueryParameterType,
												QueryParameter: &api.QueryParameter{
													Name: "payment",
												},
											}},
										},
									},
								},
							},
						},
					},
				}).
				Build(),
		}

		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resources, *xds_builders.Context().WithMeshBuilder(samples.MeshMTLSBuilder()).Build(), proxy)).To(Succeed())

		backendHashPolicies := backendRoute.GetRoute().HashPolicy
		Expect(backendHashPolicies).To(HaveLen(1))
		Expect(backendHashPolicies[0].GetHeader().GetHeaderName()).To(Equal("x-backend"))

		paymentsHashPolicies := paymentsRoute.GetRoute().HashPolicy
		Expect(paymentsHashPolicies).To(HaveLen(1))
		Expect(paymentsHashPolicies[0].GetQueryParameter().GetName()).To(Equal("payment"))
	})

	type zoneProxyTestCase struct {
		conf            api.Conf
		endpoints       []core_xds.Endpoint
		expectedCluster string
	}

	mesKRI := kri.Identifier{
		ResourceType: meshexternalservice_api.MeshExternalServiceType,
		Mesh:         "default",
		Name:         "example",
		SectionName:  strconv.Itoa(mesPort),
	}

	DescribeTable("Apply to mesh-scoped zone proxy Dataplanes",
		func(given zoneProxyTestCase) {
			endpoints := given.endpoints
			if len(endpoints) == 0 {
				endpoints = []core_xds.Endpoint{externalServiceEndpoint("192.168.0.1", 0)}
			}

			resources := core_xds.NewResourceSet()
			resources.Add(&core_xds.Resource{
				Name:   mesKRI.String(),
				Origin: metadata.OriginEgress,
				// mirrors ZoneProxyListenerGenerator.genClusterCDS
				Resource: clusters.NewClusterBuilder(envoy_common.APIV3, mesKRI.String()).
					Configure(clusters.ProvidedCustomEndpointCluster(
						false,
						true,
						endpoints...,
					)).MustBuild(),
				ResourceOrigin: mesKRI,
			})

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				Build()

			proxy := xds_builders.Proxy().
				With(func(p *core_xds.Proxy) {
					p.Dataplane = zoneProxyDataplane()
				}).
				WithPolicies(xds_builders.MatchedPolicies().WithPolicy(
					api.MeshLoadBalancingStrategyType,
					core_rules.ToRules{
						ResourceRules: outbound.ResourceRules{
							mesKRI: {Conf: []any{given.conf}},
						},
					},
					core_rules.FromRules{},
				)).
				Build()

			p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(p.Apply(resources, xdsCtx, proxy)).To(Succeed())

			resource, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", given.expectedCluster)))
		},
		Entry("round robin", zoneProxyTestCase{
			conf:            api.Conf{LoadBalancer: &api.LoadBalancer{Type: api.RoundRobinType}},
			expectedCluster: "zone-proxy-round-robin.clusters.golden.yaml",
		}),
		Entry("least request", zoneProxyTestCase{
			conf: api.Conf{LoadBalancer: &api.LoadBalancer{
				Type: api.LeastRequestType,
				LeastRequest: &api.LeastRequest{
					ChoiceCount:       pointer.To[uint32](4),
					ActiveRequestBias: pointer.To(intstr.FromString("1.3")),
				},
			}},
			expectedCluster: "zone-proxy-least-request.clusters.golden.yaml",
		}),
		Entry("ring hash", zoneProxyTestCase{
			conf: api.Conf{LoadBalancer: &api.LoadBalancer{
				Type: api.RingHashType,
				RingHash: &api.RingHash{
					MinRingSize:  pointer.To[uint32](100),
					MaxRingSize:  pointer.To[uint32](1000),
					HashFunction: pointer.To(api.MurmurHash2Type),
				},
			}},
			expectedCluster: "zone-proxy-ring-hash.clusters.golden.yaml",
		}),
		Entry("maglev", zoneProxyTestCase{
			conf:            api.Conf{LoadBalancer: &api.LoadBalancer{Type: api.MaglevType}},
			expectedCluster: "zone-proxy-maglev.clusters.golden.yaml",
		}),
		Entry("random", zoneProxyTestCase{
			conf:            api.Conf{LoadBalancer: &api.LoadBalancer{Type: api.RandomType}},
			expectedCluster: "zone-proxy-random.clusters.golden.yaml",
		}),
		Entry("locality awareness is ignored", zoneProxyTestCase{
			conf: api.Conf{
				LoadBalancer: &api.LoadBalancer{Type: api.RandomType},
				LocalityAwareness: &api.LocalityAwareness{
					CrossZone: &api.CrossZone{
						Failover: &[]api.Failover{{
							To: api.ToZone{Type: api.Only, Zones: &[]string{"zone-2"}},
						}},
					},
				},
			},
			expectedCluster: "zone-proxy-locality-awareness.clusters.golden.yaml",
		}),
		Entry("endpoints with different priorities", zoneProxyTestCase{
			conf: api.Conf{
				LoadBalancer: &api.LoadBalancer{Type: api.LeastRequestType},
				LocalityAwareness: &api.LocalityAwareness{
					LocalZone: &api.LocalZone{
						AffinityTags: &[]api.AffinityTag{{Key: "k8s.io/node", Weight: pointer.To[uint32](9000)}},
					},
					CrossZone: &api.CrossZone{
						Failover: &[]api.Failover{{
							To: api.ToZone{Type: api.Only, Zones: &[]string{"zone-2"}},
						}},
						FailoverThreshold: &api.FailoverThreshold{Percentage: intstr.FromInt32(70)},
					},
				},
			},
			endpoints: []core_xds.Endpoint{
				externalServiceEndpoint("192.168.0.1", 0),
				externalServiceEndpoint("192.168.0.2", 1),
				externalServiceEndpoint("192.168.0.3", 2),
			},
			expectedCluster: "zone-proxy-endpoint-priorities.clusters.golden.yaml",
		}),
	)

	It("configures hash policies on the zone proxy egress routes", func() {
		resources := core_xds.NewResourceSet()
		resources.Add(zoneProxyEgressListener(mesKRI))

		xdsCtx := *xds_builders.Context().
			WithMeshBuilder(samples.MeshDefaultBuilder()).
			Build()

		proxy := xds_builders.Proxy().
			With(func(p *core_xds.Proxy) {
				p.Dataplane = zoneProxyDataplane()
			}).
			WithPolicies(xds_builders.MatchedPolicies().WithPolicy(
				api.MeshLoadBalancingStrategyType,
				core_rules.ToRules{
					ResourceRules: outbound.ResourceRules{
						mesKRI: {Conf: []any{api.Conf{
							LoadBalancer: &api.LoadBalancer{Type: api.RingHashType},
							HashPolicies: &[]api.HashPolicy{{
								Type:   api.HeaderType,
								Header: &api.Header{Name: "x-hash"},
							}},
						}}},
					},
				},
				core_rules.FromRules{},
			)).
			Build()

		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(p.Apply(resources, xdsCtx, proxy)).To(Succeed())

		actual, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "zone-proxy-hash-policies.listener.golden.yaml")))
	})

	It("keeps locality awareness for a MeshExternalService on a sidecar", func() {
		resources := core_xds.NewResourceSet()
		resources.Add(&core_xds.Resource{
			Name:   mesKRI.String(),
			Origin: metadata.OriginOutbound,
			Resource: endpoints.CreateClusterLoadAssignment(mesKRI.String(), []core_xds.Endpoint{
				createEndpointWith("zone-1", "10.0.0.1", map[string]string{}),
				createEndpointWith("zone-2", "10.0.0.2", map[string]string{}),
			}),
			ResourceOrigin: mesKRI,
		})

		proxy := xds_builders.Proxy().
			WithZone("zone-1").
			With(func(p *core_xds.Proxy) {
				p.Dataplane = builders.Dataplane().
					AddInboundOfTagsMap(map[string]string{mesh_proto.ServiceTag: "backend"}).
					Build()
			}).
			WithPolicies(xds_builders.MatchedPolicies().WithPolicy(
				api.MeshLoadBalancingStrategyType,
				core_rules.ToRules{
					ResourceRules: outbound.ResourceRules{
						mesKRI: {Conf: []any{api.Conf{
							LocalityAwareness: &api.LocalityAwareness{
								CrossZone: &api.CrossZone{
									Failover: &[]api.Failover{{
										To: api.ToZone{Type: api.Only, Zones: &[]string{"zone-2"}},
									}},
								},
							},
						}}},
					},
				},
				core_rules.FromRules{},
			)).
			Build()

		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(p.Apply(resources, *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(), proxy)).To(Succeed())

		cla := resources.ListOf(envoy_resource.EndpointType)[0].Resource.(*envoy_endpoint.ClusterLoadAssignment)
		Expect(cla.Endpoints).To(HaveLen(2))
		zones := []string{cla.Endpoints[0].Locality.Zone, cla.Endpoints[1].Locality.Zone}
		Expect(zones).To(ConsistOf("zone-1", "zone-2"))
		// only claConfigurer writes the overprovisioning factor, so this proves
		// locality awareness ran rather than the endpoints merely surviving
		Expect(cla.Policy.GetOverprovisioningFactor().GetValue()).To(Equal(uint32(200)))
	})
})

// mesPort is the port of the MeshExternalService the zone proxy tests target.
const mesPort = 9000

// externalServiceEndpoint mirrors topology.createMeshExternalServiceEndpoint:
// a Locality with an empty Zone and the priority in the SubZone. The empty Zone
// is what makes locality awareness misfire on the egress.
func externalServiceEndpoint(address string, priority uint32) core_xds.Endpoint {
	ep := *xds_builders.Endpoint().WithTarget(address).WithPort(mesPort).Build()
	ep.Locality = &core_xds.Locality{Priority: priority, SubZone: "priority-" + strconv.Itoa(int(priority))}
	return ep
}

// zoneProxyEgressListener mirrors ZoneProxyListenerGenerator.generateEgressListener:
// one filter chain per MeshExternalService named after its KRI, and no
// ResourceOrigin on the listener itself, since it is shared by all destinations.
func zoneProxyEgressListener(id kri.Identifier) *core_xds.Resource {
	name := naming.ContextualZoneEgressListenerName("ze-port")
	listener, err := NewListenerBuilder(envoy_common.APIV3, name).
		Configure(InboundListener("10.0.0.1", 10002, core_xds.SocketAddressProtocolTCP, false)).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, id.String()).
			Configure(HttpConnectionManager(id.String(), false, nil, false)).
			Configure(AddFilterChainConfigurer(&meshhttproute_xds.HttpOutboundRouteConfigurer{
				RouteConfigName: id.String(),
				VirtualHostName: id.String(),
				Routes: []meshhttproute_xds.OutboundRoute{{
					Match: meshhttproute_api.Match{
						Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/"},
					},
					Split: []envoy_common.Split{
						xds.NewSplitBuilder().WithClusterName(id.String()).WithExternalService(true).WithWeight(1).Build(),
					},
				}},
			})),
		)).Build()
	Expect(err).ToNot(HaveOccurred())
	return &core_xds.Resource{
		Name:     name,
		Origin:   metadata.OriginEgress,
		Resource: listener,
	}
}

func zoneProxyDataplane() *core_mesh.DataplaneResource {
	return &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{Name: "zone-proxy", Mesh: "default"},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "10.0.0.1",
				Listeners: []*mesh_proto.Dataplane_Networking_Listener{{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
					Address: "10.0.0.1",
					Port:    10002,
					Name:    "ze-port",
					State:   mesh_proto.Dataplane_Networking_Listener_Ready,
				}},
			},
		},
	}
}

func createEndpointWith(zone string, ip string, extraTags map[string]string) core_xds.Endpoint {
	return *xds_builders.Endpoint().
		WithTarget(ip).
		WithPort(8080).
		WithTags(mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), mesh_proto.ZoneTag, zone).
		AddTagsMap(extraTags).
		WithZone(zone).
		Build()
}

// createEndpointWithLabels models an endpoint whose workload labels are
// folded into the endpoint tags at topology build time (see
// BuildDataplaneEndpointMap), so they live under the same envoy.lb key as the
// system tags.
func createEndpointWithLabels(ip string, labels map[string]string) core_xds.Endpoint {
	return *xds_builders.Endpoint().
		WithTarget(ip).
		WithPort(8080).
		WithTags(mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), mesh_proto.ZoneTag, "zone-1").
		AddTagsMap(labels).
		WithZone("zone-1").
		Build()
}

// TODO move to routing builder
func paymentsAndBackendRouting() *xds_builders.RoutingBuilder {
	return xds_builders.Routing().
		WithOutboundTargets(
			xds_builders.EndpointMap().
				AddEndpoint("backend", xds_samples.HttpEndpointBuilder()).
				AddEndpoint("payment", xds_samples.HttpEndpointBuilder()),
		)
}

// outboundRoute mirrors the single catch-all route that
// meshhttproute_plugin.GenerateOutboundListener produces for an outbound.
func outboundRoute(service string, splits ...envoy_common.Split) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	match := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.PathPrefix,
			Value: "/",
		},
	}
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		RouteConfigName: envoy_names.GetOutboundRouteName(service),
		VirtualHostName: service,
		Routes: []meshhttproute_xds.OutboundRoute{{
			Name:  string(meshhttproute_api.HashMatches([]meshhttproute_api.Match{match})),
			Match: match,
			Split: splits,
		}},
		DpTags: mesh_proto.MultiValueTagSet{"kuma.io/service": {service: true}},
	}
}

func paymentsListener() envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27778, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(HttpConnectionManager("127.0.0.1:27778", false, nil, true)).
			Configure(AddFilterChainConfigurer(outboundRoute(
				"payment",
				xds.NewSplitBuilder().WithClusterName("payment").WithWeight(100).Build(),
			))),
		)).MustBuild()
}

func backendListener() envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(HttpConnectionManager("127.0.0.1:27777", false, nil, true)).
			Configure(AddFilterChainConfigurer(outboundRoute(
				"backend",
				xds.NewSplitBuilder().WithClusterName("backend").WithWeight(100).Build(),
			))),
		)).MustBuild()
}

func outboundRealServiceHTTPListener(serviceResourceKRI kri.Identifier, port int32, routes []meshhttproute_xds.OutboundRoute) core_xds.Resource {
	listener, err := meshhttproute_plugin.GenerateOutboundListener(
		&core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
		},
		meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Address:  "127.0.0.1",
				Port:     uint32(port),
				Resource: serviceResourceKRI,
			},
			Protocol: core_meta.ProtocolHTTP,
		},
		routes,
		mesh_proto.MultiValueTagSet{"kuma.io/service": {"backend": true}},
	)
	Expect(err).ToNot(HaveOccurred())
	return *listener
}
