package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type caRequest struct {
	mesh string
}

func (r *caRequest) MeshName() []string {
	return []string{r.mesh}
}

func (r *caRequest) Name() string {
	return "mesh_ca:secret:" + r.mesh
}

type identityRequest struct {
	mesh string
}

func (r *identityRequest) Name() string {
	return "identity_cert:secret:" + r.mesh
}

var _ = Describe("CreateDownstreamTlsContext()", func() {
	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			expected string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				mesh := &core_mesh.MeshResource{
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
				}

				// when
				snippet, err := v3.CreateDownstreamTlsContext(
					&caRequest{mesh: mesh.GetMeta().GetName()},
					&identityRequest{mesh: mesh.GetMeta().GetName()},
				)
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
                      matchTypedSubjectAltNames:
                      - matcher:
                          prefix: spiffe://default/
                        sanType: URI
                    validationContextSdsSecretConfig:
                      name: mesh_ca:secret:default
                      sdsConfig:
                        ads: {}
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert:secret:default
                    sdsConfig:
                      ads: {}
                      resourceApiVersion: V3
                requireClientCertificate: true`,
			}),
		)
	})
})

var _ = Describe("CreateUpstreamTlsContext()", func() {

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			upstreamService string
			expected        string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				mesh := "default"

				// when
				snippet, err := v3.CreateUpstreamTlsContext(
					&identityRequest{mesh: mesh},
					&caRequest{mesh: mesh},
					given.upstreamService,
					"",
				)
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
                      matchTypedSubjectAltNames:
                      - matcher:
                          exact: spiffe://default/backend
                        sanType: URI
                    validationContextSdsSecretConfig:
                      name: mesh_ca:secret:default
                      sdsConfig:
                        ads: {}
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert:secret:default
                    sdsConfig:
                      ads: {}
                      resourceApiVersion: V3`,
			}),
		)
	})
})
