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
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_server "github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

func getResource(resourceSet *core_xds.ResourceSet, typ envoy_resource.Type) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshHTTPRoute", func() {
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
			gatewayGenerator := plugin_gateway.NewGenerator("test-zone")
			_, err = gatewayGenerator.Generate(context.Background(), nil, given.xdsContext, given.proxy)

			Expect(err).NotTo(HaveOccurred())
			resourceSet := core_xds.NewResourceSet()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resourceSet, given.xdsContext, given.proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			Expect(getResource(resourceSet, envoy_resource.ListenerType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.EndpointType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.RouteType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".routes.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.SecretType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".secrets.golden.yaml")))
		},
		Entry("default-route", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			externalServices := xds_builders.EndpointMap().
				AddEndpoint("external-service", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8085).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "external-service", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithExternalServicesEndpointMap(externalServices).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("external-service", core_mesh.ProtocolHTTP).
					AddExternalService("external-service").
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
						{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
							Port: builders.FirstOutboundPort + 1,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "external-service",
							},
						}},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					Build(),
			}
		}()),
		Entry("default-route-outbound-with-tags-with-mtls", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(samples.MeshMTLSBuilder()).
					WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
					WithOutbounds(xds_types.Outbounds{
						{LegacyOutbound: builders.Outbound().
							WithPort(builders.FirstOutboundPort).
							WithTags(map[string]string{
								mesh_proto.ServiceTag: "backend",
								"region":              "us",
							}).Build()},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					Build(),
			}
		}()),
		Entry("default-meshservice", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "app", "backend")).
				AddEndpoint("backend_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "app", "backend"))
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  intstr.FromInt(8084),
						AppProtocol: core_mesh.ProtocolHTTP,
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
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(builders.Mesh().WithBuiltinMTLSBackend("builtin").WithEnabledMTLSBackend("builtin")).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					AddServiceProtocol("backend_svc_80", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithSecretsTracker(envoy.NewSecretsTracker(core_model.DefaultMesh, nil)).
					WithDataplane(builders.Dataplane().
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
						{
							Resource: pointer.To(core_model.NewTypedResourceIdentifier(&meshSvc, core_model.WithSectionName("80"))),
							Address:  "10.0.0.1",
							Port:     80,
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					Build(),
			}
		}()),
		Entry("default-meshmultizoneservice", func() outboundsTestCase {
			backendDP := samples.DataplaneWebBuilder().
				AddInbound(builders.Inbound().
					WithAddress("192.168.0.4").
					WithPort(8084).
					WithTags(map[string]string{
						mesh_proto.ServiceTag:  "backend",
						mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
						"app":                  "backend",
					}),
				).Build()
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  intstr.FromInt(8084),
						AppProtocol: core_mesh.ProtocolHTTP,
					}},
					Identities: []meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentityServiceTagType,
							Value: "backend",
						},
						{
							Type:  meshservice_api.MeshServiceIdentityServiceTagType,
							Value: "other-backend",
						},
					},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.1",
					}},
				},
			}
			meshMZSvc := meshmultizoneservice_api.MeshMultiZoneServiceResource{
				Meta: &test_model.ResourceMeta{Name: "multi-backend", Mesh: "default"},
				Spec: &meshmultizoneservice_api.MeshMultiZoneService{
					Selector: meshmultizoneservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						AppProtocol: core_mesh.ProtocolHTTP,
					}},
				},
				Status: &meshmultizoneservice_api.MeshMultiZoneServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.2",
					}},
					MeshServices: []meshmultizoneservice_api.MatchedMeshService{
						{
							Name: "backend",
							Mesh: "default",
						},
					},
				},
			}

			dp := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
				Build()
			mc := meshContextWithResources(dp, backendDP, &meshSvc, &meshMZSvc)

			builder := &sync.DataplaneProxyBuilder{
				Zone:       "zone-1",
				APIVersion: envoy.APIV3,
			}
			proxy, err := builder.Build(context.Background(), core_model.ResourceKey{Name: dp.GetMeta().GetName(), Mesh: dp.GetMeta().GetMesh()}, *mc)
			Expect(err).ToNot(HaveOccurred())

			proxy.Outbounds = xds_types.Outbounds{{
				Address:  "10.0.0.2",
				Port:     80,
				Resource: pointer.To(core_model.NewTypedResourceIdentifier(&meshMZSvc, core_model.WithSectionName("80"))),
			}}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("default-meshexternalservice", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: meshexternalservice_api.HttpProtocol,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
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
			mc := meshContextWithResources(dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("httproute-meshexternalservice", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: meshexternalservice_api.HttpProtocol,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
					},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{
						IP: "10.20.20.1",
					},
				},
			}

			dp, proxy := dppForMeshExternalService(&meshExtSvc)
			proxy.Policies = core_xds.MatchedPolicies{
				Dynamic: core_xds.PluginOriginatedPolicies{},
			}
			proxy.Policies.Dynamic[api.MeshHTTPRouteType] = xds.TypedMatchingPolicies{
				Type: api.MeshHTTPRouteType,
				ToRules: core_rules.ToRules{
					ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
						backendMeshExternalServiceIdentifier: {
							Origin: []core_rules.Origin{
								{Resource: &test_model.ResourceMeta{Mesh: "default", Name: "http-route"}},
							},
							BackendRefOriginIndex: map[core_rules.MatchesHash]int{
								core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}})): 0,
							},
							Conf: []interface{}{
								api.PolicyDefault{
									Rules: []api.Rule{
										{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefMeshExternalService("example"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			egress := builders.ZoneEgress().WithPort(10002).Build()
			mc := meshContextWithResources(dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("meshexternalservice-with-tls", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: meshexternalservice_api.HttpProtocol,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
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
			mc := meshContextWithResources(dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("meshexternalservice-with-tls-and-skipall", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: meshexternalservice_api.HttpProtocol,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
					},
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
						Verification: &meshexternalservice_api.Verification{
							Mode: pointer.To(meshexternalservice_api.TLSVerificationSkipAll),
						},
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
			mc := meshContextWithResources(dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("meshexternalservice-with-tls-and-custom-settings", func() outboundsTestCase {
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "example", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
						Port:     9090,
						Protocol: meshexternalservice_api.HttpProtocol,
					},
					Endpoints: []meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    pointer.To(meshexternalservice_api.Port(10000)),
						},
						{
							Address: "example2.com",
							Port:    pointer.To(meshexternalservice_api.Port(11111)),
						},
					},
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
						Verification: &meshexternalservice_api.Verification{
							ServerName: pointer.To("example2.com"),
							SubjectAltNames: &[]meshexternalservice_api.SANMatch{{
								Type:  meshexternalservice_api.SANMatchPrefix,
								Value: "example",
							}, {
								Type:  meshexternalservice_api.SANMatchExact,
								Value: "example2.com",
							}},
							CaCert: &common_api.DataSource{
								InlineString: pointer.To("ca"),
							},
							ClientCert: &common_api.DataSource{
								InlineString: pointer.To("cert"),
							},
							ClientKey: &common_api.DataSource{
								InlineString: pointer.To("key"),
							},
						},
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
			mc := meshContextWithResources(dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("basic", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{{
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}, {
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v2",
												},
											}, {
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v3",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefServiceSubset("backend", "region", "us"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}, {
											Matches: []api.Match{{
												QueryParams: []api.QueryParamsMatch{{
													Type:  api.ExactQueryMatch,
													Name:  "v1",
													Value: "true",
												}},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									},
								}},
							}),
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
				AddEndpoint("backend_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "app", "backend")).
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						builders.Dataplane().
							WithName("web-01").
							WithAddress("192.168.0.2").
							WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: pointer.To(core_model.NewTypedResourceIdentifier(&meshSvc, core_model.WithSectionName("test-port"))),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
									backendMeshServiceIdentifier: {
										Origin: []core_rules.Origin{
											{Resource: &test_model.ResourceMeta{Mesh: "default", Name: "http-route"}},
										},
										BackendRefOriginIndex: map[core_rules.MatchesHash]int{
											core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}})): 0,
											core_rules.MatchesHash(api.HashMatches([]api.Match{
												{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v2"}},
												{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v3"}},
											})): 0,
											core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v4"}}})): 0,
										},
										Conf: []interface{}{
											api.PolicyDefault{
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/v1",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}, {
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/v2",
														},
													}, {
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/v3",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefServiceSubset("backend", "region", "us"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}, {
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/v4",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefMeshService("backend", "", "test-port"),
															Weight:    pointer.To(uint(100)),
															Port:      pointer.To(uint32(80)),
														}},
													},
												}},
											},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("match-priority", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{{
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}, {
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.Exact,
													Value: "/v1/specific",
												},
											}, {
												Method: pointer.To(api.Method("GET")),
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									},
								}},
							}),
					).
					Build(),
			}
		}()),
		Entry("mixed-tcp-and-http-outbounds", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("other-tcp", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-tcp", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "eu"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("other-tcp", core_mesh.ProtocolTCP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{{
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									},
								}},
							}),
					).
					Build(),
			}
		}()),
		Entry("request-header-modifiers", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestHeaderModifierType,
														RequestHeaderModifier: &api.HeaderModifier{
															Add: []api.HeaderKeyValue{{
																Name:  "request-add-header",
																Value: "add-value",
															}},
															Set: []api.HeaderKeyValue{{
																Name:  "request-set-header",
																Value: "set-value",
															}, {
																Name:  "request-set-header-multiple",
																Value: "one-value,second-value",
															}},
															Remove: []string{
																"request-header-to-remove",
															},
														},
													}},
												},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("response-header-modifiers", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.ResponseHeaderModifierType,
														ResponseHeaderModifier: &api.HeaderModifier{
															Add: []api.HeaderKeyValue{{
																Name:  "response-add-header",
																Value: "add-value",
															}},
															Set: []api.HeaderKeyValue{{
																Name:  "response-set-header",
																Value: "set-value",
															}},
															Remove: []string{
																"response-header-to-remove",
															},
														},
													}},
												},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("request-redirect", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestRedirectType,
														RequestRedirect: &api.RequestRedirect{
															Scheme: pointer.To("other"),
														},
													}},
												},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("url-rewrite", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
							Rules: core_rules.Rules{
								{
									Subset: core_rules.MeshService("backend"),
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												Filters: &[]api.Filter{{
													Type: api.URLRewriteType,
													URLRewrite: &api.URLRewrite{
														Path: &api.PathRewrite{
															Type:               api.ReplacePrefixMatchType,
															ReplacePrefixMatch: pointer.To("/v2"),
														},
													},
												}},
											},
										}},
									},
								},
							},
						}),
					).
					Build(),
			}
		}()),
		Entry("headers-match", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Headers: []common_api.HeaderMatch{{
														Type:  pointer.To(common_api.HeaderMatchExact),
														Name:  "foo-exact",
														Value: "bar",
													}, {
														Type: pointer.To(common_api.HeaderMatchPresent),
														Name: "foo-present",
													}, {
														Type:  pointer.To(common_api.HeaderMatchRegularExpression),
														Name:  "foo-regex",
														Value: "x.*y",
													}, {
														Type: pointer.To(common_api.HeaderMatchAbsent),
														Name: "foo-absent",
													}, {
														Type:  pointer.To(common_api.HeaderMatchPrefix),
														Name:  "foo-prefix",
														Value: "x",
													}},
												}},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("grpc-service", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolGRPC, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolGRPC).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: builders.TargetRefService("backend"),
														Weight:    pointer.To(uint(100)),
													}},
												},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("request-mirror", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("payments", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8086).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "payments", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us", "version", "v1", "env", "dev"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("payments", core_mesh.ProtocolHTTP).
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
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{
														{
															Type: api.RequestMirrorType,
															RequestMirror: &api.RequestMirror{
																Percentage: pointer.To(intstr.FromString("99.9")),
																BackendRef: common_api.BackendRef{
																	TargetRef: common_api.TargetRef{
																		Kind: common_api.MeshServiceSubset,
																		Name: "payments",
																		Tags: map[string]string{
																			"version": "v1",
																			"region":  "us",
																			"env":     "dev",
																		},
																	},
																},
															},
														},
														{
															Type: api.RequestMirrorType,
															RequestMirror: &api.RequestMirror{
																BackendRef: common_api.BackendRef{
																	TargetRef: common_api.TargetRef{
																		Kind: common_api.MeshService,
																		Name: "backend",
																	},
																},
															},
														},
													},
												},
											}},
										},
									},
								},
							}),
					).
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
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8082,
								Hostname: "*.dev",
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_TCP,
								Port:     9080,
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"),
				)
			xdsContext := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				WithEndpointMap(outboundTargets).Build()

			commonRules := core_rules.Rules{
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/wild",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Hostnames: []string{"go.dev"},
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/go-dev",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Hostnames: []string{"*.dev"},
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/wild-dev",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Hostnames: []string{"other.dev"},
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/other-dev",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
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
							WithGatewayPolicy(api.MeshHTTPRouteType, core_rules.GatewayRules{
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[rules.InboundListenerHostname]rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"):      commonRules,
										core_rules.NewInboundListenerHostname("192.168.0.1", 8081, "go.dev"): commonRules,
										core_rules.NewInboundListenerHostname("192.168.0.1", 8082, "*.dev"):  commonRules,
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("gateway-listener-specific", func() outboundsTestCase {
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"),
				)
			xdsContext := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				WithEndpointMap(outboundTargets).Build()
			return outboundsTestCase{
				xdsContext: *xdsContext,
				proxy: xds_builders.Proxy().
					WithDataplane(samples.GatewayDataplaneBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithGatewayPolicy(api.MeshHTTPRouteType, core_rules.GatewayRules{
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[rules.InboundListenerHostname]rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/wild",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8081, "go.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"go.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/go-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("gateway-listener-different-hostnames-specific", func() outboundsTestCase {
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
								Hostname: "other.dev",
								Tags: map[string]string{
									"hostname": "other",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8080,
								Hostname: "go.dev",
								Tags: map[string]string{
									"hostname": "go",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8081,
								Tags: map[string]string{
									"hostname": "wild",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Hostname: "*",
								Port:     8082,
								Tags: map[string]string{
									"hostname": "wild",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTPS,
								Hostname: "*.secure.dev",
								Port:     8083,
								Tags: map[string]string{
									"hostname": "secure",
								},
								Tls: &mesh_proto.MeshGateway_TLS_Conf{
									Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
									Certificates: []*system_proto.DataSource{{
										Type: &system_proto.DataSource_Inline{
											Inline: wrapperspb.Bytes([]byte(secureSecret)),
										},
									}},
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTPS,
								Hostname: "*.super-secure.dev",
								Port:     8083,
								Tags: map[string]string{
									"hostname": "super-secure",
								},
								Tls: &mesh_proto.MeshGateway_TLS_Conf{
									Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
									Certificates: []*system_proto.DataSource{{
										Type: &system_proto.DataSource_Inline{
											Inline: wrapperspb.Bytes([]byte(superSecureSecret)),
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"),
				)
			xdsContext := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				WithEndpointMap(outboundTargets).Build()
			return outboundsTestCase{
				xdsContext: *xdsContext,
				proxy: xds_builders.Proxy().
					WithDataplane(samples.GatewayDataplaneBuilder()).
					WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithGatewayPolicy(api.MeshHTTPRouteType, core_rules.GatewayRules{
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[rules.InboundListenerHostname]rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "other.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/to-other-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "go.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/to-go-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8081, ""): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/wild-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8082, ""): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/same-path",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend-wild"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}, {
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/same-path",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend-wild-dev"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8083, "*.secure.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"first-specific.secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/first-specific-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}, {
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"second-specific.secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/second-specific-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8083, "*.super-secure.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"first-specific.super-secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/first-specific-super-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}, {
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"second-specific.super-secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/second-specific-super-dev",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
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
const secureSecret = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA60QMsTAL8jPI+XzWlVv4e7Gc8C5Y5q5SHDMuXGEog2eyA+UB
0V3hhoNj+pz3vDSW71bRnl9otUi82jjvaZyOmvUTIwI2yLuFUqfZSwYYxoygcTQ6
zbANWas6qsElWAspgIPefsed3i44FazTMjLwbSAGdr0UDZyPm8Uh5xme2E24KDeS
tIBcxnCAKfdCVFFxKsrwFe8FaKl7sAQtWW4d9FuyVvXj9zGx2azELlmGKtB0nDC0
qyPNJw8OWZHu/snxiwhh4lURw1MgDDQJk/yTeT/dS37Y299syp36RUwkN5sG7CYU
tWHJ/dCL53dnKu2KvlKVDOO5GV07SRIB7uILWwIDAQABAoIBAFm0t9Y0CMoQXthq
dnO6/tNDVSDruzAyVdK03T+SOF1qg6Eih//p+R1OKigcBAY6Uzbtdr4ZiRZepsva
m8c8T8/cFLDrnjIJ9nsezybhKz9Bzcd8b9OQBncjaBpFzVR15RxAq+zRdmuKWg5B
uMHSVIR3ip9p1ySdhtCRaSzyQvQcay1MMLOgW99XiNmQFpTRFAZyc7olZRQyH3u8
CUItZIBotUTHyYmTQsnTq+0iint9Ag2mwQBU59qVKwhEtikjP8jtY3cBeAe3+Nja
+uNrvFJ85KkgCK6L03//G7WCZrjEGUt5sJoWxVdrK8p1fRyBFyIwpdCwVjukEmRl
WWxv7XECgYEA9NuEppSVVLjxogt8FjtWhIDGQfIlr0CFUxE36gMAFXSpwFIFY2Tb
SINAvckJU7zIvzTY9cVYx3E/2JFgc04fjQc8cJ+yDeuFd/mLdPmNYxqZ5keBkM6L
rrccepB40Z55vz8FxuNvWq15xSGWzOswudfhP9V2/FgIlJUg5/p7RFkCgYEA9fjK
6j3XeDZwTuj8Ji1ex4B+Di0JOBAIqmtv1lA5qASgLKoCNjP6OkieL0NZEpmN0vuq
KpOnTtx40nUAEctOr0HHfgaqhhw3VUSvKPsInQKrwGRFOGm6NSPB9db1hDXp7aRD
zOfOEpld54yGlqSaWNJIpEJ6HhFoqyoqg3HPptMCgYAbyiY9+bMREIRsDb2hkE57
b1oQ9fiM8VewW83qwzhpNvplF2oBI9s3WZ4pa/2hAVYPTWIqUqGG0TWb0LQPohg2
m1Givp0os0hMm4fWWNRRIR3CYu8zjh2QULvstSThNYk/yVlQf1OOCQ4+71b8Ht1C
2lt4MTP148/lfR9k9Kq00QKBgQCh3PnzEYUUh4Z6dxlPKjYfxO+u9nYFnY+GTjMH
bj2y0nBxU+MmtiepaRYndgNMmR3aRGBjqkzEOZOMsw+7pfV+oSPdTBe1LyY+h3dY
2XF+mT5a2eEvUWwHAiPmWnGwciYhiyJO2hAi7yf7ct8yjNlBMAg7h7+Cv+QIFzRo
0WFbnwKBgC/rFGOSfw1PMvjUziBM65a+J1PWCy42kY2LLmFjAUEf7DOEcvtXxpP/
qHqux6FapobFiSg5jo0Tos7U18Ri2YeoZ+q1aq2h10p4PTY1wX2ZHDQ1vQ6VH8Qv
jyYbKjRxhqr4i8pY6sVq9Og/SdZCqqwaSvJ8e9tX29j3DlzzlKot
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIC/TCCAeWgAwIBAgIQA1hlk2vouDoqB0cJ0Y9k0TANBgkqhkiG9w0BAQsFADAA
MB4XDTI0MDUxNzE4NTA1NFoXDTM0MDUxNTE4NTA1NFowADCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAOtEDLEwC/IzyPl81pVb+HuxnPAuWOauUhwzLlxh
KINnsgPlAdFd4YaDY/qc97w0lu9W0Z5faLVIvNo472mcjpr1EyMCNsi7hVKn2UsG
GMaMoHE0Os2wDVmrOqrBJVgLKYCD3n7Hnd4uOBWs0zIy8G0gBna9FA2cj5vFIecZ
nthNuCg3krSAXMZwgCn3QlRRcSrK8BXvBWipe7AELVluHfRbslb14/cxsdmsxC5Z
hirQdJwwtKsjzScPDlmR7v7J8YsIYeJVEcNTIAw0CZP8k3k/3Ut+2NvfbMqd+kVM
JDebBuwmFLVhyf3Qi+d3Zyrtir5SlQzjuRldO0kSAe7iC1sCAwEAAaNzMHEwDgYD
VR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMB
Af8wHQYDVR0OBBYEFK8bFa0edq1+GeQXW0Glohw/oUYPMBoGA1UdEQEB/wQQMA6C
DCouc2VjdXJlLmRldjANBgkqhkiG9w0BAQsFAAOCAQEAjR0BiLBOOdI0iyYxoM8j
4qnJ4kaBNhD97mm10iQN4EO7IaJUfO52YcupTtG1gQs61me5JRX/FofDOH+vvsr4
VA7ksYpcAh8mpIB67KD+eZXI7SjdO+ERsKZbIK38mUcpwc9uvBhXPQwC0hstR7BR
9VXx1xS/tTFb4U2u2mfieWFcxIAIINk2Wv9RcUxEdYjI0KS39Qt/lCOC7V+/ddvt
aUVFd7keE7LzSYdlDGbrnnPZPMcDd5dAyGGJBwIl74jHY1uXfGE+oi3AQJK0key5
hiU4YshCF/6LfC+bnH2rfAW8285BQIkK/QACLuIGGGnrybDVhO/IF5owhXybRiWb
Jw==
-----END CERTIFICATE-----
`

// nolint:gosec
const superSecureSecret = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAy2J1y0ehmGKvF9m1zooZr9UUgg3Y/xrhYXT47BLtKNfok5sI
ELbUlAKmwcCFdenOH2LwoN99ty9VXpqawiwMS/orPhRf3Z4GSMnWeRZE0ltawu1v
AypUnMw5HEy5gvJCnzLWfi/ZkBlb5Kaos/1WTuauvM75FvDfadtBcq3MHd7hZF4k
FVTMfdhDIVG1fsK8OeVabI5C1KMwdXz4JU3ZAxO1ip3PREJaXgQQXAHNZCDoPJnE
jvcPKjjNXpAbcKckFQoiNxg9m5RAMzYTlBosCemzK2QOcLcKlTv33TnDBUCysY4G
//E/dinpHzFxDhewI5k6H8gtXriD2Dx6qZ2ktwIDAQABAoIBAHL6tNE5K7f0gjwf
jlK3bBIlijSEE5sU3Tm1YUxE3uJqPUfFK2gXlFIgtZlvd4PTq/2+d37bGj1HeyHC
kZ8YO9NwGKY96nxla/QWdzN5TWsGzmbIyCun8LG8GsVO8sl+n/9UREKOVpbKX1MC
jPlETVjJvRtsfxFjF0rG81cbvftcFIwD5d91u53bq9OXzYZcWO3aFFS788J/0CI8
2y3KWb8VZrrIpBfU/LhFWXA8NzSuG+abLqUjF3zn2iMWnxS4asWFFmCdl4pytqrb
7g+ndV2Ab5B1P8LZ3hT5QYIPGzbGBDkQjjC44IuHI4xERkVpWfdaNqk3OQqGjORz
cZYD1oECgYEA9wuyjItpT4jCluYtAyTXqDDG2qE1OlmY6rTj4d2Z+DFbKn2wYsmt
FbTeSneqWt+u+2E3tvHbxI1bMFaSjEM8MAdqbRlj/WC1PUaa+iSHpabpZ4+F92kZ
ME4Olgw/Swyr9cmZc39hrCFw136r6aCz1YFC8deLZTd6cFLEYH/y6mECgYEA0sGj
N+MOY0xJ2yp+arKf1nvTNgy14XPCqD28zL5tJB9ZWCAnfF4brvL0iY9z7rfCo5Gq
A8Wx/7oK4LYCi/1fPPo1evQGlZq/KhOurLy22TZPtOskeC2C5hn3LhUFC5oazfcc
EAuksj3auZNPy+AzyA5qXIuGE7DJlbDi+rbOVhcCgYEAs5jRaNNA8A0gSct0FcEG
9sLfDbn8lDrmrFptAJq1gHWBLVbKkEbFie4/XCu6sO3ErAN1GY6ikjWhgXaue0G1
08TQXhgDVQSlPbLCn+9GnerF6/+vCLpjEXbtq6+jo8/Gg5zX7dtBCn4VJtRz7hhi
JGkgXeiw8hhu7pF9KhpaYoECgYB3g283Rf7muGA5dINzpg+V5WoEgHizfJ2qIjhq
MqJZlZ1op/M8R3GTaBrb1wl7GaG6d+Pdd8JUrf91JkGTeP8E6S5ipvcE51f4WGj5
c5qM2ougoKdxrv1H1vmgnDLcPWtt2O+E+dVPblwWWD8r8dvrWqFeEZDaoanuxPwy
CHBByQKBgQCow4u6R0nWDNjMgoZI8EOB8OVuQnuMa1qkI/Ubpyl/l9jGQSETbfwt
3iAws+5Q6/oRt2y4fj/1xvon6u3UgN6VZW2kkOS+owkyLnPpeWQ6tvyMAMN7QJBJ
IX4MZllA8W6IHRBqu31PLHn9yOtPXwwud7IfqGP0zqxycbl8qxpAzw==
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIDBDCCAeygAwIBAgIRAOXUzPeaViYNAsDUXjtyIekwDQYJKoZIhvcNAQELBQAw
ADAeFw0yNDA1MTcxODUwMjdaFw0zNDA1MTUxODUwMjdaMAAwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQDLYnXLR6GYYq8X2bXOihmv1RSCDdj/GuFhdPjs
Eu0o1+iTmwgQttSUAqbBwIV16c4fYvCg3323L1VemprCLAxL+is+FF/dngZIydZ5
FkTSW1rC7W8DKlSczDkcTLmC8kKfMtZ+L9mQGVvkpqiz/VZO5q68zvkW8N9p20Fy
rcwd3uFkXiQVVMx92EMhUbV+wrw55VpsjkLUozB1fPglTdkDE7WKnc9EQlpeBBBc
Ac1kIOg8mcSO9w8qOM1ekBtwpyQVCiI3GD2blEAzNhOUGiwJ6bMrZA5wtwqVO/fd
OcMFQLKxjgb/8T92KekfMXEOF7AjmTofyC1euIPYPHqpnaS3AgMBAAGjeTB3MA4G
A1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTAD
AQH/MB0GA1UdDgQWBBSQiw+6HCFLxc081Az9pk0UWpDEMDAgBgNVHREBAf8EFjAU
ghIqLnN1cGVyLXNlY3VyZS5kZXYwDQYJKoZIhvcNAQELBQADggEBAAIy6oTyJMye
ujgeyN7pZjbxvMouBBSF35xdxD6+GoA39fIHQyi3fuEOcs5WZlf8YrwAcS6CkLFH
C/7XdJJ93XyS7X7i6fQZ04JbHbKu0Nq3iexcHEXo0WdUL/0ZTLqjj5Xi7f5UTwq8
Ero/V00bq2iENdfRKHxa8HXb6G2OQbNvI5cDaJQw5N+Nlzfio8kc/kwrSPv7crA4
KyoUDOJWLLzLYGkRLO1wL+kEluCKMENSdwtb8gTHigwa0RjB45h4reTEPgyxcKXR
oESyXXAeWPJX3e7ZgdjUHomwhAZpUmqIWribTioaHZTb1I6OpsD+eF6USSayxUaL
9/atNWDDBSk=
-----END CERTIFICATE-----
`

func meshContextWithResources(resources ...core_model.Resource) *xds_context.MeshContext {
	resourceStore := memory.NewStore()

	mesh := builders.Mesh().WithBuiltinMTLSBackend("ca-1").WithEgressRoutingEnabled().WithEnabledMTLSBackend("ca-1").Build()
	err := resourceStore.Create(context.Background(), mesh, store.CreateByKey("default", core_model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	for _, res := range resources {
		err = resourceStore.Create(context.Background(), res, store.CreateByKey(res.GetMeta().GetName(), res.GetMeta().GetMesh()))
	}
	Expect(err).ToNot(HaveOccurred())

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
		false,
	)
	mc, err := meshContextBuilder.Build(context.Background(), "default")
	Expect(err).ToNot(HaveOccurred())

	return &mc
}

func dppForMeshExternalService(mes *meshexternalservice_api.MeshExternalServiceResource) (*builders.DataplaneBuilder, *core_xds.Proxy) {
	dp := builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")
	proxy := xds_builders.Proxy().
		WithDataplane(dp).
		WithOutbounds(xds_types.Outbounds{
			{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Port: builders.FirstOutboundPort,
				Tags: map[string]string{
					mesh_proto.ServiceTag: "backend",
				},
			}},
			{
				Address:  "10.20.20.1",
				Port:     9090,
				Resource: pointer.To(core_model.NewTypedResourceIdentifier(mes)),
			},
		}).
		WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
		WithMetadata(&core_xds.DataplaneMetadata{
			SystemCaPath: "/tmp/ca-certs.crt",
		}).
		Build()

	return dp, proxy
}
