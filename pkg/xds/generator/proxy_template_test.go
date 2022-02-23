package generator_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("ProxyTemplateGenerator", func() {
	Context("Error case", func() {
		type testCase struct {
			proxy    *model.Proxy
			template *mesh_proto.ProxyTemplate
			err      interface{}
		}

		DescribeTable("Avoid producing invalid Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateGenerator{
					ProxyTemplate: given.template,
				}
				ctx := xds_context.Context{
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
				rs, err := gen.Generate(ctx, given.proxy)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(given.err))
				Expect(rs).To(BeNil())
			},
			Entry("should fail when raw xDS resource is not valid", testCase{
				proxy: &model.Proxy{
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
					APIVersion: envoy_common.APIV3,
				},
				template: &mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{
							core_mesh.ProfileDefaultProxy,
						},
						Resources: []*mesh_proto.ProxyTemplateRawResource{{
							Name:     "raw-name",
							Version:  "raw-version",
							Resource: `{`,
						}},
					},
				},
				err: "resources: raw.resources[0]{name=\"raw-name\"}.resource: unexpected EOF",
			}),
		)
	})

	Context("Happy case", func() {

		type testCase struct {
			dataplane         string
			proxyTemplateFile string
			expected          string
		}

		DescribeTable("Generate Envoy xDS resources",
			func(given testCase) {
				// setup
				proxyTemplate := mesh_proto.ProxyTemplate{}
				ptBytes, err := os.ReadFile(filepath.Join("testdata", "template-proxy", given.proxyTemplateFile))
				Expect(err).ToNot(HaveOccurred())
				Expect(util_proto.FromYAML(ptBytes, &proxyTemplate)).To(Succeed())
				gen := &generator.ProxyTemplateGenerator{
					ProxyTemplate: &proxyTemplate,
				}

				// given
				ctx := xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						Secrets: &xds.TestSecrets{},
					},
					Mesh: xds_context.MeshContext{
						Resource: &core_mesh.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "demo",
							},
							Spec: &mesh_proto.Mesh{
								Mtls: &mesh_proto.Mesh_Mtls{
									EnabledBackend: "builtin",
									Backends: []*mesh_proto.CertificateAuthorityBackend{
										{
											Name: "builtin",
											Type: "builtin",
										},
									},
								},
							},
						},
					},
				}

				dataplane := &mesh_proto.Dataplane{}
				Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane)).To(Succeed())
				proxy := &model.Proxy{
					Id: *model.BuildProxyId("", "demo.backend-01"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name:    "backend-01",
							Mesh:    "demo",
							Version: "1",
						},
						Spec: dataplane,
					},
					APIVersion: envoy_common.APIV3,
					Metadata:   &model.DataplaneMetadata{},
				}

				// when
				rs, err := gen.Generate(ctx, proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				resp, err := rs.List().ToDeltaDiscoveryResponse()
				// then
				Expect(err).ToNot(HaveOccurred())
				// when
				actual, err := util_proto.ToYAML(resp)
				// then
				Expect(err).ToNot(HaveOccurred())

				// and output matches golden files
				Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "template-proxy", given.expected)))
			},
			Entry("should support a combination of pre-defined profiles and raw xDS resources", testCase{
				dataplane: `
                networking:
                  transparentProxying:
                    redirectPortOutbound: 15001
                    redirectPortInbound: 15006
                  address: 192.168.0.1
                  inbound:
                    - port: 80
                      servicePort: 8080
`,
				proxyTemplateFile: "1-proxy-template.input.yaml",
				expected:          "1-envoy-config.golden.yaml",
			}),
			Entry("should support merging non entire seconds durations", testCase{
				dataplane: `
                networking:
                  transparentProxying:
                    redirectPortOutbound: 15001
                    redirectPortInbound: 15006
                  address: 192.168.0.1
                  inbound:
                    - port: 80
                      servicePort: 8080
`,
				proxyTemplateFile: "2-proxy-template.input.yaml",
				expected:          "2-envoy-config.golden.yaml",
			}),
		)

	})
})
