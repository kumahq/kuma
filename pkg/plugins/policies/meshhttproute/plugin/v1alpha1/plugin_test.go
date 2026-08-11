package v1alpha1_test

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"time"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
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
	meshmultizoneservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v3/pkg/metrics"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
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
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	xds_server "github.com/kumahq/kuma/v3/pkg/xds/server"
	"github.com/kumahq/kuma/v3/pkg/xds/sync"
)

var _ = Describe("MeshHTTPRoute", func() {
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
	unifiedNaming := func() *core_xds.DataplaneMetadata { return &core_xds.DataplaneMetadata{} }
	meshServiceSpiffeIdentities := func(values ...string) *[]meshservice_api.MeshServiceIdentity {
		identities := make([]meshservice_api.MeshServiceIdentity, 0, len(values))
		for _, value := range values {
			identities = append(identities, meshservice_api.MeshServiceIdentity{
				Type:  meshservice_api.MeshServiceIdentitySpiffeIDType,
				Value: value,
			})
		}
		return &identities
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

			secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None())
			dataSourceLoader := datasource.NewDataSourceLoader(secretManager)
			given.xdsContext.Mesh.DataSourceLoader = dataSourceLoader

			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), given.xdsContext.Mesh, given.proxy)).To(Succeed(), n)
			}
			resourceSet := core_xds.NewResourceSet()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resourceSet, given.xdsContext, given.proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			resource, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.EndpointType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.RouteType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".routes.golden.yaml")))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.SecretType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".secrets.golden.yaml")))
		},
		Entry("default-route", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
				},
			}
			meshExtSvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{Name: "external-service", Mesh: "default"},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     8085,
						Protocol: core_meta.ProtocolHTTP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{{
						Address: "192.168.0.5",
						Port:    8085,
					}},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{IP: "10.20.20.1"},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}
			resources.MeshLocalResources[meshexternalservice_api.MeshExternalServiceType] = &meshexternalservice_api.MeshExternalServiceResourceList{
				Items: []*meshexternalservice_api.MeshExternalServiceResource{&meshExtSvc},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us")).
				AddEndpoint("default_external-service___extsvc_8085", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8085).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{OwnerResource: kri.From(&meshExtSvc)}).
					WithTags(mesh_proto.ServiceTag, "external-service", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					AddServiceProtocol("default_backend___msvc_80", core_meta.ProtocolHTTP).
					AddServiceProtocol("default_external-service___extsvc_8085", core_meta.ProtocolHTTP).
					AddExternalService("default_external-service___extsvc_8085").
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
						{
							Address:  "10.20.20.1",
							Port:     8085,
							Resource: kri.WithSectionName(kri.From(&meshExtSvc), "8085"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithInternalAddresses(core_xds.InternalAddress{AddressPrefix: "192.168.0.0", PrefixLen: 16}, core_xds.InternalAddress{AddressPrefix: "::1", PrefixLen: 128}).
					Build(),
			}
		}()),
		Entry("default-route-with-mtls", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(samples.MeshMTLSBuilder()).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					AddServiceProtocol("default_backend___msvc_80", core_meta.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					Build(),
			}
		}()),
		Entry("default-meshservice", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
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
					AddServiceProtocol("default_backend___svc_80", core_meta.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
							Address:  "10.0.0.1",
							Port:     80,
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					Build(),
			}
		}()),
		Entry("default-meshservice-mesh-scoped-zone", func() outboundsTestCase {
			// MeshService located in a remote zone that exposes a mesh-scoped zone proxy.
			// The cluster SNI must use the KRI-derived format (sni.msvc.<mesh>.<zone>.<name>.<port>)
			// instead of the legacy hash-based format.
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend__remote-zone_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend", Mesh: "default",
					Labels: map[string]string{
						mesh_proto.ZoneTag:             "remote-zone",
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{{
						Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						Address:  "10.0.0.1",
						Port:     80,
					}}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithWorkloadIdentity(&core_xds.WorkloadIdentity{
						IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
							return bldrs_tls.SdsSecretConfigSource(
								"identity_cert:secret:default",
								bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
							)
						},
					}).
					Build(),
			}
		}()),
		Entry("httproute-meshservice-mesh-scoped-zone-port-by-number", func() outboundsTestCase {
			// A backendRef addressing a named port by number must still produce the
			// port-name SNI, otherwise it matches no filter chain on the zone proxy.
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend", Mesh: "default",
					Labels: map[string]string{
						mesh_proto.ZoneTag:             "remote-zone",
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
						"app":                          "backend",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}
			dpBuilder := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")
			mc := meshContextWithResources(builders.Mesh(), dpBuilder.Build(), &meshSvc)
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend__remote-zone_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshContext(mc).
					WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(dpBuilder).
					WithOutbounds(xds_types.Outbounds{{
						Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						Address:  "10.0.0.1",
						Port:     80,
					}}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									kri.From(&meshSvc): test_policies.NewOutboundRule(meshSvc.Meta, api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/with-retry",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: common_api.TargetRef{
														Kind:   common_api.MeshService,
														Labels: &map[string]string{"app": "backend"},
													},
													Weight: pointer.To(uint(100)),
													Port:   pointer.To(uint32(80)),
												}},
											},
										}},
									}),
								},
							}),
					).
					WithWorkloadIdentity(&core_xds.WorkloadIdentity{
						IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
							return bldrs_tls.SdsSecretConfigSource(
								"identity_cert:secret:default",
								bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
							)
						},
					}).
					Build(),
			}
		}()),
		Entry("default-meshexternalservice-mesh-scoped-zone", func() outboundsTestCase {
			// MeshExternalService in a remote zone that exposes a mesh-scoped zone proxy.
			// The cluster SNI must use the KRI-derived format (sni.extsvc.<mesh>.<zone>.<name>.<port>).
			meshExtSvc := &meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{
					Name:   "ext-backend",
					Mesh:   "default",
					Labels: map[string]string{mesh_proto.ZoneTag: "remote-zone"},
				},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9000,
						Protocol: core_meta.ProtocolHTTP,
					},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{IP: "10.20.20.1"},
				},
			}
			extSvcKRI := kri.From(meshExtSvc)
			const mesServiceName = "default_ext-backend__remote-zone_extsvc_9000"

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint(mesServiceName, xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(27017).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{OwnerResource: extSvcKRI}))

			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshexternalservice_api.MeshExternalServiceType] = &meshexternalservice_api.MeshExternalServiceResourceList{
				Items: []*meshexternalservice_api.MeshExternalServiceResource{meshExtSvc},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(builders.Mesh()).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					AddExternalService(mesServiceName).
					AddServiceProtocol(mesServiceName, core_meta.ProtocolHTTP).
					With(func(ctx *xds_context.Context) {
						ctx.Mesh.ZoneEgresses = []core_xds.ZoneEgressInstance{
							{Address: "10.0.0.1", Port: 10002, SAN: "spiffe://default/zone-egress"},
						}
					}).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web"),
					).
					WithOutbounds(xds_types.Outbounds{{
						Resource: kri.WithSectionName(extSvcKRI, "9000"),
						Address:  "10.20.20.1",
						Port:     9000,
					}}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(&core_xds.DataplaneMetadata{}).
					WithWorkloadIdentity(&core_xds.WorkloadIdentity{
						IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
							return bldrs_tls.SdsSecretConfigSource(
								"identity_cert:secret:default",
								bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
							)
						},
					}).
					Build(),
			}
		}()),
		Entry("default-meshmultizoneservice-mesh-scoped-zone", func() outboundsTestCase {
			// MeshMultiZoneService (global, no zone) with mesh-scoped proxy enabled for zone "".
			// The cluster SNI must use the KRI-derived format (sni.mzsvc.<mesh>.<name>.<port>).
			backendDP := builders.Dataplane().
				WithName("backend").
				WithAddress("192.168.0.4").
				AddInbound(builders.Inbound().
					WithPort(8084).
					WithTags(map[string]string{
						mesh_proto.ServiceTag:  "backend",
						mesh_proto.ProtocolTag: string(core_meta.ProtocolHTTP),
						"app":                  "backend",
					}),
				).Build()
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend", Mesh: "default",
					Labels: map[string]string{
						"service": "backend",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{
						DataplaneLabels: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						},
					},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
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
					Selector: meshmultizoneservice_api.Selector{
						MeshService: common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"service": "backend",
							},
						},
					},
					Ports: []meshmultizoneservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
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
				WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web").
				Build()
			mc := meshContextWithResources(builders.Mesh(), dp, backendDP, &meshSvc, &meshMZSvc)

			builder := &sync.DataplaneProxyBuilder{
				Zone:       "zone-1",
				APIVersion: envoy.APIV3,
			}
			proxy, err := builder.Build(context.Background(), core_model.ResourceKey{Name: dp.GetMeta().GetName(), Mesh: dp.GetMeta().GetMesh()}, &core_xds.DataplaneMetadata{}, *mc)
			Expect(err).ToNot(HaveOccurred())

			proxy.Outbounds = xds_types.Outbounds{{
				Address:  "10.0.0.2",
				Port:     80,
				Resource: kri.WithSectionName(kri.From(&meshMZSvc), "80"),
			}}
			proxy.WorkloadIdentity = &core_xds.WorkloadIdentity{
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"identity_cert:secret:default",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshContext(mc).
					Build(),
				proxy: proxy,
			}
		}()),
		Entry("default-meshmultizoneservice", func() outboundsTestCase {
			backendDP := builders.Dataplane().
				WithName("backend").
				WithAddress("192.168.0.4").
				AddInbound(builders.Inbound().
					WithPort(8084).
					WithTags(map[string]string{
						mesh_proto.ServiceTag:  "backend",
						mesh_proto.ProtocolTag: string(core_meta.ProtocolHTTP),
						"app":                  "backend",
					}),
				).Build()
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend", Mesh: "default",
					Labels: map[string]string{
						"service": "backend",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{
						DataplaneLabels: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						},
					},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
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
					Selector: meshmultizoneservice_api.Selector{
						MeshService: common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"service": "backend",
							},
						},
					},
					Ports: []meshmultizoneservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
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
				WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web").
				Build()
			mc := meshContextWithResources(builders.Mesh(), dp, backendDP, &meshSvc, &meshMZSvc)

			builder := &sync.DataplaneProxyBuilder{
				Zone:       "zone-1",
				APIVersion: envoy.APIV3,
			}
			proxy, err := builder.Build(context.Background(), core_model.ResourceKey{Name: dp.GetMeta().GetName(), Mesh: dp.GetMeta().GetMesh()}, &core_xds.DataplaneMetadata{}, *mc)
			Expect(err).ToNot(HaveOccurred())

			proxy.Outbounds = xds_types.Outbounds{{
				Address:  "10.0.0.2",
				Port:     80,
				Resource: kri.WithSectionName(kri.From(&meshMZSvc), "80"),
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
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolHTTP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    10000,
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
			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, zoneEgressDataplane())

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
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolHTTP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    10000,
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
			proxy.Policies.Dynamic[api.MeshHTTPRouteType] = core_xds.TypedMatchingPolicies{
				Type: api.MeshHTTPRouteType,
				ToRules: core_rules.ToRules{
					ResourceRules: map[kri.Identifier]outbound.ResourceRule{
						backendMeshExternalServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
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
						}),
					},
				},
			}

			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, zoneEgressDataplane())

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
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
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.1",
					}},
				},
			}
			meshSvcUs := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend-us", Mesh: "default"},
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
						IP: "10.0.0.9",
					}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc, &meshSvcUs},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("default_backend___msvc_80",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us")).
				AddEndpoints("default_backend-us___msvc_80",
					xds_builders.Endpoint().
						WithTarget("192.168.0.6").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend-us", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.7").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend-us", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
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
							Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefMeshService("backend", "", "test-port"),
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
													TargetRef: builders.TargetRefMeshService("backend-us", "", "test-port"),
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
										}, {
											Matches: []api.Match{{
												QueryParams: &[]api.QueryParamsMatch{{
													Type:  api.ExactQueryMatch,
													Name:  "v1",
													Value: "true",
												}},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefMeshService("backend", "", "test-port"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("basic-real-meshservice-labels", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend",
					Mesh: "default",
					Labels: map[string]string{
						"app":     "backend",
						"version": "first",
					},
				},
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
			meshSvc2 := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend-second",
					Mesh: "default",
					Labels: map[string]string{
						"app":     "backend",
						"version": "second",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend-second"),
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.2",
					}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc, &meshSvc2},
			}

			dpBuilder := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")
			mc := meshContextWithResources(builders.Mesh(), dpBuilder.Build(), &meshSvc, &meshSvc2)

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoint("default_backend-second___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend-second", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend-second"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshContext(mc).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						dpBuilder,
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(meshSvc.Meta, api.PolicyDefault{
										Rules: []api.Rule{
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/version1",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.MeshService,
															Labels: &map[string]string{
																"app":     "backend",
																"version": "first",
															},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
												},
											},
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/version2",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.MeshService,
															Labels: &map[string]string{
																"app":     "backend",
																"version": "second",
															},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
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
		Entry("basic-real-meshservice-and-mzms-labels", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend",
					Mesh: "default",
					Labels: map[string]string{
						"app": "backend",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
					Identities: meshServiceSpiffeIdentities("spiffe://default/backend"),
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.1",
					}},
				},
			}
			meshMZSvc := meshmultizoneservice_api.MeshMultiZoneServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend",
					Mesh: "default",
					Labels: map[string]string{
						"app": "backend",
					},
				},
				Spec: &meshmultizoneservice_api.MeshMultiZoneService{
					Selector: meshmultizoneservice_api.Selector{
						MeshService: common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"app": "backend",
							},
						},
					},
					Ports: []meshmultizoneservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
				},
				Status: &meshmultizoneservice_api.MeshMultiZoneServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "11.0.0.2",
					}},
					MeshServices: []meshmultizoneservice_api.MatchedMeshService{
						{
							Name: "backend",
							Mesh: "default",
						},
					},
				},
			}

			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}
			resources.MeshLocalResources[meshmultizoneservice_api.MeshMultiZoneServiceType] = &meshmultizoneservice_api.MeshMultiZoneServiceResourceList{
				Items: []*meshmultizoneservice_api.MeshMultiZoneServiceResource{&meshMZSvc},
			}

			dpBuilder := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")
			mc := meshContextWithResources(builders.Mesh(), dpBuilder.Build(), &meshSvc, &meshMZSvc)

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoint("default_backend___mzsvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshContext(mc).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						dpBuilder,
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(meshSvc.Meta, api.PolicyDefault{
										Rules: []api.Rule{
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/ms",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.MeshService,
															Labels: &map[string]string{
																"app": "backend",
															},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
												},
											},
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/mzms",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.MeshMultiZoneService,
															Labels: &map[string]string{
																"app": "backend",
															},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
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
		Entry("match-priority", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefMeshService("backend", "", "80"),
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
													TargetRef: builders.TargetRefMeshService("backend", "", "80"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("mixed-tcp-and-http-outbounds", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
				},
			}
			tcpMeshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "other-tcp", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolTCP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.2"}},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc, &tcpMeshSvc},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("default_backend___msvc_80",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us")).
				AddEndpoint("default_other-tcp___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-tcp", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP), "app", "other-tcp"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
						{
							Address:  "10.0.0.2",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&tcpMeshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefMeshService("backend", "", "80"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("unresolvable-backend", func() outboundsTestCase {
			// alias-backend is intentionally NOT registered in mesh context, simulating a race
			// where the VIP is not yet allocated when the MeshHTTPRoute xDS snapshot is generated.
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
										Rules: []api.Rule{
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v2",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: builders.TargetRefMeshService("alias-backend", "", "80"),
														Weight:    pointer.To(uint(100)),
													}},
												},
											},
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: builders.TargetRefMeshService("backend", "", "80"),
														Weight:    pointer.To(uint(100)),
													}},
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
		Entry("request-header-modifiers", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
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
														Add: &[]api.HeaderKeyValue{{
															Name:  "request-add-header",
															Value: "add-value",
														}},
														Set: &[]api.HeaderKeyValue{{
															Name:  "request-set-header",
															Value: "set-value",
														}, {
															Name:  "request-set-header-multiple",
															Value: "one-value,second-value",
														}},
														Remove: &[]string{
															"request-header-to-remove",
														},
													},
												}},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("response-header-modifiers", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
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
														Add: &[]api.HeaderKeyValue{{
															Name:  "response-add-header",
															Value: "add-value",
														}},
														Set: &[]api.HeaderKeyValue{{
															Name:  "response-set-header",
															Value: "set-value",
														}},
														Remove: &[]string{
															"response-header-to-remove",
														},
													},
												}},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("request-redirect", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
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
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("url-rewrite", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
							ResourceRules: map[kri.Identifier]outbound.ResourceRule{
								backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
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
								}),
							},
						}),
					).
					Build(),
			}
		}()),
		Entry("headers-match", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))

			matches := []api.Match{{
				Headers: &[]common_api.HeaderMatch{{
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
			}}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: matches,
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("grpc-service", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolGRPC,
					}},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.1"}},
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolGRPC), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")).
					WithOutbounds(xds_types.Outbounds{
						{
							Address:  "10.0.0.1",
							Port:     80,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(unifiedNaming()).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(nil, api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefMeshService("backend", "", "80"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("request-mirror-real-resources", func() outboundsTestCase {
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend", Mesh: "default",
					Labels: map[string]string{
						"app": "backend",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentitySpiffeIDType,
							Value: "spiffe://default/backend",
						},
					},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.1",
					}},
				},
			}
			mirrorSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "payments-mirror", Mesh: "default",
					Labels: map[string]string{
						"app": "payments-mirror",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8086)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentitySpiffeIDType,
							Value: "spiffe://default/payments-mirror",
						},
					},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.2",
					}},
				},
			}
			mirrorMZSvc := meshmultizoneservice_api.MeshMultiZoneServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "payments-mz-mirror", Mesh: "default",
					Labels: map[string]string{
						"app": "payments-mz-mirror",
					},
				},
				Spec: &meshmultizoneservice_api.MeshMultiZoneService{
					Selector: meshmultizoneservice_api.Selector{
						MeshService: common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"app": "payments-mirror",
							},
						},
					},
					Ports: []meshmultizoneservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshmultizoneservice_api.MeshMultiZoneServiceStatus{
					VIPs: []meshservice_api.VIP{{
						IP: "10.0.0.3",
					}},
					MeshServices: []meshmultizoneservice_api.MatchedMeshService{
						{
							Name: "payments-mirror",
							Mesh: "default",
						},
					},
				},
			}

			mirrorMESvc := meshexternalservice_api.MeshExternalServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "payments-mes-mirror", Mesh: "default",
					Labels: map[string]string{
						"app": "payments-mes-mirror",
					},
				},
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolHTTP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{{
						Address: "payments.example.com",
						Port:    10000,
					}},
				},
				Status: &meshexternalservice_api.MeshExternalServiceStatus{
					VIP: meshexternalservice_api.VIP{
						IP: "10.20.20.1",
					},
				},
			}

			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc, &mirrorSvc},
			}
			resources.MeshLocalResources[meshmultizoneservice_api.MeshMultiZoneServiceType] = &meshmultizoneservice_api.MeshMultiZoneServiceResourceList{
				Items: []*meshmultizoneservice_api.MeshMultiZoneServiceResource{&mirrorMZSvc},
			}
			resources.MeshLocalResources[meshexternalservice_api.MeshExternalServiceType] = &meshexternalservice_api.MeshExternalServiceResourceList{
				Items: []*meshexternalservice_api.MeshExternalServiceResource{&mirrorMESvc},
			}

			dpBuilder := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")
			mc := meshContextWithResources(builders.Mesh(), dpBuilder.Build(), &meshSvc, &mirrorSvc, &mirrorMZSvc, &mirrorMESvc)

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("default_backend___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP))).
				AddEndpoint("default_payments-mirror___msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8086).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "payments-mirror", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP))).
				AddEndpoint("default_payments-mz-mirror___mzsvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.7").
					WithPort(8086).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "payments-mirror", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP))).
				AddEndpoint("default_payments-mes-mirror___extsvc_9090", xds_builders.Endpoint().
					WithTarget("payments.example.com").
					WithPort(10000).
					WithWeight(1).
					With(func(e *core_xds.Endpoint) {
						e.ExternalService = &core_xds.ExternalService{
							Protocol:      core_meta.ProtocolHTTP,
							OwnerResource: kri.From(&mirrorMESvc),
						}
					}))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithResources(resources).
					WithMeshContext(mc).
					WithEndpointMap(outboundTargets).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						dpBuilder,
					).
					WithOutbounds(xds_types.Outbounds{
						{
							Port:     builders.FirstOutboundPort,
							Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								ResourceRules: map[kri.Identifier]outbound.ResourceRule{
									backendMeshServiceIdentifier: test_policies.NewOutboundRule(meshSvc.Meta, api.PolicyDefault{
										Rules: []api.Rule{
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/ms",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestMirrorType,
														RequestMirror: &api.RequestMirror{
															Percentage: pointer.To(intstr.FromString("99.9")),
															BackendRef: common_api.BackendRef{
																TargetRef: common_api.TargetRef{
																	Kind:   common_api.MeshService,
																	Labels: &map[string]string{"app": "payments-mirror"},
																},
																Port: pointer.To(uint32(80)),
															},
														},
													}},
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind:   common_api.MeshService,
															Labels: &map[string]string{"app": "backend"},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
												},
											},
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/mzms",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestMirrorType,
														RequestMirror: &api.RequestMirror{
															BackendRef: common_api.BackendRef{
																TargetRef: common_api.TargetRef{
																	Kind:   common_api.MeshMultiZoneService,
																	Labels: &map[string]string{"app": "payments-mz-mirror"},
																},
																Port: pointer.To(uint32(80)),
															},
														},
													}},
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind:   common_api.MeshService,
															Labels: &map[string]string{"app": "backend"},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
												},
											},
											{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/mes",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestMirrorType,
														RequestMirror: &api.RequestMirror{
															BackendRef: common_api.BackendRef{
																TargetRef: common_api.TargetRef{
																	Kind:   common_api.MeshExternalService,
																	Labels: &map[string]string{"app": "payments-mes-mirror"},
																},
																Port: pointer.To(uint32(9090)),
															},
														},
													}},
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind:   common_api.MeshService,
															Labels: &map[string]string{"app": "backend"},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
												},
											},
											{
												// mirror to a MeshService that doesn't exist, the route
												// is generated without any mirror policy
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/missing",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestMirrorType,
														RequestMirror: &api.RequestMirror{
															BackendRef: common_api.BackendRef{
																TargetRef: common_api.TargetRef{
																	Kind:   common_api.MeshService,
																	Labels: &map[string]string{"app": "not-existing-mirror"},
																},
																Port: pointer.To(uint32(80)),
															},
														},
													}},
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind:   common_api.MeshService,
															Labels: &map[string]string{"app": "backend"},
														},
														Weight: pointer.To(uint(100)),
														Port:   pointer.To(uint32(80)),
													}},
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
	)
})

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

