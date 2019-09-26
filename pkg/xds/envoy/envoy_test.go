package envoy_test

import (
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

var _ = Describe("Envoy", func() {

	It("should generate 'static' Endpoints", func() {
		// given
		expected := `
        clusterName: localhost:8080
        endpoints:
        - lbEndpoints:
          - endpoint:
              address:
                socketAddress:
                  address: 127.0.0.1
                  portValue: 8080
`
		// when
		resource := envoy.CreateStaticEndpoint("localhost:8080", "127.0.0.1", 8080)

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'local' Cluster", func() {
		// given
		expected := `
        name: localhost:8080
        type: STATIC
        connectTimeout: 5s
        loadAssignment:
          clusterName: localhost:8080
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 8080
`
		// when
		resource := envoy.CreateLocalCluster("localhost:8080", "127.0.0.1", 8080)

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'pass-through' Cluster", func() {
		// given
		expected := `
        name: pass_through
        type: ORIGINAL_DST
        lbPolicy: ORIGINAL_DST_LB
        connectTimeout: 5s
`
		// when
		resource := envoy.CreatePassThroughCluster("pass_through")

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	Describe("'EDS' Cluster", func() {

		type testCase struct {
			ctx      xds_context.Context
			expected string
		}

		DescribeTable("should generate 'EDS' Cluster",
			func(given testCase) {
				// when
				resource := envoy.CreateEdsCluster(given.ctx, "192.168.0.1:8080")

				// then
				actual, err := util_proto.ToYAML(resource)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("without mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                type: EDS
`,
			}),
			Entry("with mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                tlsContext:
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
                type: EDS
`,
			}),
			Entry("with mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                tlsContext:
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
                type: EDS
`,
			}),
			Entry("with mTLS and Dataplane credentials", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "/var/secret/token",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                tlsContext:
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
                                      filename: /var/secret/token
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
                                      filename: /var/secret/token
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                type: EDS
`,
			}),
		)
	})

	It("should generate ClusterLoadAssignment", func() {
		// given
		expected := `
        clusterName: 127.0.0.1:8080
        endpoints:
        - lbEndpoints:
          - endpoint:
              address:
                socketAddress:
                  address: 192.168.0.1
                  portValue: 8081
          - endpoint:
              address:
                socketAddress:
                  address: 192.168.0.2
                  portValue: 8082
`
		// when
		resource := envoy.CreateClusterLoadAssignment("127.0.0.1:8080",
			[]net.SRV{
				{Target: "192.168.0.1", Port: 8081},
				{Target: "192.168.0.2", Port: 8082},
			})

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	Describe("'inbound' listener", func() {

		type testCase struct {
			ctx      xds_context.Context
			virtual  bool
			expected string
		}

		DescribeTable("should generate 'inbound' Listener",
			func(given testCase) {
				// given
				permissions := &mesh_core.TrafficPermissionResourceList{
					Items: []*mesh_core.TrafficPermissionResource{
						&mesh_core.TrafficPermissionResource{
							Meta: &test_model.ResourceMeta{
								Name:      "tp-1",
								Mesh:      "default",
								Namespace: "default",
							},
							Spec: mesh_proto.TrafficPermission{
								Rules: []*mesh_proto.TrafficPermission_Rule{
									{
										Sources: []*mesh_proto.Selector{
											{
												Match: map[string]string{
													"service": "web1",
													"version": "1.0",
												},
											},
										},
										Destinations: []*mesh_proto.Selector{
											{
												Match: map[string]string{
													"service": "backend1",
													"env":     "dev",
												},
											},
										},
									},
								},
							},
						},
					},
				}

				// when
				resource := envoy.CreateInboundListener(given.ctx, "inbound:192.168.0.1:8080", "192.168.0.1", 8080, "localhost:8080", given.virtual, permissions)

				// then
				actual, err := util_proto.ToYAML(resource)
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("without transparent proxying", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				virtual: false,
				expected: `
                name: inbound:192.168.0.1:8080
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      cluster: localhost:8080
                      statPrefix: localhost:8080
`,
			}),
			Entry("with access logs", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled:     false,
						LoggingEnabled: true,
						LoggingPath:    "/tmp/access.logs",
					},
				},
				virtual: false,
				expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.tcp_proxy
    typedConfig:
      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
      accessLog:
      - name: envoy.file_access_log
        typedConfig:
          '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
          format: |
            [%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS%->%UPSTREAM_HOST% took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
          path: /tmp/access.logs
      cluster: localhost:8080
      statPrefix: localhost:8080
name: inbound:192.168.0.1:8080
`,
			}),
			Entry("with transparent proxying", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				virtual: true,
				expected: `
                name: inbound:192.168.0.1:8080
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      cluster: localhost:8080
                      statPrefix: localhost:8080
                deprecatedV1:
                  bindToPort: false
`,
			}),
			Entry("with mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual: false,
				expected: `
          address:
            socketAddress:
              address: 192.168.0.1
              portValue: 8080
          filterChains:
          - filters:
            - name: envoy.filters.network.rbac
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.rbac.v2.RBAC
                rules:
                  policies:
                    default.tp-1:
                      permissions:
                      - any: true
                      principals:
                      - authenticated:
                          principalName:
                            exact: spiffe://default/web1
                statPrefix: inbound:192.168.0.1:8080
            - name: envoy.tcp_proxy
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                cluster: localhost:8080
                statPrefix: localhost:8080
            tlsContext:
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
          name: inbound:192.168.0.1:8080
`,
			}),
			Entry("with mTLS and Dataplane credentials", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "/var/secret/token",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual: false,
				expected: `
          address:
            socketAddress:
              address: 192.168.0.1
              portValue: 8080
          filterChains:
          - filters:
            - name: envoy.filters.network.rbac
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.rbac.v2.RBAC
                rules:
                  policies:
                    default.tp-1:
                      permissions:
                      - any: true
                      principals:
                      - authenticated:
                          principalName:
                            exact: spiffe://default/web1
                statPrefix: inbound:192.168.0.1:8080
            - name: envoy.tcp_proxy
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                cluster: localhost:8080
                statPrefix: localhost:8080
            tlsContext:
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
                                  filename: /var/secret/token
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
                                  filename: /var/secret/token
                          channelCredentials:
                            sslCredentials:
                              rootCerts:
                                inlineBytes: Q0VSVElGSUNBVEU=
                          credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                          statPrefix: sds_mesh_ca
                          targetUri: kuma-control-plane:5677
              requireClientCertificate: true
          name: inbound:192.168.0.1:8080
`,
			}),
		)
	})

	Describe("'outbound' listener", func() {

		type testCase struct {
			ctx      xds_context.Context
			virtual  bool
			expected string
			logs     []*mesh_proto.LoggingBackend
		}

		DescribeTable("should generate 'outbound' Listener",
			func(given testCase) {
				// when
				resource, err := envoy.CreateOutboundListener(given.ctx, "outbound:127.0.0.1:18080", "127.0.0.1", 18080, "outbound:127.0.0.1:18080", given.virtual, given.logs)
				Expect(err).ToNot(HaveOccurred())

				// then
				actual, err := util_proto.ToYAML(resource)
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("without transparent proxying", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				virtual: false,
				expected: `
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 18080
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      cluster: outbound:127.0.0.1:18080
                      statPrefix: outbound:127.0.0.1:18080
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with transparent proxying", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				virtual: true,
				expected: `
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 18080
                deprecatedV1:
                  bindToPort: false
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      cluster: outbound:127.0.0.1:18080
                      statPrefix: outbound:127.0.0.1:18080
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual: false,
				expected: `
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 18080
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      cluster: outbound:127.0.0.1:18080
                      statPrefix: outbound:127.0.0.1:18080
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with mTLS and Dataplane credentials", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation:        "kuma-control-plane:5677",
						SdsTlsCert:         []byte("CERTIFICATE"),
						DataplaneTokenFile: "/var/secret/token",
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual: false,
				expected: `
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 18080
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      cluster: outbound:127.0.0.1:18080
                      statPrefix: outbound:127.0.0.1:18080
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with traffic logs", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				logs: []*mesh_proto.LoggingBackend{
					{
						Name: "file-1",
						Type: &mesh_proto.LoggingBackend_File_{
							File: &mesh_proto.LoggingBackend_File{
								Path: "/tmp/log",
							},
						},
					},
					{
						Name:   "file-2",
						Format: "custom format",
						Type: &mesh_proto.LoggingBackend_File_{
							File: &mesh_proto.LoggingBackend_File{
								Path: "/tmp/log2",
							},
						},
					},
				},
				expected: `
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 18080
                filterChains:
                - filters:
                  - name: envoy.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                      accessLog:
                      - name: envoy.file_access_log
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
                          format: |
                            [%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS%->%UPSTREAM_HOST%(%UPSTREAM_CLUSTER%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                          path: /tmp/log
                      - name: envoy.file_access_log
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
                          format: custom format
                          path: /tmp/log2
                      cluster: outbound:127.0.0.1:18080
                      statPrefix: outbound:127.0.0.1:18080
                name: outbound:127.0.0.1:18080
`,
			}),
		)
	})

	It("should generate 'catch all' Listener", func() {
		// given
		expected := `
        name: catch_all
        address:
          socketAddress:
            address: 0.0.0.0
            portValue: 15001
        filterChains:
        - filters:
          - name: envoy.tcp_proxy
            typedConfig:
              '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
              cluster: pass_through
              statPrefix: pass_through
        useOriginalDst: true
`
		ctx := xds_context.Context{}

		// when
		resource := envoy.CreateCatchAllListener(ctx, "catch_all", "0.0.0.0", 15001, "pass_through")

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
