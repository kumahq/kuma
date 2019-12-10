package envoy_test

import (
	"github.com/Kong/kuma/pkg/core/xds"

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

	Describe("CreateLbMetadata()", func() {

		It("should handle `nil` map of tags", func() {
			// when
			metadata := envoy.CreateLbMetadata(nil)
			// then
			Expect(metadata).To(BeNil())
		})

		It("should handle empty map of tags", func() {
			// when
			metadata := envoy.CreateLbMetadata(map[string]string{})
			// then
			Expect(metadata).To(BeNil())
		})

		type testCase struct {
			tags     map[string]string
			expected string
		}
		DescribeTable("should generate Envoy metadata",
			func(given testCase) {
				// when
				metadata := envoy.CreateLbMetadata(given.tags)
				// and
				actual, err := util_proto.ToYAML(metadata)
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("map with 1 tag", testCase{
				tags: map[string]string{
					"service": "redis",
				},
				expected: `
        filterMetadata:
          envoy.lb:
            service: redis
`,
			}),
			Entry("map with multiple tags", testCase{
				tags: map[string]string{
					"service": "redis",
					"version": "v1",
					"region":  "eu",
				},
				expected: `
        filterMetadata:
          envoy.lb:
            service: redis
            version: v1
            region: eu
`,
			}),
		)
	})

	Describe("'EDS' Cluster", func() {

		type testCase struct {
			ctx      xds_context.Context
			metadata xds.DataplaneMetadata
			expected string
		}

		DescribeTable("should generate 'EDS' Cluster",
			func(given testCase) {
				// when
				resource := envoy.CreateEdsCluster(given.ctx, "192.168.0.1:8080", &given.metadata)

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
				metadata: xds.DataplaneMetadata{},
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
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				metadata: xds.DataplaneMetadata{},
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
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				metadata: xds.DataplaneMetadata{
					DataplaneTokenPath: "/var/secret/token",
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

	Describe("ClusterLoadAssignment()", func() {
		type testCase struct {
			cluster   string
			endpoints []xds.Endpoint
			expected  string
		}
		DescribeTable("should generate ClusterLoadAssignment",
			func(given testCase) {
				// when
				resource := envoy.CreateClusterLoadAssignment(given.cluster, given.endpoints)

				// then
				actual, err := util_proto.ToYAML(resource)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("without tags", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []xds.Endpoint{
					{Target: "192.168.0.1", Port: 8081},
					{Target: "192.168.0.2", Port: 8082},
				},
				expected: `
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
`,
			}),
			Entry("with tags", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []xds.Endpoint{
					{Target: "192.168.0.1", Port: 8081, Tags: map[string]string{"service": "backend", "region": "us"}},
					{Target: "192.168.0.2", Port: 8082, Tags: map[string]string{"service": "backend", "region": "eu"}},
				},
				expected: `
        clusterName: 127.0.0.1:8080
        endpoints:
        - lbEndpoints:
          - endpoint:
              address:
                socketAddress:
                  address: 192.168.0.1
                  portValue: 8081
            metadata:
              filterMetadata:
                envoy.lb:
                  region: us
                  service: backend
          - endpoint:
              address:
                socketAddress:
                  address: 192.168.0.2
                  portValue: 8082
            metadata:
              filterMetadata:
                envoy.lb:
                  region: eu
                  service: backend
`,
			}),
		)
	})

	Describe("'inbound' listener", func() {

		type testCase struct {
			ctx      xds_context.Context
			virtual  bool
			expected string
			metadata xds.DataplaneMetadata
		}

		DescribeTable("should generate 'inbound' Listener",
			func(given testCase) {
				// given
				permissions := &mesh_core.TrafficPermissionResourceList{
					Items: []*mesh_core.TrafficPermissionResource{
						&mesh_core.TrafficPermissionResource{
							Meta: &test_model.ResourceMeta{
								Name: "tp-1",
								Mesh: "default",
							},
							Spec: mesh_proto.TrafficPermission{
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
				}

				// when
				resource := envoy.CreateInboundListener(given.ctx, "inbound:192.168.0.1:8080", "192.168.0.1", 8080, "localhost:8080", given.virtual, permissions, &given.metadata)

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
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
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
                    tp-1:
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
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual: false,
				metadata: xds.DataplaneMetadata{
					DataplaneTokenPath: "/var/secret/token",
				},
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
                    tp-1:
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
			clusters []envoy.ClusterInfo
			expected string
			log      *mesh_proto.LoggingBackend
		}

		singleCluster := []envoy.ClusterInfo{{
			Name:   "db",
			Weight: 100,
			Tags:   map[string]string{"service": "db"},
		}}

		DescribeTable("should generate 'outbound' Listener",
			func(given testCase) {
				proxy := xds.Proxy{
					Id: xds.ProxyId{
						Name: "backend",
						Mesh: "example",
					},
					Dataplane: &mesh_core.DataplaneResource{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Interface: "192.168.0.1:1234:8765",
									Tags: map[string]string{
										"service": "backend",
									},
								}},
								Outbound: []*mesh_proto.Dataplane_Networking_Outbound{{
									Interface: ":15432",
									Service:   "db",
								}},
							},
						},
					},
				}
				sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
				destinationService := "db"

				// when
				resource, err := envoy.CreateOutboundListener(given.ctx, "outbound:127.0.0.1:18080", "127.0.0.1", 18080, "db", given.clusters, given.virtual, sourceService, destinationService, given.log, &proxy)
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
				virtual:  false,
				clusters: singleCluster,
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
                      cluster: db
                      statPrefix: db
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
				virtual:  true,
				clusters: singleCluster,
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
                      cluster: db
                      statPrefix: db
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual:  false,
				clusters: singleCluster,
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
                      cluster: db
                      statPrefix: db
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with mTLS and Dataplane credentials", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-control-plane:5677",
						SdsTlsCert:  []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				},
				virtual:  false,
				clusters: singleCluster,
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
                      cluster: db
                      statPrefix: db
                name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with file traffic logs", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				clusters: singleCluster,
				log: &mesh_proto.LoggingBackend{
					Name: "file",
					Type: &mesh_proto.LoggingBackend_File_{
						File: &mesh_proto.LoggingBackend_File{
							Path: "/tmp/log",
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
                      [%START_TIME%] 192.168.0.1:0(backend)->%UPSTREAM_HOST%(db) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                    path: /tmp/log
                cluster: db
                statPrefix: db
          name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with tcp traffic logs", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				clusters: singleCluster,
				log: &mesh_proto.LoggingBackend{
					Name:   "tcp",
					Format: "custom format",
					Type: &mesh_proto.LoggingBackend_Tcp_{
						Tcp: &mesh_proto.LoggingBackend_Tcp{
							Address: "127.0.0.1:1234",
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
                - name: envoy.http_grpc_access_log
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.accesslog.v2.HttpGrpcAccessLogConfig
                    commonConfig:
                      grpcService:
                        envoyGrpc:
                          clusterName: access_log_sink
                      logName: 127.0.0.1:1234;custom format
                cluster: db
                statPrefix: db
          name: outbound:127.0.0.1:18080
`,
			}),
			Entry("with multiple weighted clusters", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						TlsEnabled: false,
					},
				},
				virtual: false,
				clusters: []envoy.ClusterInfo{{
					Name:   "db{version=v1}",
					Weight: 10,
					Tags:   map[string]string{"service": "db", "version": "v1"},
				}, {
					Name:   "db{version=v2}",
					Weight: 90,
					Tags:   map[string]string{"service": "db", "version": "v2"},
				}},
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
                      statPrefix: db
                      weightedClusters:
                        clusters:
                        - name: db{version=v1}
                          weight: 10
                        - name: db{version=v2}
                          weight: 90
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
