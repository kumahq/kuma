package generator_test

import (
	"context"
	"path/filepath"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	bldrs_common "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v2/pkg/envoy/builders/tls"
	. "github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	"github.com/kumahq/kuma/v2/pkg/xds/generator"
)

var _ = Describe("ZoneProxyListenerGenerator", func() {
	const cpZone = "east"

	testWorkloadIdentity := &core_xds.WorkloadIdentity{
		IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return bldrs_tls.SdsSecretConfigSource(
				"my-identity-secret",
				bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
			)
		},
	}

	type testCase struct {
		proxy    *core_xds.Proxy
		expected string
	}

	It("returns nil when DataplaneZoneListeners is nil", func() {
		gen := generator.ZoneProxyListenerGenerator{}
		proxy := &core_xds.Proxy{
			Id:         *core_xds.BuildProxyId("default", "dp-1"),
			APIVersion: envoy_common.APIV3,
			Dataplane:  samples.DataplaneBackend(),
			Metadata: &core_xds.DataplaneMetadata{
				Features: map[string]bool{
					xds_types.FeatureUnifiedResourceNaming: true,
				},
			},
		}

		rs, err := gen.Generate(context.Background(), nil, xds_context.Context{
			ControlPlane: &xds_context.ControlPlaneContext{Zone: cpZone},
		}, proxy)

		Expect(err).ToNot(HaveOccurred())
		Expect(rs).To(BeNil())
	})

	DescribeTable("should generate Envoy xDS resources",
		func(given testCase) {
			gen := generator.ZoneProxyListenerGenerator{}

			rs, err := gen.Generate(context.Background(), nil, xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Zone:            cpZone,
					SystemNamespace: "kuma-system",
				},
			}, given.proxy)
			Expect(err).ToNot(HaveOccurred())

			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "zone-proxy-listener", given.expected)))
		},
		Entry("ingress: local MeshService with zone label", testCase{
			proxy: func() *core_xds.Proxy {
				ms := samples.MeshServiceBackendBuilder().
					WithLabels(map[string]string{mesh_proto.ZoneTag: cpZone}).
					Build()
				// LegacyServiceName for default/backend/zone=east/port=80: default_backend__east_msvc_80
				const legacySvcName = "default_backend__east_msvc_80"
				return &core_xds.Proxy{
					Id:         *core_xds.BuildProxyId("default", "dp-1"),
					APIVersion: envoy_common.APIV3,
					Dataplane:  samples.DataplaneBackend(),
					Metadata: &core_xds.DataplaneMetadata{
						Features: map[string]bool{
							xds_types.FeatureUnifiedResourceNaming: true,
						},
					},
					DataplaneZoneListeners: &core_xds.DataplaneZoneListeners{
						IngressListeners: []*core_xds.DataplaneListener{
							{
								Listener: &mesh_proto.Dataplane_Networking_Listener{
									Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
									Address: "10.0.0.1",
									Port:    10001,
									Name:    "zone-ingress-port",
								},
								MeshResources: &core_xds.MeshProxyResources{
									Mesh: builders.Mesh().WithName("default").WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).Build(),
									EndpointMap: core_xds.EndpointMap{
										legacySvcName: []core_xds.Endpoint{
											{
												Target: "192.168.0.1",
												Port:   2521,
												Tags:   map[string]string{"kuma.io/service": "backend"},
												Weight: 1,
											},
										},
									},
									Resources: map[core_model.ResourceType]core_model.ResourceList{
										meshservice_api.MeshServiceType: &meshservice_api.MeshServiceResourceList{
											Items: []*meshservice_api.MeshServiceResource{ms},
										},
									},
								},
							},
						},
					},
					InternalAddresses: DummyInternalAddresses,
				}
			}(),
			expected: "ingress-meshservice-zone.envoy.golden.yaml",
		}),
		Entry("ingress: non-local MeshService filtered out", testCase{
			proxy: func() *core_xds.Proxy {
				msLocal := samples.MeshServiceBackendBuilder().
					WithLabels(map[string]string{mesh_proto.ZoneTag: cpZone}).
					Build()
				msRemote := samples.MeshServiceBackendBuilder().
					WithName("backend-remote").
					WithLabels(map[string]string{mesh_proto.ZoneTag: "west"}).
					Build()
				svcName := kri.From(msLocal).String()
				return &core_xds.Proxy{
					Id:         *core_xds.BuildProxyId("default", "dp-1"),
					APIVersion: envoy_common.APIV3,
					Dataplane:  samples.DataplaneBackend(),
					Metadata: &core_xds.DataplaneMetadata{
						Features: map[string]bool{
							xds_types.FeatureUnifiedResourceNaming: true,
						},
					},
					DataplaneZoneListeners: &core_xds.DataplaneZoneListeners{
						IngressListeners: []*core_xds.DataplaneListener{
							{
								Listener: &mesh_proto.Dataplane_Networking_Listener{
									Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
									Address: "10.0.0.1",
									Port:    10001,
									Name:    "zone-ingress-port",
								},
								MeshResources: &core_xds.MeshProxyResources{
									Mesh: builders.Mesh().WithName("default").WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).Build(),
									EndpointMap: core_xds.EndpointMap{
										svcName: []core_xds.Endpoint{
											{
												Target: "192.168.0.1",
												Port:   2521,
												Tags:   map[string]string{"kuma.io/service": "backend"},
												Weight: 1,
											},
										},
									},
									Resources: map[core_model.ResourceType]core_model.ResourceList{
										meshservice_api.MeshServiceType: &meshservice_api.MeshServiceResourceList{
											Items: []*meshservice_api.MeshServiceResource{msLocal, msRemote},
										},
									},
								},
							},
						},
					},
					InternalAddresses: DummyInternalAddresses,
				}
			}(),
			expected: "ingress-meshservice-filtered.envoy.golden.yaml",
		}),
		Entry("egress: no WorkloadIdentity → no resources generated", testCase{
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("default", "dp-1"),
				APIVersion: envoy_common.APIV3,
				Dataplane:  samples.DataplaneBackend(),
				Metadata: &core_xds.DataplaneMetadata{
					Features: map[string]bool{
						xds_types.FeatureUnifiedResourceNaming: true,
					},
				},
				DataplaneZoneListeners: &core_xds.DataplaneZoneListeners{
					EgressListeners: []*core_xds.DataplaneListener{
						{
							Listener: &mesh_proto.Dataplane_Networking_Listener{
								Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
								Address: "10.0.0.1",
								Port:    10002,
								Name:    "zone-egress-port",
							},
							MeshResources: &core_xds.MeshProxyResources{
								Mesh:        builders.Mesh().WithName("default").Build(),
								EndpointMap: core_xds.EndpointMap{},
								Resources:   map[core_model.ResourceType]core_model.ResourceList{},
							},
						},
					},
				},
				InternalAddresses: DummyInternalAddresses,
			},
			expected: "egress-no-identity.envoy.golden.yaml",
		}),
		Entry("egress: MeshExternalService with WorkloadIdentity", testCase{
			proxy: func() *core_xds.Proxy {
				mes := builders.MeshExternalService().WithKumaVIP("242.0.0.1").Build()
				mesKRI := kri.From(mes)
				// UnifiedName (unified naming enabled + Exclusive mesh): kri_extsvc_default___example_9000
				unifiedSvcName := kri.WithSectionName(mesKRI, mes.Spec.Match.GetName()).String()
				return &core_xds.Proxy{
					Id:         *core_xds.BuildProxyId("default", "dp-1"),
					APIVersion: envoy_common.APIV3,
					Dataplane:  samples.DataplaneBackend(),
					Metadata: &core_xds.DataplaneMetadata{
						SystemCaPath: "/etc/ssl/certs/ca-certificates.crt",
						Features: map[string]bool{
							xds_types.FeatureUnifiedResourceNaming: true,
						},
					},
					WorkloadIdentity: testWorkloadIdentity,
					DataplaneZoneListeners: &core_xds.DataplaneZoneListeners{
						EgressListeners: []*core_xds.DataplaneListener{
							{
								Listener: &mesh_proto.Dataplane_Networking_Listener{
									Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
									Address: "10.0.0.1",
									Port:    10002,
									Name:    "zone-egress-port",
								},
								MeshResources: &core_xds.MeshProxyResources{
									Mesh: builders.Mesh().WithName("default").WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).Build(),
									EndpointMap: core_xds.EndpointMap{
										unifiedSvcName: []core_xds.Endpoint{
											{
												Target: "192.168.0.1",
												Port:   27017,
												Tags:   map[string]string{},
												Weight: 1,
												ExternalService: &core_xds.ExternalService{
													Protocol:      core_meta.ProtocolHTTP,
													OwnerResource: mesKRI,
												},
											},
										},
									},
									Resources: map[core_model.ResourceType]core_model.ResourceList{
										meshexternalservice_api.MeshExternalServiceType: &meshexternalservice_api.MeshExternalServiceResourceList{
											Items: []*meshexternalservice_api.MeshExternalServiceResource{mes},
										},
									},
								},
							},
						},
					},
					InternalAddresses: DummyInternalAddresses,
				}
			}(),
			expected: "egress-meshexternalservice.envoy.golden.yaml",
		}),
	)
})
