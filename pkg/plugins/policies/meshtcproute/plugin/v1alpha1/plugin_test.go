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
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/datasource"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/metrics"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_policies "github.com/kumahq/kuma/v3/pkg/test/policies"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_server "github.com/kumahq/kuma/v3/pkg/xds/server"
)

var _ = Describe("MeshTCPRoute", func() {
	backendMeshServiceIdentifier := kri.Identifier{
		ResourceType: "MeshService",
		Mesh:         "default",
		Name:         "backend",
		SectionName:  "",
	}

	backendMeshExternalServiceIdentifier := kri.Identifier{
		ResourceType: "MeshExternalService",
		Mesh:         "default",
		Name:         "example",
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
									TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
									To: &[]api.To{
										{
											TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
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
									TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefDataplaneName("web-01"))),
									To: &[]api.To{
										{
											TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
											Rules: []api.Rule{
												{
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{
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
						Subset: subsetutils.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{
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
						OriginByMatches: map[common_api.MatchesHash]core_model.ResourceMeta{},
					},
				},
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{},
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

			secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None())
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

			resource, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(listenersGolden))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(clustersGolden))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.EndpointType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(endpointsGolden))
		},

		Entry("split-traffic", func() outboundsTestCase {
			meshSvcBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8004)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.1.1"}},
				},
			}
			meshSvcOtherBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "other-backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8006)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("http-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.1.2"}},
				},
			}
			meshSvcBackendEU := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend-eu", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8004)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.1.3"}},
				},
			}
			meshSvcBackendUS := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend-us", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8005)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("http-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.1.4"}},
				},
			}
			meshExtSvcExternal := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "externalservice", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     8007,
						Protocol: core_meta.ProtocolHTTP2,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "192.168.0.7",
							Port:    8007,
						},
					},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{
						IP: "10.20.20.7",
					},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvcBackend, &meshSvcOtherBackend, &meshSvcBackendEU, &meshSvcBackendUS},
			}
			resources.MeshLocalResources[meshexternalservice_api.MeshExternalServiceType] = &meshexternalservice_api.MeshExternalServiceResourceList{
				Items: []*meshexternalservice_api.MeshExternalServiceResource{&meshExtSvcExternal},
			}

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend-eu___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend-eu", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP))).
				AddEndpoint("default_backend-us___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend-us", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP))).
				AddEndpoint("default_other-backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP))).
				AddEndpoint("default_externalservice___extsvc_8007", xds_builders.Endpoint().
					WithTarget("192.168.0.7").
					WithPort(8007).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{OwnerResource: kri.From(&meshExtSvcExternal)}).
					WithTags(mesh_proto.ServiceTag, "externalservice", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP2)))
			rules := core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					kri.From(&meshSvcBackend): test_policies.NewOutboundRule(nil, api.Rule{
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{
								{
									TargetRef: builders.TargetRefMeshService(
										"backend-eu", "", "tcp-port",
									),
									Weight: pointer.To(uint(40)),
								},
								{
									TargetRef: builders.TargetRefMeshService(
										"backend-us", "", "http-port",
									),
									Weight: pointer.To(uint(15)),
								},
								{
									TargetRef: builders.TargetRefMeshService(
										"other-backend", "", "http-port",
									),
									Weight: pointer.To(uint(15)),
								},
								{
									TargetRef: builders.TargetRefMeshExternalService(
										"externalservice",
									),
									Weight: pointer.To(uint(15)),
								},
							},
						},
					}),
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvcBackend), "tcp-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 1,
							Resource: kri.WithSectionName(kri.From(&meshSvcOtherBackend), "http-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 2,
							Resource: kri.From(&meshExtSvcExternal),
						},
					}).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets),
					).
					WithMetadata(&core_xds.DataplaneMetadata{}).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),
		Entry("default-meshexternalservice", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolTCP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    10000,
						},
						{
							Address: "192.168.1.1",
							Port:    10000,
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

			dp, proxy, backendMeshSvc := dppForMeshExternalService(&meshExtSvc)
			mc := meshContextForMeshExternalService(dp.Build(), backendDataplane(), &meshExtSvc, zoneEgressDataplane(), backendMeshSvc)

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
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolTCP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    10000,
						},
						{
							Address: "192.168.1.1",
							Port:    10000,
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
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolTCP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    10000,
						},
						{
							Address: "192.168.1.1",
							Port:    10000,
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

			dp, proxy, backendMeshSvc := dppForMeshExternalService(&meshExtSvc, &meshExtSvc2)
			mc := meshContextForMeshExternalService(zoneEgressDataplane(), &meshExtSvc, &meshExtSvc2, dp.Build(), backendDataplane(), backendMeshSvc)

			proxy.Policies = core_xds.MatchedPolicies{
				Dynamic: core_xds.PluginOriginatedPolicies{},
			}
			proxy.Policies.Dynamic[api.MeshTCPRouteType] = core_xds.TypedMatchingPolicies{
				Type: api.MeshTCPRouteType,
				ToRules: core_rules.ToRules{
					ResourceRules: map[kri.Identifier]outbound.ResourceRule{
						backendMeshExternalServiceIdentifier: test_policies.NewOutboundRule(nil, api.Rule{
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{
									{
										TargetRef: builders.TargetRefMeshExternalService("example2"),
										Weight:    pointer.To(uint(100)),
										Port:      pointer.To(uint32(9090)),
									},
								},
							},
						}),
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("basic-no-policies", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8004)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.2.1"}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("default_backend___msvc_80",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "tcp-port"),
						},
					}).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets),
					).
					WithMetadata(&core_xds.DataplaneMetadata{}).
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
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						},
					}).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.Rule{
										Default: api.RuleConf{
											BackendRefs: &[]common_api.BackendRef{
												{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
													Port:      pointer.To(uint32(80)),
												},
											},
										},
									}),
								},
							}),
					).
					Build(),
			}
		}()),

		Entry("redirect-traffic", func() outboundsTestCase {
			meshSvcBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8004)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("http-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.3.1"}},
				},
			}
			meshSvcTCPBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "tcp-backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8005)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.3.2"}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvcBackend, &meshSvcTCPBackend},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_tcp-backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP)))
			rules := core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					kri.From(&meshSvcBackend): test_policies.NewOutboundRule(nil, api.Rule{
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{
								{
									TargetRef: builders.TargetRefMeshService(
										"tcp-backend", "", "tcp-port",
									),
									Weight: pointer.To(uint(1)),
								},
							},
						},
					}),
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvcBackend), "http-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 1,
							Resource: kri.WithSectionName(kri.From(&meshSvcTCPBackend), "tcp-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(&core_xds.DataplaneMetadata{}).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),

		Entry("meshhttproute-clash-http-destination", func() outboundsTestCase {
			meshSvcBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8004)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("http-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.4.1"}},
				},
			}
			meshSvcTCPBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "tcp-backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8005)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.4.2"}},
				},
			}
			meshSvcHTTPBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "http-backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8006)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("http-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.4.3"}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvcBackend, &meshSvcTCPBackend, &meshSvcHTTPBackend},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_tcp-backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP)))
			tcpRules := core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					kri.From(&meshSvcBackend): test_policies.NewOutboundRule(nil, api.Rule{
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{
								{
									TargetRef: builders.TargetRefMeshService(
										"tcp-backend", "", "tcp-port",
									),
									Weight: pointer.To(uint(1)),
								},
							},
						},
					}),
				},
			}

			httpRules := core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					kri.From(&meshSvcBackend): test_policies.NewOutboundRule(nil, meshhttproute_api.PolicyDefault{
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
											TargetRef: builders.TargetRefMeshService("http-backend", "", "http-port"),
											Weight:    pointer.To(uint(1)),
										},
									},
								},
							},
						},
					}),
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvcBackend), "http-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 1,
							Resource: kri.WithSectionName(kri.From(&meshSvcTCPBackend), "tcp-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 2,
							Resource: kri.WithSectionName(kri.From(&meshSvcHTTPBackend), "http-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(&core_xds.DataplaneMetadata{}).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, tcpRules).
							WithToPolicy(meshhttproute_api.MeshHTTPRouteType, httpRules),
					).
					Build(),
			}
		}()),

		Entry("meshhttproute-clash-tcp-destination", func() outboundsTestCase {
			meshSvcBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8004)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.5.1"}},
				},
			}
			meshSvcTCPBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "tcp-backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8005)),
						AppProtocol: core_meta.ProtocolTCP,
						Name:        pointer.To("tcp-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.5.2"}},
				},
			}
			meshSvcHTTPBackend := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "http-backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8006)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("http-port"),
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.5.3"}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvcBackend, &meshSvcTCPBackend, &meshSvcHTTPBackend},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_tcp-backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP))).
				AddEndpoint("default_http-backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "http-backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP)))
			tcpRules := core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					kri.From(&meshSvcBackend): test_policies.NewOutboundRule(nil, api.Rule{
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{
								{
									TargetRef: builders.TargetRefMeshService(
										"tcp-backend", "", "tcp-port",
									),
									Weight: pointer.To(uint(1)),
								},
							},
						},
					}),
				},
			}

			httpRules := core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					kri.From(&meshSvcBackend): test_policies.NewOutboundRule(nil, meshhttproute_api.PolicyDefault{
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
											TargetRef: builders.TargetRefMeshService("http-backend", "", "http-port"),
											Weight:    pointer.To(uint(1)),
										},
									},
								},
							},
						},
					}),
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvcBackend), "tcp-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 1,
							Resource: kri.WithSectionName(kri.From(&meshSvcTCPBackend), "tcp-port"),
						},
						{
							Port:     builders.FirstOutboundPort + 2,
							Resource: kri.WithSectionName(kri.From(&meshSvcHTTPBackend), "http-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(&core_xds.DataplaneMetadata{}).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, tcpRules).
							WithToPolicy(meshhttproute_api.MeshHTTPRouteType, httpRules),
					).
					Build(),
			}
		}()),
	)
})

