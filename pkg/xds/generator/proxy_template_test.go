package generator_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("ProxyTemplateGenerator", func() {
	Context("Error case", func() {
		type testCase struct {
			template *mesh_proto.ProxyTemplate
		}

		DescribeTable("should reject a template that still carries the legacy raw-Envoy fields",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateGenerator{
					ProxyTemplate: given.template,
				}
				proxy := &model.Proxy{
					Id: *model.BuildProxyId("", "demo.backend-01"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name:    "backend-01",
							Mesh:    "demo",
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
								TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
									RedirectPortOutbound: 15001,
									RedirectPortInbound:  15006,
								},
							},
						},
					},
					Metadata:       &model.DataplaneMetadata{},
					SecretsTracker: envoy_common.NewSecretsTracker("demo", []string{"demo"}),
					APIVersion:     envoy_common.APIV3,
				}
				xdsCtx := xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						Secrets: &xds.TestSecrets{},
					},
					Mesh: xds_context.MeshContext{
						Resource: &core_mesh.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "demo",
							},
							Spec: &mesh_proto.Mesh{},
						},
					},
				}

				// when
				rs, err := gen.Generate(context.Background(), xdsCtx, proxy)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("MeshProxyPatch"))
				Expect(rs).To(BeNil())
			},
			Entry("should fail when Conf.Modifications is non-empty", testCase{
				template: &mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{
							core_mesh.ProfileDefaultProxy,
						},
						Modifications: []*mesh_proto.ProxyTemplate_Modifications{
							{
								Type: &mesh_proto.ProxyTemplate_Modifications_Cluster_{
									Cluster: &mesh_proto.ProxyTemplate_Modifications_Cluster{
										Operation: mesh_proto.OpRemove,
										Match: &mesh_proto.ProxyTemplate_Modifications_Cluster_Match{
											Name: "to-be-removed",
										},
									},
								},
							},
						},
					},
				},
			}),
			Entry("should fail when Conf.Resources is non-empty", testCase{
				template: &mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{
							core_mesh.ProfileDefaultProxy,
						},
						Resources: []*mesh_proto.ProxyTemplateRawResource{{
							Name:     "raw-name",
							Version:  "raw-version",
							Resource: `{}`,
						}},
					},
				},
			}),
		)
	})
})
