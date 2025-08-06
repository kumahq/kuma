package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/matchers"
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
		goldenFile    string
		unifiedNaming bool
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			tracker := envoy.NewSecretsTracker(given.mesh.GetMeta().GetName(), nil)
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, given.clusterName).
				Configure(clusters.EdsCluster()).
				Configure(clusters.ClientSideMTLS(tracker, given.unifiedNaming, given.mesh, given.clientService, true, given.tags)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(matchers.MatchGoldenYAML(given.goldenFile))
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
			goldenFile: "testdata/client_side_mtls_configurer/cluster-with-mtls.golden.yaml",
		}),
		Entry("cluster with mTLS and unified naming", testCase{
			clusterName:   "testCluster",
			clientService: "backend",
			unifiedNaming: true,
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
			goldenFile: "testdata/client_side_mtls_configurer/cluster-with-mtls-unified-naming.golden.yaml",
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
			goldenFile: "testdata/client_side_mtls_configurer/cluster-with-many-different-tag-sets.golden.yaml",
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
			goldenFile: "testdata/client_side_mtls_configurer/cluster-with-many-different-tag-sets.golden.yaml",
		}),
	)
})
