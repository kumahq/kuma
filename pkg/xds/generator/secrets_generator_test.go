package generator_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("SecretsGenerator", func() {

	type testCase struct {
		ctx      xds_context.Context
		proxy    *core_xds.Proxy
		expected string
	}

	DescribeTable("should not generate Envoy xDS resources unless mTLS is present",
		func(given testCase) {
			// setup
			gen := &generator.SecretsProxyGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(rs.List()).To(BeEmpty())
		},
		Entry("Mesh has no mTLS configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{},
				},
			},
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "demo.backend-01"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
				},
				APIVersion: envoy_common.APIV3,
			},
		}),
		Entry("Mesh has no mTLS configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{},
				},
			},
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", mesh_proto.ZoneEgressServiceName),
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					MeshResourcesList: []*core_xds.MeshResources{
						{
							Mesh: &core_mesh.MeshResource{
								Meta: &test_model.ResourceMeta{
									Name: "demo",
								},
							},
						},
					},
					ZoneEgressResource: &core_mesh.ZoneEgressResource{
						Meta: &test_model.ResourceMeta{
							Name: mesh_proto.ZoneEgressServiceName,
						},
					},
				},
				APIVersion: envoy_common.APIV3,
			},
		}),
	)

	DescribeTable("should generate Envoy xDS resources if secret backend is present",
		func(given testCase) {
			// given
			gen := &generator.SecretsProxyGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "secrets", given.expected)))
		},
		Entry("should create cluster for Zipkin", testCase{
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{
							Mtls: &mesh_proto.Mesh_Mtls{
								EnabledBackend: "ca-1",
								Backends: []*mesh_proto.CertificateAuthorityBackend{
									{
										Name: "ca-1",
										Type: "builtin",
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "default.backend-01"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "default",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				APIVersion: envoy_common.APIV3,
			},
			expected: "envoy-config-zipkin.golden.yaml",
		}),
		Entry("should create multiple secrets when multiple meshes present (dataplane)", testCase{
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "mesh-1",
						},
						Spec: &mesh_proto.Mesh{
							Mtls: &mesh_proto.Mesh_Mtls{
								EnabledBackend: "ca-1",
								Backends: []*mesh_proto.CertificateAuthorityBackend{
									{
										Name: "ca-1",
										Type: "builtin",
									},
								},
							},
						},
					},
					Resources: xds_context.Resources{
						MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
							core_mesh.MeshType: &core_mesh.MeshResourceList{
								Items: []*core_mesh.MeshResource{{
									Spec: &mesh_proto.Mesh{
										Mtls: &mesh_proto.Mesh_Mtls{
											EnabledBackend: "ca-1",
											Backends: []*mesh_proto.CertificateAuthorityBackend{
												{
													Name: "ca-1",
													Type: "builtin",
												},
											},
										},
									},
									Meta: &test_model.ResourceMeta{
										Name: "mesh-2",
									},
								}},
							}},
						CrossMeshResources: map[string]map[core_model.ResourceType]core_model.ResourceList{
							"mesh-2": {
								core_mesh.MeshGatewayType: &core_mesh.MeshGatewayResourceList{
									Items: []*core_mesh.MeshGatewayResource{{
										Meta: &test_model.ResourceMeta{
											Name: "mesh2",
										},
										Spec: &mesh_proto.MeshGateway{
											Conf: &mesh_proto.MeshGateway_Conf{
												Listeners: []*mesh_proto.MeshGateway_Listener{{
													Hostname: "gateway1.mesh",
													Port:     80,
													Protocol: mesh_proto.MeshGateway_Listener_HTTP,
													Tags: map[string]string{
														"listener": "internal",
													},
												}, {
													Hostname: "*",
													Port:     80,
													Protocol: mesh_proto.MeshGateway_Listener_HTTP,
													Tags: map[string]string{
														"listener": "wildcard",
													},
												}},
											},
											Selectors: []*mesh_proto.Selector{{
												Match: map[string]string{
													mesh_proto.ServiceTag: "gateway",
												},
											}},
											Tags: map[string]string{
												"gateway": "prod",
											},
										},
									}},
								},
							}},
					}},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", mesh_proto.ZoneEgressServiceName),
				APIVersion: envoy_common.APIV3,
			},
			expected: "envoy-config-dataplane.golden.yaml",
		}),
		Entry("should create secrets when multiple meshes present (egress)", testCase{
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", mesh_proto.ZoneEgressServiceName),
				APIVersion: envoy_common.APIV3,
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					MeshResourcesList: []*core_xds.MeshResources{
						{
							Mesh: &core_mesh.MeshResource{
								Meta: &test_model.ResourceMeta{
									Name: "mesh-1",
								},
								Spec: &mesh_proto.Mesh{
									Mtls: &mesh_proto.Mesh_Mtls{
										EnabledBackend: "ca-1",
										Backends: []*mesh_proto.CertificateAuthorityBackend{
											{
												Name: "ca-1",
												Type: "builtin",
											},
										},
									},
								},
							},
						},
						{
							Mesh: &core_mesh.MeshResource{
								Meta: &test_model.ResourceMeta{
									Name: "mesh-2",
								},
								Spec: &mesh_proto.Mesh{
									Mtls: &mesh_proto.Mesh_Mtls{
										EnabledBackend: "ca-1",
										Backends: []*mesh_proto.CertificateAuthorityBackend{
											{
												Name: "ca-1",
												Type: "builtin",
											},
										},
									},
								},
							},
							// only meshes with external services are taken into account
							ExternalServices: []*core_mesh.ExternalServiceResource{
								{
									Meta: &test_model.ResourceMeta{
										Name: "es-mesh-2",
										Mesh: "mesh-2",
									},
									Spec: &mesh_proto.ExternalService{
										Networking: &mesh_proto.ExternalService_Networking{
											Address: "example.com:80",
										},
										Tags: map[string]string{
											"kuma.io/service": "service2",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "envoy-config-egress.golden.yaml",
		}),
	)
})
