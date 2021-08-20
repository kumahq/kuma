package v3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

var _ = Describe("CreateDownstreamTlsContext()", func() {

	Context("when mTLS is disabled on a given Mesh", func() {

		It("should return `nil`", func() {
			// given
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: core_mesh.NewMeshResource(),
				},
			}

			// when
			snippet, err := v3.CreateDownstreamTlsContext(ctx)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(snippet).To(BeNil())
		})
	})

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			expected string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					Mesh: xds_context.MeshContext{
						Resource: &core_mesh.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "default",
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

				// when
				snippet, err := v3.CreateDownstreamTlsContext(ctx)
				// then
				Expect(err).ToNot(HaveOccurred())
				// when
				actual, err := util_proto.ToYAML(snippet)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("metadata is `nil`", testCase{
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        ads: {}
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      ads: {}
                      resourceApiVersion: V3
                requireClientCertificate: true`,
			}),
		)
	})
})

var _ = Describe("CreateUpstreamTlsContext()", func() {

	Context("when mTLS is disabled on a given Mesh", func() {

		It("should return `nil`", func() {
			// given
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: core_mesh.NewMeshResource(),
				},
			}

			// when
			snippet, err := v3.CreateUpstreamTlsContext(ctx, "backend", "backend")
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(snippet).To(BeNil())
		})
	})

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			upstreamService string
			expected        string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					Mesh: xds_context.MeshContext{
						Resource: &core_mesh.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "default",
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

				// when
				snippet, err := v3.CreateUpstreamTlsContext(ctx, given.upstreamService, "")
				// then
				Expect(err).ToNot(HaveOccurred())
				// when
				actual, err := util_proto.ToYAML(snippet)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("metadata is `nil`", testCase{
				upstreamService: "backend",
				expected: `
                commonTlsContext:
                  alpnProtocols:
                  - kuma
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        ads: {}
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      ads: {}
                      resourceApiVersion: V3`,
			}),
		)
	})
})