func meshContextForMeshExternalService(resources ...core_model.Resource) *xds_context.MeshContext {
	resourceStore := memory.NewStore()
	mesh := builders.Mesh().Build()
	err := resourceStore.Create(context.Background(), mesh, store.CreateByKey("default", core_model.NoMesh))
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
	)
	mc, err := meshContextBuilder.Build(context.Background(), "default")
	Expect(err).ToNot(HaveOccurred())

	return &mc
}

// zoneEgressDataplane is a Dataplane exposing an embedded zone egress listener, which is
// how MeshExternalServices become reachable through an egress.
func zoneEgressDataplane() *core_mesh.DataplaneResource {
	return builders.Dataplane().
		WithName("zone-egress-01").
		WithAddress("127.0.0.1").
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{{
				Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
				Address: "127.0.0.1",
				Port:    10002,
				Name:    "ze-port",
				State:   mesh_proto.Dataplane_Networking_Listener_Ready,
			}}
		}).Build()
}

// backendDataplane backs the "backend" MeshService of dppForMeshExternalService, so its
// cluster gets a non-empty ClusterLoadAssignment.
func backendDataplane() *core_mesh.DataplaneResource {
	return builders.Dataplane().
		WithName("backend-01").
		WithAddress("192.168.0.4").
		AddInbound(builders.Inbound().
			WithPort(80).
			WithTags(map[string]string{
				mesh_proto.ServiceTag:  "backend",
				mesh_proto.ProtocolTag: string(core_meta.ProtocolTCP),
			}),
		).Build()
}

