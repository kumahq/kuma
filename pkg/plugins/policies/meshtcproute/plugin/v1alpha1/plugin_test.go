package v1alpha1_test

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/metrics"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_protocol "github.com/kumahq/kuma/pkg/util/protocol"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_server "github.com/kumahq/kuma/pkg/xds/server"
)

func getResource(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshTCPRoute", func() {
	backendMeshServiceIdentifier := core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name: "backend",
			Mesh: "default",
		},
		ResourceType: "MeshService",
		SectionName:  "",
	}

	backendMeshExternalServiceIdentifier := core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name: "example",
			Mesh: "default",
		},
		ResourceType: "MeshExternalService",
	}

	type policiesTestCase struct {
		dataplane      *core_mesh.DataplaneResource
		resources      xds_context.Resources
		expectedRoutes core_rules.ToRules
	}

	DescribeTable("MatchedPolicies",
		func(given policiesTestCase) {
			routes, err := plugin.NewPlugin().(core_plugins.PolicyPlugin).
				MatchedPolicies(given.dataplane, given.resources)
			Expect(err).ToNot(HaveOccurred())
			Expect(routes.ToRules).To(Equal(given.expectedRoutes))
		},

		Entry("basic", policiesTestCase{
			dataplane: samples.DataplaneWeb(),
			resources: xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					api.MeshTCPRouteType: &api.MeshTCPRouteResourceList{
						Items: []*api.MeshTCPRouteResource{
							{
								Meta: &test_model.ResourceMeta{
									Mesh: core_model.DefaultMesh,
									Name: "route-1",
								},
								Spec: &api.MeshTCPRoute{
									TargetRef: pointer.To(builders.TargetRefMesh()),
									To: []api.To{
										{
											TargetRef: builders.TargetRefService("backend"),
										},
									},
								},
							},
							{
								Meta: &test_model.ResourceMeta{
									Mesh: core_model.DefaultMesh,
									Name: "route-2",
								},
								Spec: &api.MeshTCPRoute{
									TargetRef: pointer.To(builders.TargetRefService("web")),
									To: []api.To{
										{
											TargetRef: builders.TargetRefService("backend"),
											Rules: []api.Rule{
												{
													Default: api.RuleConf{
														BackendRefs: []common_api.BackendRef{
															{
																TargetRef: builders.TargetRefServiceSubset(
																	"backend",
																	"version", "v1",
																),
																Weight: pointer.To(uint(50)),
															},
															{
																TargetRef: builders.TargetRefServiceSubset(
																	"backend",
																	"version", "v2",
																),
																Weight: pointer.To(uint(50)),
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
			expectedRoutes: core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"version", "v1",
										),
										Weight: pointer.To(uint(50)),
									},
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"version", "v2",
										),
										Weight: pointer.To(uint(50)),
									},
								},
							},
						},
						Origin: []core_model.ResourceMeta{
							&test_model.ResourceMeta{
								Mesh: "default",
								Name: "route-1",
							},
							&test_model.ResourceMeta{
								Mesh: "default",
								Name: "route-2",
							},
						},
						BackendRefOriginIndex: map[common_api.MatchesHash]int{
							core_rules.EmptyMatches: 1,
						},
					},
				},
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{},
			},
		}),
	)

	type outboundsTestCase struct {
		proxy      *core_xds.Proxy
		xdsContext xds_context.Context
	}

	DescribeTable("Apply",
		func(given outboundsTestCase) {
			metrics, err := metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())

			claCache, err := cla.NewCache(1*time.Second, metrics)
			Expect(err).ToNot(HaveOccurred())
			given.xdsContext.ControlPlane.CLACache = claCache

			secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil, false)
			dataSourceLoader := datasource.NewDataSourceLoader(secretManager)
			given.xdsContext.Mesh.DataSourceLoader = dataSourceLoader

			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), given.xdsContext.Mesh, given.proxy)).To(Succeed(), n)
			}
			resourceSet := core_xds.NewResourceSet()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resourceSet, given.xdsContext, given.proxy)).
				To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			listenersGolden := filepath.Join("testdata",
				name+".listeners.golden.yaml",
			)
			clustersGolden := filepath.Join("testdata",
				name+".clusters.golden.yaml",
			)
			endpointsGolden := filepath.Join("testdata",
				name+".endpoints.golden.yaml",
			)

			Expect(getResource(resourceSet, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(listenersGolden))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(clustersGolden))
			Expect(getResource(resourceSet, envoy_resource.EndpointType)).
				To(matchers.MatchGoldenYAML(endpointsGolden))
		},

		Entry("split-traffic", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8005).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("other-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			externalServiceOutboundTargets := xds_builders.EndpointMap().
				AddEndpoint("externalservice", xds_builders.Endpoint().
					WithTarget("192.168.0.7").
					WithPort(8007).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "externalservice", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP2))
			rules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "eu",
										),
										Weight: pointer.To(uint(40)),
									},
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "us",
										),
										Weight: pointer.To(uint(15)),
									},
									{
										TargetRef: builders.TargetRefService(
											"other-backend",
										),
										Weight: pointer.To(uint(15)),
									},
									{
										TargetRef: builders.TargetRefService(
											"externalservice",
										),
										Weight: pointer.To(uint(15)),
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithExternalServicesEndpointMap(externalServiceOutboundTargets).
					AddServiceProtocol("backend", util_protocol.GetCommonProtocol(core_mesh.ProtocolTCP, core_mesh.ProtocolHTTP)).
					AddServiceProtocol("other-backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("externalservice", core_mesh.ProtocolHTTP2).
					AddExternalService("externalservice").
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 1,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "other-backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 2,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "externalservice",
							},
						}},
					}).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets).
							WithExternalServiceOutboundTargets(externalServiceOutboundTargets),
					).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),
		Entry("default-meshexternalservice", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: core_mesh.ProtocolTCP,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
						{
							Address: "192.168.1.1",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
					},
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
					},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{
						IP: "10.20.20.1",
					},
				},
			}

			dp, proxy := dppForMeshExternalService(&meshExtSvc)
			egress := builders.ZoneEgress().WithPort(10002).Build()
			mc := meshContextForMeshExternalService(dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("tcproute-meshexternalservice", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: core_mesh.ProtocolTCP,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
						{
							Address: "192.168.1.1",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
					},
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
					},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{
						IP: "10.20.20.1",
					},
				},
			}
			meshExtSvc2 := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example2", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: core_mesh.ProtocolTCP,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
						{
							Address: "192.168.1.1",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
					},
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
					},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{
						IP: "10.20.20.2",
					},
				},
			}

			dp, proxy := dppForMeshExternalService(&meshExtSvc, &meshExtSvc2)
			egress := builders.ZoneEgress().WithPort(10002).Build()
			mc := meshContextForMeshExternalService(egress, &meshExtSvc, &meshExtSvc2, dp.Build())

			proxy.Policies = core_xds.MatchedPolicies{
				Dynamic: core_xds.PluginOriginatedPolicies{},
			}
			proxy.Policies.Dynamic[api.MeshTCPRouteType] = core_xds.TypedMatchingPolicies{
				Type: api.MeshTCPRouteType,
				ToRules: core_rules.ToRules{
					ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
						backendMeshExternalServiceIdentifier: {
							Origin: []core_rules.Origin{
								{Resource: &test_model.ResourceMeta{Mesh: "default", Name: "tcp-route"}},
							},
							BackendRefOriginIndex: map[common_api.MatchesHash]int{
								core_rules.EmptyMatches: 0,
							},
							Conf: []interface{}{
								api.Rule{
									Default: api.RuleConf{
										BackendRefs: []common_api.BackendRef{
											{
												TargetRef: builders.TargetRefMeshExternalService("example2"),
												Weight:    pointer.To(uint(100)),
												Port:      pointer.To(uint32(9090)),
											},
										},
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("basic-no-policies", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8005).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("other-service", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			externalServiceOutboundTargets := xds_builders.EndpointMap().
				AddEndpoint("externalservice", xds_builders.Endpoint().
					WithTarget("192.168.0.7").
					WithPort(8007).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "externalservice", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP2))

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", util_protocol.GetCommonProtocol(core_mesh.ProtocolTCP, core_mesh.ProtocolHTTP)).
					AddServiceProtocol("other-backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("externalservice", core_mesh.ProtocolHTTP2).
					AddExternalService("externalservice").
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						}},
					}).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets).
							WithExternalServiceOutboundTargets(externalServiceOutboundTargets),
					).
					Build(),
			}
		}()),

		Entry("basic-real-meshservice", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  intstr.FromInt(8084),
						AppProtocol: core_mesh.ProtocolHTTP,
						Name:        "test-port",
					}},
					Identities: []meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentityServiceTagType,
							Value: "backend",
						},
					},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.1",
					}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "app", "backend"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("default_backend___msvc_80", core_mesh.ProtocolHTTP).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: pointer.To(core_model.NewTypedResourceIdentifier(&meshSvc, core_model.WithSectionName("test-port"))),
						},
					}).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, core_rules.ToRules{
								ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
									backendMeshServiceIdentifier: {
										Origin: []core_rules.Origin{
											{Resource: &test_model.ResourceMeta{Mesh: "default", Name: "tcp-route"}},
										},
										BackendRefOriginIndex: map[common_api.MatchesHash]int{
											core_rules.EmptyMatches: 0,
										},
										Conf: []interface{}{
											api.Rule{
												Default: api.RuleConf{
													BackendRefs: []common_api.BackendRef{
														{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
															Port:      pointer.To(uint32(80)),
														},
													},
												},
											},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),

		Entry("redirect-traffic", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP)).
				AddEndpoint("tcp-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP))
			rules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefService(
											"tcp-backend",
										),
										Weight: pointer.To(uint(1)),
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("tcp-backend", core_mesh.ProtocolTCP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 1,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "tcp-backend",
							},
						}},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),

		Entry("meshhttproute-clash-http-destination", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP)).
				AddEndpoint("tcp-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)).
				AddEndpoint("http-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "http-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			tcpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefService(
											"tcp-backend",
										),
										Weight: pointer.To(uint(1)),
									},
								},
							},
						},
					},
				},
			}

			httpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: meshhttproute_api.PolicyDefault{
							Rules: []meshhttproute_api.Rule{
								{
									Matches: []meshhttproute_api.Match{
										{
											Path: &meshhttproute_api.PathMatch{
												Type:  meshhttproute_api.PathPrefix,
												Value: "/",
											},
										},
									},
									Default: meshhttproute_api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{
											{
												TargetRef: builders.TargetRefService("http-backend"),
												Weight:    pointer.To(uint(1)),
											},
										},
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("tcp-backend", core_mesh.ProtocolTCP).
					AddServiceProtocol("http-backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 1,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "tcp-backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 2,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "http-backend",
							},
						}},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, tcpRules).
							WithToPolicy(meshhttproute_api.MeshHTTPRouteType, httpRules),
					).
					Build(),
			}
		}()),

		Entry("meshhttproute-clash-tcp-destination", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)).
				AddEndpoint("tcp-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)).
				AddEndpoint("http-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "http-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			tcpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefService(
											"tcp-backend",
										),
										Weight: pointer.To(uint(1)),
									},
								},
							},
						},
					},
				},
			}

			httpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: meshhttproute_api.PolicyDefault{
							Rules: []meshhttproute_api.Rule{
								{
									Matches: []meshhttproute_api.Match{
										{
											Path: &meshhttproute_api.PathMatch{
												Type:  meshhttproute_api.PathPrefix,
												Value: "/",
											},
										},
									},
									Default: meshhttproute_api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{
											{
												TargetRef: builders.TargetRefService("http-backend"),
												Weight:    pointer.To(uint(1)),
											},
										},
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolTCP).
					AddServiceProtocol("tcp-backend", core_mesh.ProtocolTCP).
					AddServiceProtocol("http-backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 1,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "tcp-backend",
							},
						}},
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 2,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "http-backend",
							},
						}},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, tcpRules).
							WithToPolicy(meshhttproute_api.MeshHTTPRouteType, httpRules),
					).
					Build(),
			}
		}()),
		Entry("kafka", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolKafka, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8005).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolKafka, "region", "us"))
			rules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "eu",
										),
										Weight: pointer.To(uint(60)),
									},
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "us",
										),
										Weight: pointer.To(uint(40)),
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolKafka).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						}},
					}).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets),
					).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),
		Entry("gateway", func() outboundsTestCase {
			gateway := &core_mesh.MeshGatewayResource{
				Meta: &test_model.ResourceMeta{Name: "sample-gateway", Mesh: "default"},
				Spec: &mesh_proto.MeshGateway{
					Selectors: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								mesh_proto.ServiceTag: "sample-gateway",
							},
						},
					},
					Conf: &mesh_proto.MeshGateway_Conf{
						Listeners: []*mesh_proto.MeshGateway_Listener{
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8080,
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8081,
								Hostname: "go.dev",
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_TCP,
								Port:     9080,
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_TLS,
								Port:     9081,
								Hostname: "go.dev",
								Tls: &mesh_proto.MeshGateway_TLS_Conf{
									Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
									Certificates: []*system_proto.DataSource{{
										Type: &system_proto.DataSource_Inline{
											Inline: wrapperspb.Bytes([]byte(secret)),
										},
									}},
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_TLS,
								Port:     9081,
								Hostname: "other.dev",
								Tls: &mesh_proto.MeshGateway_TLS_Conf{
									Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
									Certificates: []*system_proto.DataSource{{
										Type: &system_proto.DataSource_Inline{
											Inline: wrapperspb.Bytes([]byte(secret)),
										},
									}},
								},
							},
						},
					},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{gateway},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "us"),
				)
			xdsContext := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				WithEndpointMap(outboundTargets).Build()

			rules := core_rules.Rule{
				Subset: core_rules.MeshSubset(),
				Conf: api.Rule{
					Default: api.RuleConf{
						BackendRefs: []common_api.BackendRef{{
							TargetRef: builders.TargetRefService("backend"),
							Weight:    pointer.To(uint(100)),
						}},
					},
				},
			}
			tlsGoRules := core_rules.Rule{
				Subset: core_rules.MeshSubset(),
				Conf: api.Rule{
					Default: api.RuleConf{
						BackendRefs: []common_api.BackendRef{{
							TargetRef: builders.TargetRefService("go-backend-1"),
							Weight:    pointer.To(uint(50)),
						}, {
							TargetRef: builders.TargetRefService("go-backend-2"),
							Weight:    pointer.To(uint(50)),
						}},
					},
				},
			}
			tlsOtherRules := core_rules.Rule{
				Subset: core_rules.MeshSubset(),
				Conf: api.Rule{
					Default: api.RuleConf{
						BackendRefs: []common_api.BackendRef{{
							TargetRef: builders.TargetRefService("other-backend"),
							Weight:    pointer.To(uint(100)),
						}},
					},
				},
			}
			return outboundsTestCase{
				xdsContext: *xdsContext,
				proxy: xds_builders.Proxy().
					WithDataplane(samples.GatewayDataplaneBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithGatewayPolicy(api.MeshTCPRouteType, core_rules.GatewayRules{
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 9080, "*"):         {&rules, &tlsOtherRules},
										core_rules.NewInboundListenerHostname("192.168.0.1", 9081, "other.dev"): {&tlsOtherRules},
										core_rules.NewInboundListenerHostname("192.168.0.1", 9081, "go.dev"):    {&tlsGoRules},
									},
								},
							}),
					).
					Build(),
			}
		}()),
	)
})

