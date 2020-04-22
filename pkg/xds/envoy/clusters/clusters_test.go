package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/clusters"

	"github.com/golang/protobuf/ptypes"

	envoy_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Clusters", func() {

	It("should generate 'local' Cluster", func() {
		// given
		expected := `
        name: localhost:8080
        altStatName: localhost_8080
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
		resource := CreateLocalCluster("localhost:8080", "127.0.0.1", 8080)

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
        lbPolicy: CLUSTER_PROVIDED
        connectTimeout: 5s
`
		// when
		resource := CreatePassThroughCluster("pass_through")

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	Describe("'EDS' Cluster", func() {

		type testCase struct {
			ctx      xds_context.Context
			metadata core_xds.DataplaneMetadata
			expected string
		}

		DescribeTable("should generate 'EDS' Cluster",
			func(given testCase) {
				// when
				resource, err := CreateEdsCluster(given.ctx, "192.168.0.1:8080", &given.metadata)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := util_proto.ToYAML(resource)
				// then
				Expect(err).ToNot(HaveOccurred())

				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("without mTLS", testCase{
				ctx: xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{},
					},
				},
				metadata: core_xds.DataplaneMetadata{},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                altStatName: "192_168_0_1_8080"
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
				},
				metadata: core_xds.DataplaneMetadata{},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                altStatName: "192_168_0_1_8080"
                transportSocket:
                  name: envoy.transport_sockets.tls
                  typedConfig:
                    '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
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
				},
				metadata: core_xds.DataplaneMetadata{
					DataplaneTokenPath: "/var/secret/token",
				},
				expected: `
                connectTimeout: 5s
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: 192.168.0.1:8080
                altStatName: "192_168_0_1_8080"
                transportSocket:
                  name: envoy.transport_sockets.tls
                  typedConfig:
                    '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
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

	Describe("ClusterWithHealthChecks()", func() {

		type testCase struct {
			healthCheck *mesh_core.HealthCheckResource
			expected    string
		}
		DescribeTable("should add health checks to a given Cluster",
			func(given testCase) {
				// given
				cluster := &envoy_v2.Cluster{
					Name: "example",
				}
				// when
				metadata := ClusterWithHealthChecks(cluster, given.healthCheck)
				// and
				actual, err := util_proto.ToYAML(metadata)
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("`nil` HealthCheck", testCase{
				healthCheck: nil,
				expected: `
                name: example
`,
			}),
			Entry("HealthCheck with neither active nor passive checks", testCase{
				healthCheck: &mesh_core.HealthCheckResource{},
				expected: `
                name: example
`,
			}),
			Entry("HealthCheck with active checks", testCase{
				healthCheck: &mesh_core.HealthCheckResource{
					Spec: mesh_proto.HealthCheck{
						Sources: []*mesh_proto.Selector{
							{Match: mesh_proto.TagSelector{"service": "backend"}},
						},
						Destinations: []*mesh_proto.Selector{
							{Match: mesh_proto.TagSelector{"service": "redis"}},
						},
						Conf: &mesh_proto.HealthCheck_Conf{
							ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
								Interval:           ptypes.DurationProto(5 * time.Second),
								Timeout:            ptypes.DurationProto(4 * time.Second),
								UnhealthyThreshold: 3,
								HealthyThreshold:   2,
							},
						},
					},
				},
				expected: `
                healthChecks:
                - healthyThreshold: 2
                  interval: 5s
                  tcpHealthCheck: {}
                  timeout: 4s
                  unhealthyThreshold: 3
                name: example
`,
			}),
			Entry("HealthCheck with passive checks", testCase{
				healthCheck: &mesh_core.HealthCheckResource{
					Spec: mesh_proto.HealthCheck{
						Sources: []*mesh_proto.Selector{
							{Match: mesh_proto.TagSelector{"service": "backend"}},
						},
						Destinations: []*mesh_proto.Selector{
							{Match: mesh_proto.TagSelector{"service": "redis"}},
						},
						Conf: &mesh_proto.HealthCheck_Conf{
							PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
								UnhealthyThreshold: 20,
								PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
							},
						},
					},
				},
				expected: `
                name: example
                outlierDetection:
                  consecutive5xx: 20
                  interval: 30s
`,
			}),
			Entry("HealthCheck with both active and passive checks", testCase{
				healthCheck: &mesh_core.HealthCheckResource{
					Spec: mesh_proto.HealthCheck{
						Sources: []*mesh_proto.Selector{
							{Match: mesh_proto.TagSelector{"service": "backend"}},
						},
						Destinations: []*mesh_proto.Selector{
							{Match: mesh_proto.TagSelector{"service": "redis"}},
						},
						Conf: &mesh_proto.HealthCheck_Conf{
							ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
								Interval:           ptypes.DurationProto(5 * time.Second),
								Timeout:            ptypes.DurationProto(4 * time.Second),
								UnhealthyThreshold: 3,
								HealthyThreshold:   2,
							},
							PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
								UnhealthyThreshold: 20,
								PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
							},
						},
					},
				},
				expected: `
                healthChecks:
                - healthyThreshold: 2
                  interval: 5s
                  tcpHealthCheck: {}
                  timeout: 4s
                  unhealthyThreshold: 3
                name: example
                outlierDetection:
                  consecutive5xx: 20
                  interval: 30s
`,
			}),
		)
	})

})