func dppForMeshExternalService(mesList ...*meshexternalservice_api.MeshExternalServiceResource) (*builders.DataplaneBuilder, *core_xds.Proxy, *meshservice_api.MeshServiceResource) {
	backendMeshSvc := &meshservice_api.MeshServiceResource{
		Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
		Spec: &meshservice_api.MeshService{
			Selector: meshservice_api.Selector{
				DataplaneLabels: &common_api.LabelSelector{
					MatchLabels: &map[string]string{mesh_proto.ServiceTag: "backend"},
				},
			},
			Ports: []meshservice_api.Port{{
				Port:        80,
				TargetPort:  pointer.To(intstr.FromInt(80)),
				AppProtocol: core_meta.ProtocolTCP,
				Name:        pointer.To("tcp-port"),
			}},
		},
		Status: &meshservice_api.MeshServiceStatus{
			VIPs: []meshservice_api.VIP{{IP: "10.30.30.1"}},
		},
	}
	outbounds := xds_types.Outbounds{
		{
			Port:     builders.FirstOutboundPort,
			Resource: kri.WithSectionName(kri.From(backendMeshSvc), "tcp-port"),
		},
	}
	for _, mes := range mesList {
		outbounds = append(outbounds, &xds_types.Outbound{
			Address:  mes.Status.VIP.IP,
			Port:     uint32(mes.Spec.Match.Port),
			Resource: kri.From(mes),
		})
	}
	dp := builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")
	proxy := xds_builders.Proxy().
		WithDataplane(dp).
		WithOutbounds(outbounds).
		WithMetadata(&core_xds.DataplaneMetadata{
			SystemCaPath: "/tmp/ca-certs.crt",
		}).
		Build()

	return dp, proxy, backendMeshSvc
}
