package clusters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/envoy"
	"github.com/Kong/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("EdsClusterConfigurer", func() {

	type testCase struct {
		clusterName   string
		clientService string
		tags          []envoy.Tags
		ctx           xds_context.Context
		metadata      *core_xds.DataplaneMetadata
		expected      string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder().
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.ClientSideMTLS(given.ctx, given.metadata, given.clientService, given.tags)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("cluster with mTLS", testCase{
			clusterName:   "testCluster",
			clientService: "backend",
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsLocation: "kuma-control-plane:5677",
					SdsTlsCert:  []byte("CERTIFICATE"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "default",
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
			},
			// no tags therefore SNI is empty
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            name: testCluster
            transportSocket:
              name: envoy.transport_sockets.tls
              typedConfig:
                '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
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
                          - googleGrpc:
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
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
            type: EDS`,
		}),
		Entry("cluster with many different tag sets", testCase{
			clusterName:   "testCluster",
			clientService: "backend",
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsLocation: "kuma-control-plane:5677",
					SdsTlsCert:  []byte("CERTIFICATE"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "default",
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
			},
			tags: []envoy.Tags{
				map[string]string{
					"service": "backend",
					"cluster": "1",
				},
				map[string]string{
					"service": "backend",
					"cluster": "2",
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            name: testCluster
            transportSocketMatches:
            - match:
                cluster: "1"
              name: backend{cluster=1}
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
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
                            - googleGrpc:
                                channelCredentials:
                                  sslCredentials:
                                    rootCerts:
                                      inlineBytes: Q0VSVElGSUNBVEU=
                                statPrefix: sds_mesh_ca
                                targetUri: kuma-control-plane:5677
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
                  sni: backend{cluster=1}
            - match:
                cluster: "2"
              name: backend{cluster=2}
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
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
                            - googleGrpc:
                                channelCredentials:
                                  sslCredentials:
                                    rootCerts:
                                      inlineBytes: Q0VSVElGSUNBVEU=
                                statPrefix: sds_mesh_ca
                                targetUri: kuma-control-plane:5677
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
                  sni: backend{cluster=2}
            type: EDS`,
		}),
		Entry("cluster with mTLS and credentials", testCase{
			clusterName:   "testCluster",
			clientService: "backend",
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsLocation: "kuma-control-plane:5677",
					SdsTlsCert:  []byte("CERTIFICATE"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "default",
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
			},
			metadata: &core_xds.DataplaneMetadata{
				DataplaneTokenPath: "/var/secret/token",
			},
			tags: []envoy.Tags{
				{
					"service": "backend",
					"version": "v1",
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            name: testCluster
            transportSocket:
              name: envoy.transport_sockets.tls
              typedConfig:
                '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
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
                          - googleGrpc:
                              callCredentials:
                              - fromPlugin:
                                  name: envoy.grpc_credentials.file_based_metadata
                                  typedConfig:
                                    '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                    secretData:
                                      filename: /var/secret/token
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
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
                                    filename: /var/secret/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                sni: backend{version=v1}
            type: EDS`,
		}),
	)
})
