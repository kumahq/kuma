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
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
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
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			externalServices := xds_builders.EndpointMap().
				AddEndpoint("external-service", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8085).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "external-service", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithExternalServicesEndpointMap(externalServices).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
					AddServiceProtocol("external-service", core_meta.ProtocolHTTP).
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
					WithInternalAddresses(core_xds.InternalAddress{AddressPrefix: "192.168.0.0", PrefixLen: 16}, core_xds.InternalAddress{AddressPrefix: "::1", PrefixLen: 128}).
					Build(),
			}
		}()),
		Entry("default-route-outbound-with-tags-with-mtls", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(samples.MeshMTLSBuilder()).
					WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
					AddServiceProtocol("default_backend___svc_80", core_meta.ProtocolHTTP).
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
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
							Address:  "10.0.0.1",
							Port:     80,
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					Build(),
			}
		}()),
		Entry("default-meshservice-unified-naming", func() outboundsTestCase {
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
				Meta: &test_model.ResourceMeta{
					Name: "backend",
					Mesh: "default",
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{
						{
							Port:        80,
							TargetPort:  pointer.To(intstr.FromInt32(8084)),
							AppProtocol: core_meta.ProtocolHTTP,
						},
					},
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentityServiceTagType,
							Value: "backend",
						},
					},
				},
				Status: &meshservice_api.MeshServiceStatus{
					VIPs: []meshservice_api.VIP{
						{
							IP: "10.0.0.1",
						},
					},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}

			mesh := builders.Mesh().
				WithBuiltinMTLSBackend("builtin").
				WithEnabledMTLSBackend("builtin")

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(mesh).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					AddServiceProtocol("default_backend___svc_80", core_meta.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithSecretsTracker(envoy.NewSecretsTracker(core_model.DefaultMesh, nil)).
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{
						{
							LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
								Port: builders.FirstOutboundPort,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "backend",
								},
							},
						},
						{
							Resource: kri.WithSectionName(kri.From(&meshSvc), "80"),
							Address:  "10.0.0.1",
							Port:     80,
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(&core_xds.DataplaneMetadata{
						Features: map[string]bool{
							xds_types.FeatureUnifiedResourceNaming: true,
						},
					}).
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
						mesh_proto.ZoneTag: "remote-zone",
					},
				},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{{
						Type:  meshservice_api.MeshServiceIdentityServiceTagType,
						Value: "backend",
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
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithMeshBuilder(builders.Mesh().WithBuiltinMTLSBackend("builtin").WithEnabledMTLSBackend("builtin")).
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					With(func(ctx *xds_context.Context) {
						ctx.Mesh.ZonesWithMeshScopedProxy = map[string]bool{"remote-zone": true}
					}).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
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
						ctx.Mesh.ZonesWithMeshScopedProxy = map[string]bool{"remote-zone": true}
						ctx.Mesh.ZoneEgresses = []core_xds.ZoneEgressInstance{
							{Address: "10.0.0.1", Port: 10002, SAN: "spiffe://default/zone-egress"},
						}
					}).
					Build(),
				proxy: xds_builders.Proxy().
					WithSecretsTracker(envoy.NewSecretsTracker(core_model.DefaultMesh, nil)).
					WithDataplane(builders.Dataplane().
						WithName("web-01").
						WithAddress("192.168.0.2").
						WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
					).
					WithOutbounds(xds_types.Outbounds{{
						Resource: kri.WithSectionName(extSvcKRI, "9000"),
						Address:  "10.20.20.1",
						Port:     9000,
					}}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithMetadata(&core_xds.DataplaneMetadata{
						Features: map[string]bool{
							xds_types.FeatureUnifiedResourceNaming: true,
						},
					}).
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
						DataplaneTags: &map[string]string{
							mesh_proto.ServiceTag: "backend",
						},
					},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
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
					With(func(ctx *xds_context.Context) {
						ctx.Mesh.ZonesWithMeshScopedProxy = map[string]bool{"local": true}
					}).
					Build(),
				proxy: proxy,
			}
		}()),
		Entry("default-meshmultizoneservice-legacy-zone", func() outboundsTestCase {
			// Regression test: MZMS with a remote endpoint in "remote-zone" which has only a legacy
			// ZoneIngress (absent from ZonesWithMeshScopedProxy). With WorkloadIdentity set, the
			// cluster SNI must use the old hash-based format, not the new KRI format.
			remoteMeshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend-remote", Mesh: "default",
					Labels: map[string]string{
						"service":                      "backend",
						mesh_proto.ZoneTag:             "remote-zone",
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentityServiceTagType,
							Value: "backend",
						},
					},
					State: meshservice_api.StateAvailable,
				},
				Status: &meshservice_api.MeshServiceStatus{},
			}
			meshMZSvc := meshmultizoneservice_api.MeshMultiZoneServiceResource{
				Meta: &test_model.ResourceMeta{Name: "multi-backend", Mesh: "default"},
				Spec: &meshmultizoneservice_api.MeshMultiZoneService{
					Selector: meshmultizoneservice_api.Selector{
						MeshService: common_api.LabelSelector{
							MatchLabels: &map[string]string{"service": "backend"},
						},
					},
					Ports: []meshmultizoneservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshmultizoneservice_api.MeshMultiZoneServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.2"}},
					MeshServices: []meshmultizoneservice_api.MatchedMeshService{
						{Name: "backend-remote", Zone: "remote-zone", Mesh: "default"},
					},
				},
			}
			zoneIngress := builders.ZoneIngress().
				WithZone("remote-zone").
				WithAddress("10.10.10.1").
				WithAdvertisedAddress("10.10.10.10").
				WithAdvertisedPort(15050).
				Build()

			dp := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
				Build()
			mc := meshContextWithResources(builders.Mesh(), dp, &remoteMeshSvc, &meshMZSvc, zoneIngress)

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
					With(func(ctx *xds_context.Context) {
						// "remote-zone" absent: it uses a legacy ZoneIngress, not a new zone proxy
						ctx.Mesh.ZonesWithMeshScopedProxy = map[string]bool{}
					}).
					Build(),
				proxy: proxy,
			}
		}()),
		Entry("default-meshmultizoneservice-mixed-zones", func() outboundsTestCase {
			// Regression test for the reviewer's concern: an MZMS spanning a zone with a new
			// mesh-scoped zone proxy ("new-zone") and a zone with only a legacy ZoneIngress
			// ("legacy-zone"). A single cluster-wide SNI can't satisfy both, so the cluster
			// keeps the KRI SNI as the default transport socket (new-zone) and adds a per-zone
			// transport_socket_match with the hash-based SNI for "legacy-zone".
			newMeshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend-new", Mesh: "default",
					Labels: map[string]string{
						"service":                      "backend",
						mesh_proto.ZoneTag:             "new-zone",
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{Type: meshservice_api.MeshServiceIdentityServiceTagType, Value: "backend"},
					},
					State: meshservice_api.StateAvailable,
				},
				Status: &meshservice_api.MeshServiceStatus{},
			}
			legacyMeshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend-legacy", Mesh: "default",
					Labels: map[string]string{
						"service":                      "backend",
						mesh_proto.ZoneTag:             "legacy-zone",
						mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
					},
				},
				Spec: &meshservice_api.MeshService{
					Ports: []meshservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{Type: meshservice_api.MeshServiceIdentityServiceTagType, Value: "backend"},
					},
					State: meshservice_api.StateAvailable,
				},
				Status: &meshservice_api.MeshServiceStatus{},
			}
			meshMZSvc := meshmultizoneservice_api.MeshMultiZoneServiceResource{
				Meta: &test_model.ResourceMeta{Name: "multi-backend", Mesh: "default"},
				Spec: &meshmultizoneservice_api.MeshMultiZoneService{
					Selector: meshmultizoneservice_api.Selector{
						MeshService: common_api.LabelSelector{
							MatchLabels: &map[string]string{"service": "backend"},
						},
					},
					Ports: []meshmultizoneservice_api.Port{{
						Port:        80,
						AppProtocol: core_meta.ProtocolHTTP,
					}},
				},
				Status: &meshmultizoneservice_api.MeshMultiZoneServiceStatus{
					VIPs: []meshservice_api.VIP{{IP: "10.0.0.2"}},
					MeshServices: []meshmultizoneservice_api.MatchedMeshService{
						{Name: "backend-new", Zone: "new-zone", Mesh: "default"},
						{Name: "backend-legacy", Zone: "legacy-zone", Mesh: "default"},
					},
				},
			}
			meshZoneAddress := meshzoneaddress_api.MeshZoneAddressResource{
				Meta: &test_model.ResourceMeta{
					Name: "mza-new-zone", Mesh: "default",
					Labels: map[string]string{mesh_proto.ZoneTag: "new-zone"},
				},
				Spec: &meshzoneaddress_api.MeshZoneAddress{
					Address: "10.20.20.20",
					Port:    15050,
				},
			}
			zoneIngress := builders.ZoneIngress().
				WithZone("legacy-zone").
				WithAddress("10.10.10.1").
				WithAdvertisedAddress("10.10.10.10").
				WithAdvertisedPort(15050).
				Build()

			dp := builders.Dataplane().
				WithName("web-01").
				WithAddress("192.168.0.2").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
				Build()
			mc := meshContextWithResources(builders.Mesh(), dp, &newMeshSvc, &legacyMeshSvc, &meshMZSvc, &meshZoneAddress, zoneIngress)

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
					With(func(ctx *xds_context.Context) {
						// "new-zone" has a mesh-scoped zone proxy; "legacy-zone" only a ZoneIngress
						ctx.Mesh.ZonesWithMeshScopedProxy = map[string]bool{"new-zone": true}
					}).
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
						DataplaneTags: &map[string]string{
							mesh_proto.ServiceTag: "backend",
						},
					},
					Ports: []meshservice_api.Port{{
						Port:        80,
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
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
			egress := builders.ZoneEgress().WithPort(10002).Build()
			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, egress)

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

			egress := builders.ZoneEgress().WithPort(10002).Build()
			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, egress)

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
			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, egress)

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
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
						Verification: &meshexternalservice_api.Verification{
							Mode: meshexternalservice_api.TLSVerificationSkipAll,
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
			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, egress)

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
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     9090,
						Protocol: core_meta.ProtocolHTTP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{
						{
							Address: "example.com",
							Port:    10000,
						},
						{
							Address: "example2.com",
							Port:    11111,
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
			mc := meshContextWithResources(builders.Mesh(), dp.Build(), &meshExtSvc, egress)

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithMeshContext(mc).Build(),
				proxy:      proxy,
			}
		}()),
		Entry("meshexternalservice-with-tls-and-custom-settings-unified-naming", func() outboundsTestCase {
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
						{
							Address: "example2.com",
							Port:    11111,
						},
					},
					Tls: &meshexternalservice_api.Tls{
						Enabled: true,
						Verification: &meshexternalservice_api.Verification{
							ServerName: pointer.To("example2.com"),
							SubjectAltNames: &[]meshexternalservice_api.SANMatch{
								{
									Type:  meshexternalservice_api.SANMatchPrefix,
									Value: "example",
								},
								{
									Type:  meshexternalservice_api.SANMatchExact,
									Value: "example2.com",
								},
							},
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

			dp, proxy := dppForMeshExternalService(&meshExtSvc, xds_types.FeatureUnifiedResourceNaming)
			egress := builders.ZoneEgress().WithPort(10002).Build()

			mc := meshContextWithResources(
				builders.Mesh(),
				dp.Build(),
				&meshExtSvc,
				egress,
			)

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
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(nil, api.PolicyDefault{
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
												QueryParams: &[]api.QueryParamsMatch{{
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
									}),
								},
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
						TargetPort:  pointer.To(intstr.FromInt(8084)),
						AppProtocol: core_meta.ProtocolHTTP,
						Name:        pointer.To("test-port"),
					}},
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoints("backend",
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
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
							Resource: kri.WithSectionName(kri.From(&meshSvc), "test-port"),
						},
					}).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
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
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
					Identities: &[]meshservice_api.MeshServiceIdentity{
						{
							Type:  meshservice_api.MeshServiceIdentityServiceTagType,
							Value: "backend-second",
						},
					},
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
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")
			mc := meshContextWithResources(builders.Mesh(), dpBuilder.Build(), &meshSvc, &meshSvc2)

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoint("backend-second_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend-second", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend-second")).
				AddEndpoints("backend",
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
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
					AddServiceProtocol("backend-second", core_meta.ProtocolHTTP).
					WithResources(resources).
					WithMeshContext(mc).
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
					WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
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
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")
			mc := meshContextWithResources(builders.Mesh(), dpBuilder.Build(), &meshSvc, &meshMZSvc)

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoint("backend_mzsvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					WithMeshContext(mc).
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
					WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
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
		Entry("basic-real-meshservice-and-mzms-labels-unified-naming", func() outboundsTestCase {
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
					Identities: &[]meshservice_api.MeshServiceIdentity{
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
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")

			mc := meshContextWithResources(
				builders.Mesh(),
				dpBuilder.Build(),
				&meshSvc,
				&meshMZSvc,
			)

			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend_msvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend")).
				AddEndpoint("backend_mzsvc_80", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "app", "backend"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					WithResources(resources).
					WithMeshContext(mc).
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
					WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
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
					WithMetadata(&core_xds.DataplaneMetadata{Features: map[string]bool{xds_types.FeatureUnifiedResourceNaming: true}}).
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
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(nil, api.PolicyDefault{
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
									}),
								},
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
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us")).
				AddEndpoint("other-tcp", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-tcp", mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP), "region", "eu"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
					AddServiceProtocol("other-tcp", core_meta.ProtocolTCP).
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
									test_policies.NewRule(nil, api.PolicyDefault{
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
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
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(nil, api.PolicyDefault{
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
														TargetRef: builders.TargetRefServiceSubset("alias-backend", mesh_proto.ZoneTag, "zone-2"),
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
														TargetRef: builders.TargetRefService("backend"),
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
								test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
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
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolGRPC), "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolGRPC).
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
									test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
									}),
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us")).
				AddEndpoint("payments", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8086).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "payments", mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP), "region", "us", "version", "v1", "env", "dev"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_meta.ProtocolHTTP).
					AddServiceProtocol("payments", core_meta.ProtocolHTTP).
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
									test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
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
																	Name: pointer.To("payments"),
																	Tags: &map[string]string{
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
																	Name: pointer.To("backend"),
																},
															},
														},
													},
												},
											},
										}},
									}),
								},
							}),
					).
					Build(),
			}
		}()),
	)
})

func meshContextWithResources(
	meshBuilder *builders.MeshBuilder,
	resources ...core_model.Resource,
) *xds_context.MeshContext {
	resourceStore := memory.NewStore()

	if meshBuilder == nil {
		meshBuilder = builders.Mesh()
	}

	mesh := meshBuilder.WithBuiltinMTLSBackend("ca-1").WithEgressRoutingEnabled().WithEnabledMTLSBackend("ca-1").Build()
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
		nil,
	)
	mc, err := meshContextBuilder.Build(context.Background(), "default")
	Expect(err).ToNot(HaveOccurred())

	return &mc
}

func dppForMeshExternalService(mes *meshexternalservice_api.MeshExternalServiceResource, feature ...string) (*builders.DataplaneBuilder, *core_xds.Proxy) {
	features := xds_types.Features{}
	for _, f := range feature {
		features[f] = true
	}

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
				Resource: kri.From(mes),
			},
		}).
		WithSecretsTracker(envoy.NewSecretsTracker("default", nil)).
		WithMetadata(&core_xds.DataplaneMetadata{
			SystemCaPath: "/tmp/ca-certs.crt",
			Features:     features,
		}).
		Build()

	return dp, proxy
}