// nolint:gosec
const secret = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAqEhj+XS8qgm3raPrP554uXDiPv0np2lCx1wJF4KiwFGJMAV8
qHul/0pcUCX742irsAV39f6sMytXlRpMfBAbJNyZDuqx36s0yMolMxsqMjUHmI+X
W7zrj1xAkPLjB+kireohXkyXESBy3QbcaAW+ftZdFDcNHC7a9W8eSZzCB5R0Sb1S
YMrMq8gXrDbB99fLb2G5wsGXq9xW1g6u7TqWy5TAvYkErMfXfsx3BcvJPPAaI4vb
hPp034KVlucB3h5QSEDF1AIV7A8r1m3I/yHRZqyvhg6Dp4ZTgZw1Sh7QwYsJr6/h
kIVaBjq+gT+I6oPBOnVrc5W3N/fO8yalaAdvswIDAQABAoIBAQCS8ywCMRNy9Ktl
wQdz9aF8Zfvbf1t6UGvVBSSXWCdhA5Jl0dTKl7ccGEZGYvTz33pVamEX+j1LLaT8
eguiJrpdVRl/MikDpVChqgwT9bvCPhaU/YbxwCZ/eNKVANSKGuaCsjpTS1R7yzci
lZQwbhusTOrY9T3Ih44C1va+11mEHY7rAy96r2MgTdpDdWAqhGKxQ88IyNCTvp6u
1I/oWXYDm7QW7HCEWcw2PyFfcfLy4LCPYG7BMX6n1DMSSu6U2PeV1fm6wleawCCN
KxuKQSBHARM9B0pcPpAhGuXO9fHBllz3Tmw0yJYCUopIxPK/r+yMufpsto6KRJOz
had7o4XJAoGBAMSdr1eRG2TBwfQtGS9WxMUrYiCdNCDMFnXsFrKp5kF7ebRyX0lY
41O/KS3SPRmqn6F8t77+VjAvIcCtVWPgTLGo4QyOV09UAcPOrv4qBHRkT8tNyM1n
q15DGd7ICE0LFuK1zjWu1HBz/64hNqJJxC8tcJ1HgQ7sO9Vl0FMHeXcNAoGBANsb
/QqyRixj0UMhST4MoZzxwV+3Y+//mpEL4R1kcFa0K1BrIq80xCzJzK7jrU7XtaeG
0WZpksYqexzN6kXvuJy3w5rC4LC2/+MHspYKvdkUMjctB1XIAPF2FtdrSfMDjweS
ItJ1QqALcc83XzAMkrrCUUeL45SGWxRp3yLljtG/AoGAcPAWwRkEADtf+q9RESUp
QAysgAls4Q36NOBZJWV8cs7HWQR9gXdClV9v+vcRy8V7jlpCfb5AqcrY+4FVVFqK
E17rbrfwpQufO+dkE3D1QBpCz4gtuPc8s5edq5+BTSf6jF1cRu/W7YVkL5S6ejwf
Ke5TCrUBCB5gPDMQmDDp750CgYAHMdwVRdVYD88HTUiCaRfFd4rKAdOeRd5ldOZn
eKzXrALgGSSCbFEkx1uZQpCmTh8A6URnAIB5UVvJjllrAnwlaUNbCZsnMlsksVQD
6UZiom8jsK7U+kRNqXsGh9ddy3ge34WVM5SEfNu32jGd+ku3JjpVBxrp/Z9wBCn3
k2IlMQKBgQCWsVuAoLcEvtKYSBb4KZZY3+pHkLLxe+K7Cpq5fK7RnueaH9+o1g+8
AdY6vX/j9yVHqfF6DI2tyq0qMcuNkjDirlY3yosZEQOXjW8SIGk3YaHwd4JMqVL6
vBGM7k3/smF7hEG97wUeaMe3IDkP7G4SNZOWbLUy1IjLw8BBK+2FVQ==
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIRAK2DKOd4qR4eTfFpTHCY0KAwDQYJKoZIhvcNAQELBQAw
GzEZMBcGA1UEAxMQZWNoby5leGFtcGxlLmNvbTAeFw0yMTExMDEwNDMzNDhaFw0z
MTEwMzAwNDMzNDhaMBsxGTAXBgNVBAMTEGVjaG8uZXhhbXBsZS5jb20wggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCoSGP5dLyqCbeto+s/nni5cOI+/Sen
aULHXAkXgqLAUYkwBXyoe6X/SlxQJfvjaKuwBXf1/qwzK1eVGkx8EBsk3JkO6rHf
qzTIyiUzGyoyNQeYj5dbvOuPXECQ8uMH6SKt6iFeTJcRIHLdBtxoBb5+1l0UNw0c
Ltr1bx5JnMIHlHRJvVJgysyryBesNsH318tvYbnCwZer3FbWDq7tOpbLlMC9iQSs
x9d+zHcFy8k88Boji9uE+nTfgpWW5wHeHlBIQMXUAhXsDyvWbcj/IdFmrK+GDoOn
hlOBnDVKHtDBiwmvr+GQhVoGOr6BP4jqg8E6dWtzlbc3987zJqVoB2+zAgMBAAGj
dDByMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMB
Af8EBTADAQH/MB0GA1UdDgQWBBS+iZdWqEBq5IT4b9Dcdx09MTUuCzAbBgNVHREE
FDASghBlY2hvLmV4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQBRUD8uWq0s
IM3sW+MCAtBQq5ppNstlAeH24w3yO+4v64FqjDUwRLq7uMJza9iNdbYDQZW/NRrv
30Om9PSn02WzlANa2Knm/EoCwgPyA4ED1UD77uWnxOUxfEWeqdOYDElJpIRb+7RO
tW9zD7ZJ89ipvEjL2zGuvKCQKkdYaIm7W2aljDz1olsMgQolHpbTEPjN+RMWiyNs
tDaan+pwBI0OoXzuWPpB8o9jfL7I8YeOQXOmNy/qpvELV8ji3vdPH1xu1NSt1EGV
rZigv0SZ20Y+BHgf0y3Tv0X+Rx96lYiUtfU+54vjokEjSsfF+iauxfL75QuVvAf9
7G3tiTJPwFKA
-----END CERTIFICATE-----
`

func meshContextForMeshExternalService(resources ...core_model.Resource) *xds_context.MeshContext {
	resourceStore := memory.NewStore()
	mesh := builders.Mesh().WithBuiltinMTLSBackend("ca-1").WithEgressRoutingEnabled().WithEnabledMTLSBackend("ca-1").Build()
	err := resourceStore.Create(context.Background(), mesh, store.CreateByKey("default", model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	for _, res := range resources {
		err = resourceStore.Create(context.Background(), res, store.CreateByKey(res.GetMeta().GetName(), res.GetMeta().GetMesh()))
		Expect(err).ToNot(HaveOccurred())
	}

	lookupIPFunc := func(s string) ([]net.IP, error) {
		return []net.IP{net.ParseIP(s)}, nil
	}
	meshContextBuilder := xds_context.NewMeshContextBuilder(
		resourceStore,
		xds_server.MeshResourceTypes(),
		lookupIPFunc,
		"zone-1",
		vips.NewPersistence(core_manager.NewResourceManager(resourceStore), manager.NewConfigManager(resourceStore), false),
		"mesh",
		80,
		xds_context.AnyToAnyReachableServicesGraphBuilder,
	)
	mc, err := meshContextBuilder.Build(context.Background(), "default")
	Expect(err).ToNot(HaveOccurred())

	return &mc
}

func dppForMeshExternalService(mesList ...*meshexternalservice_api.MeshExternalServiceResource) (*builders.DataplaneBuilder, *core_xds.Proxy) {
	outbounds := xds_types.Outbounds{
		{
			LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Port: builders.FirstOutboundPort,
				Tags: map[string]string{
					mesh_proto.ServiceTag: "backend",
				},
			},
		},
	}
	for _, mes := range mesList {
		outbounds = append(outbounds, &xds_types.Outbound{
			Address:  mes.Status.VIP.IP,
			Port:     uint32(mes.Spec.Match.Port),
			Resource: pointer.To(core_model.NewTypedResourceIdentifier(mes)),
		})
	}
	dp := builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")
	proxy := xds_builders.Proxy().
		WithDataplane(dp).
		WithOutbounds(outbounds).
		WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
		WithMetadata(&core_xds.DataplaneMetadata{
			SystemCaPath: "/tmp/ca-certs.crt",
		}).
		Build()

	return dp, proxy
}
