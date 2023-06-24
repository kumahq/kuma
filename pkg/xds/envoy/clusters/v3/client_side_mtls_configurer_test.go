package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

var _ = Describe("EdsClusterConfigurer", func() {
	type testCase struct {
		clusterName   string
		clientService string
		tags          []tags.Tags
		mesh          *core_mesh.MeshResource
		expected      string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			tracker := envoy.NewSecretsTracker(given.mesh.GetMeta().GetName(), nil)
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, given.clusterName).
				Configure(clusters.EdsCluster()).
				Configure(clusters.ClientSideMTLS(tracker, given.mesh, given.clientService, true, given.tags)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
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
			mesh: &core_mesh.MeshResource{
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
			// no tags therefore SNI is empty
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: testCluster
            transportSocket:
              name: envoy.transport_sockets.tls
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
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
                      resourceApiVersion: V3
            type: EDS`,
		}),
		Entry("cluster with many different tag sets", testCase{
			clusterName:   "testCluster",
			clientService: "backend",
			mesh: &core_mesh.MeshResource{
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
			tags: []tags.Tags{
				map[string]string{
					"kuma.io/service": "backend",
					"cluster":         "1",
				},
				map[string]string{
					"kuma.io/service": "backend",
					"cluster":         "2",
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: testCluster
            transportSocketMatches:
            - match:
                cluster: "1"
              name: backend{cluster=1,mesh=default}
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
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
                        resourceApiVersion: V3
                  sni: backend{cluster=1,mesh=default}
            - match:
                cluster: "2"
              name: backend{cluster=2,mesh=default}
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
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
                        resourceApiVersion: V3
                  sni: backend{cluster=2,mesh=default}
            type: EDS`,
		}),
		Entry("cluster with mTLS and credentials", testCase{
			clusterName:   "testCluster",
			clientService: "backend",
			mesh: &core_mesh.MeshResource{
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
			tags: []tags.Tags{
				{
					"kuma.io/service": "backend",
					"version":         "v1",
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: testCluster
            transportSocket:
              name: envoy.transport_sockets.tls
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
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
                      resourceApiVersion: V3
                sni: backend{mesh=default,version=v1}
            type: EDS`,
		}),
	)
})
