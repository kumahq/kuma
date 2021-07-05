package v3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("CreateDownstreamTlsContext()", func() {

	Context("when mTLS is disabled on a given Mesh", func() {

		It("should return `nil`", func() {
			// given
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: mesh_core.NewMeshResource(),
				},
			}
			metadata := &core_xds.DataplaneMetadata{}

			// when
			snippet, err := v3.CreateDownstreamTlsContext(ctx, metadata)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(snippet).To(BeNil())
		})
	})

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			metadata *core_xds.DataplaneMetadata
			expected string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					ConnectionInfo: xds_context.ConnectionInfo{
						Authority: "kuma-control-plane:5677",
					},
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsTlsCert: []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
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
				snippet, err := v3.CreateDownstreamTlsContext(ctx, given.metadata)
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
				metadata: nil,
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - envoyGrpc:
                              clusterName: ads_cluster
                          transportApiVersion: V3
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - envoyGrpc:
                            clusterName: ads_cluster
                        transportApiVersion: V3
                      resourceApiVersion: V3
                requireClientCertificate: true
`,
			}),
			Entry("dataplane without a token", testCase{
				metadata: &core_xds.DataplaneMetadata{},
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - envoyGrpc:
                              clusterName: ads_cluster
                          transportApiVersion: V3
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - envoyGrpc:
                            clusterName: ads_cluster
                        transportApiVersion: V3
                      resourceApiVersion: V3
                requireClientCertificate: true
`,
			}),
			Entry("dataplane with a token", testCase{
				metadata: &core_xds.DataplaneMetadata{
					DataplaneToken: "sampletoken",
				},
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - envoyGrpc:
                              clusterName: ads_cluster
                            initialMetadata:
                            - key: authorization
                              value: sampletoken
                          transportApiVersion: V3
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - envoyGrpc:
                            clusterName: ads_cluster
                          initialMetadata:
                          - key: authorization
                            value: sampletoken
                        transportApiVersion: V3
                      resourceApiVersion: V3
                requireClientCertificate: true
`,
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
					Resource: mesh_core.NewMeshResource(),
				},
			}
			metadata := &core_xds.DataplaneMetadata{}

			// when
			snippet, err := v3.CreateUpstreamTlsContext(ctx, metadata, "backend", "backend")
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(snippet).To(BeNil())
		})
	})

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			metadata        *core_xds.DataplaneMetadata
			upstreamService string
			expected        string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					ConnectionInfo: xds_context.ConnectionInfo{
						Authority: "kuma-control-plane:5677",
					},
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsTlsCert: []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
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
				snippet, err := v3.CreateUpstreamTlsContext(ctx, given.metadata, given.upstreamService, "")
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
				metadata:        nil,
				upstreamService: "backend",
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - envoyGrpc:
                              clusterName: ads_cluster
                          transportApiVersion: V3
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - envoyGrpc:
                            clusterName: ads_cluster
                        transportApiVersion: V3
                      resourceApiVersion: V3
`,
			}),
			Entry("dataplane without a token", testCase{
				metadata:        &core_xds.DataplaneMetadata{},
				upstreamService: "backend",
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - envoyGrpc:
                              clusterName: ads_cluster
                          transportApiVersion: V3
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - envoyGrpc:
                            clusterName: ads_cluster
                        transportApiVersion: V3
                      resourceApiVersion: V3
`,
			}),
			Entry("dataplane with token", testCase{
				metadata: &core_xds.DataplaneMetadata{
					DataplaneToken: "sampletoken",
				},
				upstreamService: "backend",
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - envoyGrpc:
                              clusterName: ads_cluster
                            initialMetadata:
                            - key: authorization
                              value: sampletoken
                          transportApiVersion: V3
                        resourceApiVersion: V3
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - envoyGrpc:
                            clusterName: ads_cluster
                          initialMetadata:
                          - key: authorization
                            value: sampletoken
                        transportApiVersion: V3
                      resourceApiVersion: V3
`,
			}),
		)
	})
})