func meshContextWithResources(
	meshBuilder *builders.MeshBuilder,
	resources ...core_model.Resource,
) *xds_context.MeshContext {
	resourceStore := memory.NewStore()

	if meshBuilder == nil {
		meshBuilder = builders.Mesh()
	}

	mesh := meshBuilder.WithBuiltinMTLSBackend("ca-1").WithEnabledMTLSBackend("ca-1").Build()
	err := resourceStore.Create(context.Background(), mesh, store.CreateByKey("default", core_model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	for _, res := range resources {
		err = resourceStore.Create(
			context.Background(),
			res,
			store.CreateByKey(res.GetMeta().GetName(), res.GetMeta().GetMesh()),
			store.CreateWithLabels(res.GetMeta().GetLabels()),
		)
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
	)
	mc, err := meshContextBuilder.Build(context.Background(), "default")
	Expect(err).ToNot(HaveOccurred())

	return &mc
}

func dppForMeshExternalService(mes *meshexternalservice_api.MeshExternalServiceResource) (*builders.DataplaneBuilder, *core_xds.Proxy) {
	dp := builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithInboundOfTagsAndProtocol("http", mesh_proto.ServiceTag, "web")
	proxy := xds_builders.Proxy().
		WithDataplane(dp).
		WithOutbounds(xds_types.Outbounds{
			{
				Address:  "10.20.20.1",
				Port:     9090,
				Resource: kri.From(mes),
			},
		}).
		WithMetadata(&core_xds.DataplaneMetadata{
			SystemCaPath: "/tmp/ca-certs.crt",
		}).
		Build()

	return dp, proxy
}
