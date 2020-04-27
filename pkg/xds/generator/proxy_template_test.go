package generator_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"
)

var _ = Describe("TemplateProxyGenerator", func() {
	Context("Error case", func() {
		type testCase struct {
			proxy    *model.Proxy
			template *mesh_proto.ProxyTemplate
			err      interface{}
		}

		DescribeTable("Avoid producing invalid Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.TemplateProxyGenerator{
					ProxyTemplate: given.template,
				}
				ctx := xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-system:5677",
						SdsTlsCert:  []byte("12345"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "demo",
							},
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
					Id: model.ProxyId{Name: "demo.backend-01"},
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name:    "backend-01",
							Mesh:    "demo",
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
								TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
									RedirectPort: 15001,
								},
							},
						},
					},
				},
				template: &mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{
							mesh_core.ProfileDefaultProxy,
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
			envoyConfigFile   string
		}

		DescribeTable("Generate Envoy xDS resources",
			func(given testCase) {
				// setup
				proxyTemplate := mesh_proto.ProxyTemplate{}
				ptBytes, err := ioutil.ReadFile(filepath.Join("testdata", "template-proxy", given.proxyTemplateFile))
				Expect(err).ToNot(HaveOccurred())
				Expect(util_proto.FromYAML(ptBytes, &proxyTemplate)).To(Succeed())
				gen := &generator.TemplateProxyGenerator{
					ProxyTemplate: &proxyTemplate,
				}

				// given
				ctx := xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-system:5677",
						SdsTlsCert:  []byte("12345"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "demo",
							},
							Spec: mesh_proto.Mesh{
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

				dataplane := mesh_proto.Dataplane{}
				Expect(util_proto.FromYAML([]byte(given.dataplane), &dataplane)).To(Succeed())
				proxy := &model.Proxy{
					Id: model.ProxyId{Name: "demo.backend-01"},
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name:    "backend-01",
							Mesh:    "demo",
							Version: "1",
						},
						Spec: dataplane,
					},
					Metadata: &model.DataplaneMetadata{},
				}

				// when
				rs, err := gen.Generate(ctx, proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				resp, err := model.ResourceList(rs).ToDeltaDiscoveryResponse()
				// then
				Expect(err).ToNot(HaveOccurred())
				// when
				actual, err := util_proto.ToYAML(resp)
				// then
				Expect(err).ToNot(HaveOccurred())

				expected, err := ioutil.ReadFile(filepath.Join("testdata", "template-proxy", given.envoyConfigFile))
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(expected))
			},
			Entry("should support a combination of pre-defined profiles and raw xDS resources", testCase{
				dataplane: `
                networking:
                  transparentProxying:
                    redirectPort: 15001
                  address: 192.168.0.1
                  inbound:
                    - port: 80
                      servicePort: 8080
`,
				proxyTemplateFile: "1-proxy-template.input.yaml",
				envoyConfigFile:   "1-envoy-config.golden.yaml",
			}),
		)

	})
})
