package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("ServerMtlsConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol core_xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		statsName        string
		clusters         []envoy_common.ClusterSubset
		ctx              xds_context.Context
		metadata         core_xds.DataplaneMetadata
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV2).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV2).
					Configure(ServerSideMTLS(given.ctx, &given.metadata)).
					Configure(TcpProxy(given.statsName, given.clusters...)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy with mTLS", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "localhost:8080", Weight: 200}},
			ctx: xds_context.Context{
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
			},
			expected: `
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost_8080
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.api.v2.auth.DownstreamTlsContext
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
                  requireClientCertificate: true
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
`,
		}),
		Entry("basic tcp_proxy with mTLS and Dataplane credentials", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "localhost:8080", Weight: 200}},
			ctx: xds_context.Context{
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
			},
			metadata: core_xds.DataplaneMetadata{
				DataplaneTokenPath: "/var/secret/token",
			},
			expected: `
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost_8080
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.api.v2.auth.DownstreamTlsContext
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
                  requireClientCertificate: true
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
`,
		}),
	)
})
