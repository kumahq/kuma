package reconcile_test

import (
	"context"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	mads_cache "github.com/kumahq/kuma/pkg/mads/v1/cache"
	mads_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	. "github.com/kumahq/kuma/pkg/mads/v1/reconcile"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("snapshotGenerator", func() {
	Describe("GenerateSnapshot()", func() {
		var resourceManager core_manager.ResourceManager
		var store core_store.ResourceStore
		node1Id := "one"
		node2Id := "two"

		BeforeEach(func() {
			store = memory.NewStore()
			resourceManager = core_manager.NewResourceManager(store)
		})

		type testCase struct {
			meshes      []*core_mesh.MeshResource
			meshMetrics []*v1alpha1.MeshMetricResource
			dataplanes  []*core_mesh.DataplaneResource
			expected    *mads_cache.Snapshot
			expected2   *mads_cache.Snapshot
		}

		DescribeTable("",
			func(given testCase) {
				// setup
				ctx := context.Background()
				node1 := corev3.Node{Id: node1Id}
				node2 := corev3.Node{Id: node2Id}
				for _, mesh := range given.meshes {
					// when
					err := resourceManager.Create(ctx, mesh, core_store.CreateBy(core_model.MetaToResourceKey(mesh.GetMeta())))
					// then
					Expect(err).ToNot(HaveOccurred())
				}
				for _, dataplane := range given.dataplanes {
					// when
					err := resourceManager.Create(ctx, dataplane, core_store.CreateBy(core_model.MetaToResourceKey(dataplane.GetMeta())))
					// then
					Expect(err).ToNot(HaveOccurred())
				}
				for _, meshMetric := range given.meshMetrics {
					// when
					err := resourceManager.Create(ctx, meshMetric, core_store.CreateBy(core_model.MetaToResourceKey(meshMetric.GetMeta())))
					// then
					Expect(err).ToNot(HaveOccurred())
				}

				// given
				snapshotter := NewSnapshotGenerator(resourceManager, mads_generator.MonitoringAssignmentsGenerator{})
				// when
				snapshot, err := snapshotter.GenerateSnapshot(context.Background(), &node1)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(snapshot).To(Equal(given.expected))

				// when
				snapshot2, err := snapshotter.GenerateSnapshot(context.Background(), &node2)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(snapshot2).To(Equal(given.expected))
			},
			Entry("no Meshes, no Dataplanes", testCase{
				expected: mads_cache.NewSnapshot("", nil),
			}),
			Entry("no Meshes with Prometheus enabled", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					samples.DataplaneBackendBuilder().
						WithName("backend-01").
						Build(),
				},
				expected: mads_cache.NewSnapshot("", nil),
			}),
			Entry("Mesh with Prometheus enabled but no Dataplanes", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{
							Metrics: &mesh_proto.Metrics{
								EnabledBackend: "prometheus-1",
								Backends: []*mesh_proto.MetricsBackend{
									{
										Name: "prometheus-1",
										Type: mesh_proto.MetricsPrometheusType,
										Conf: proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 1234,
											Path: "/non-standard-path",
										}),
									},
								},
							},
						},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					samples.DataplaneBackendBuilder().
						WithName("backend-01").
						Build(),
				},
				expected: mads_cache.NewSnapshot("", nil),
			}),
			Entry("Mesh with Prometheus enabled and some Dataplanes", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{
							Metrics: &mesh_proto.Metrics{
								EnabledBackend: "prometheus-1",
								Backends: []*mesh_proto.MetricsBackend{
									{
										Name: "prometheus-1",
										Type: mesh_proto.MetricsPrometheusType,
										Conf: proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 1234,
										}),
									},
								},
							},
						},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					builders.Dataplane().
						WithName("backend-01").
						WithAddress("192.168.0.1").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend", "env", "prod").
						Build(),
					builders.Dataplane().
						WithName("backend-02").
						WithMesh("demo").
						WithAddress("192.168.0.2").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend", "env", "intg").
						Build(),
					builders.Dataplane().
						WithName("web-01").
						WithMesh("demo").
						WithAddress("192.168.0.3").
						WithInboundOfTags(mesh_proto.ServiceTag, "web", "env", "test").
						WithPrometheusMetrics(&mesh_proto.PrometheusMetricsBackendConfig{
							Port: 8765,
							Path: "/even-more-non-standard-path",
						}).
						Build(),
				},
				// TODO: generate this resource map on the fly using the mads/v1/generator pkg
				expected: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
					"/meshes/demo/dataplanes/backend-02": &observability_v1.MonitoringAssignment{
						Mesh:    "demo",
						Service: "backend",
						Targets: []*observability_v1.MonitoringAssignment_Target{{
							Name:    "backend-02",
							Address: "192.168.0.2:1234",
							Scheme:  "http",
							Labels: map[string]string{
								"env":              "intg",
								"envs":             ",intg,",
								"kuma_io_service":  "backend",
								"kuma_io_services": ",backend,",
							},
						}},
					},
					"/meshes/demo/dataplanes/web-01": &observability_v1.MonitoringAssignment{
						Mesh:    "demo",
						Service: "web",
						Targets: []*observability_v1.MonitoringAssignment_Target{{
							Name:        "web-01",
							Address:     "192.168.0.3:8765",
							Scheme:      "http",
							MetricsPath: "/even-more-non-standard-path",
							Labels: map[string]string{
								"env":              "test",
								"envs":             ",test,",
								"kuma_io_service":  "web",
								"kuma_io_services": ",web,",
							},
						}},
					},
				}),
			}),
			FEntry("MeshMetric with Prometheus enabled", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					samples.DataplaneBackendBuilder().
						WithName("backend-01").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: common_api.TargetRef{
								Kind: common_api.Mesh,
								Name: "default",
							},
							Default:   v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port: 5670,
											Path: "/metrics",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled
											},
										},
									},
								},
							},
						},
					},
				},
				expected: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
					"/meshes/demo/dataplanes/backend-02": &observability_v1.MonitoringAssignment{
						Mesh:    "demo",
						Service: "backend",
						Targets: []*observability_v1.MonitoringAssignment_Target{{
							Name:    "backend-02",
							Address: "192.168.0.2:1234",
							Scheme:  "http",
							Labels: map[string]string{
								"env":              "intg",
								"envs":             ",intg,",
								"kuma_io_service":  "backend",
								"kuma_io_services": ",backend,",
							},
						}},
					},
				}),
			}),
		)
	})
})
