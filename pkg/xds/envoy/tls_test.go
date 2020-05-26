package envoy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("CreateDownstreamTlsContext()", func() {

	Context("when mTLS is disabled on a given Mesh", func() {

		It("should return `nil`", func() {
			// given
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{},
				},
			}
			metadata := &core_xds.DataplaneMetadata{}

			// when
			snippet, err := CreateDownstreamTlsContext(ctx, metadata)
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
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
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

				// when
				snippet, err := CreateDownstreamTlsContext(ctx, given.metadata)
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
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                  validationContextSdsSecretConfig:
                    name: mesh_ca
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_mesh_ca
                            targetUri: kuma-control-plane:5677
                requireClientCertificate: true
`,
			}),
			Entry("dataplane without a token", testCase{
				metadata: &core_xds.DataplaneMetadata{},
				expected: `
                commonTlsContext:
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                  validationContextSdsSecretConfig:
                    name: mesh_ca
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_mesh_ca
                            targetUri: kuma-control-plane:5677
                requireClientCertificate: true
`,
			}),
			Entry("dataplane with a token", testCase{
				metadata: &core_xds.DataplaneMetadata{
					DataplaneTokenPath: "/path/to/token",
				},
				expected: `
                commonTlsContext:
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            callCredentials:
                            - fromPlugin:
                                name: envoy.grpc_credentials.file_based_metadata
                                typedConfig:
                                  '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                  secretData:
                                    filename: /path/to/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                  validationContextSdsSecretConfig:
                    name: mesh_ca
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            callCredentials:
                            - fromPlugin:
                                name: envoy.grpc_credentials.file_based_metadata
                                typedConfig:
                                  '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                  secretData:
                                    filename: /path/to/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_mesh_ca
                            targetUri: kuma-control-plane:5677
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
					Resource: &mesh_core.MeshResource{},
				},
			}
			metadata := &core_xds.DataplaneMetadata{}

			// when
			snippet, err := CreateUpstreamTlsContext(ctx, metadata)
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
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
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

				// when
				snippet, err := CreateUpstreamTlsContext(ctx, given.metadata)
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
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                  validationContextSdsSecretConfig:
                    name: mesh_ca
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_mesh_ca
                            targetUri: kuma-control-plane:5677
`,
			}),
			Entry("dataplane without a token", testCase{
				metadata: &core_xds.DataplaneMetadata{},
				expected: `
                commonTlsContext:
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                  validationContextSdsSecretConfig:
                    name: mesh_ca
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_mesh_ca
                            targetUri: kuma-control-plane:5677
`,
			}),
			Entry("dataplane with a token", testCase{
				metadata: &core_xds.DataplaneMetadata{
					DataplaneTokenPath: "/path/to/token",
				},
				expected: `
                commonTlsContext:
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            callCredentials:
                            - fromPlugin:
                                name: envoy.grpc_credentials.file_based_metadata
                                typedConfig:
                                  '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                  secretData:
                                    filename: /path/to/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                  validationContextSdsSecretConfig:
                    name: mesh_ca
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            callCredentials:
                            - fromPlugin:
                                name: envoy.grpc_credentials.file_based_metadata
                                typedConfig:
                                  '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                  secretData:
                                    filename: /path/to/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_mesh_ca
                            targetUri: kuma-control-plane:5677
`,
			}),
		)
	})
})
