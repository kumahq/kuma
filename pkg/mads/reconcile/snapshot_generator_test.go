package reconcile_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/mads/reconcile"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/util/proto"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"

	mads_cache "github.com/Kong/kuma/pkg/mads/cache"
	mads_generator "github.com/Kong/kuma/pkg/mads/generator"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("snapshotGenerator", func() {

	Describe("GenerateSnapshot()", func() {

		var resourceManager core_manager.ResourceManager
		var store core_store.ResourceStore

		BeforeEach(func() {
			store = memory.NewStore()
			resourceManager = core_manager.NewResourceManager(store)
		})

		type testCase struct {
			meshes     []*mesh_core.MeshResource
			dataplanes []*mesh_core.DataplaneResource
			expected   *mads_cache.Snapshot
		}

		DescribeTable("",
			func(given testCase) {
				// setup
				ctx := context.Background()
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

				// given
				snapshotter := NewSnapshotGenerator(resourceManager, mads_generator.MonitoringAssignmentsGenerator{})
				// when
				snapshot, err := snapshotter.GenerateSnapshot(context.Background(), nil)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(snapshot).To(Equal(given.expected))
			},
			Entry("no Meshes, no Dataplanes", testCase{
				expected: mads_cache.NewSnapshot("", nil),
			}),
			Entry("no Meshes with Prometheus enabled", testCase{
				meshes: []*mesh_core.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
					},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "default",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"service": "backend",
									},
								}},
							},
						},
					},
				},
				expected: mads_cache.NewSnapshot("", nil),
			}),
			Entry("Mesh with Prometheus enabled but no Dataplanes", testCase{
				meshes: []*mesh_core.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
							Mesh: "demo",
						},
						Spec: mesh_proto.Mesh{
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
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "default",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"service": "backend",
									},
								}},
							},
						},
					},
				},
				expected: mads_cache.NewSnapshot("", nil),
			}),
			Entry("Mesh with Prometheus enabled and some Dataplanes", testCase{
				meshes: []*mesh_core.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
							Mesh: "demo",
						},
						Spec: mesh_proto.Mesh{
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
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "default",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"service": "backend",
										"env":     "prod",
									},
								}},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-02",
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        443,
									ServicePort: 8443,
									Tags: map[string]string{
										"service": "backend",
										"env":     "intg",
									},
								}},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "web-01",
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.3",
								Gateway: &mesh_proto.Dataplane_Networking_Gateway{
									Tags: map[string]string{
										"service": "web",
										"env":     "test",
									},
								},
							},
							Metrics: &mesh_proto.MetricsBackend{
								Name: "prometheus-1",
								Type: mesh_proto.MetricsPrometheusType,
								Conf: proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
									Port: 8765,
									Path: "/even-more-non-standard-path",
								}),
							},
						},
					},
				},
				expected: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
					"/meshes/demo/dataplanes/backend-02": &observability_proto.MonitoringAssignment{
						Name: "/meshes/demo/dataplanes/backend-02",
						Targets: []*observability_proto.MonitoringAssignment_Target{{
							Labels: map[string]string{
								"__address__": "192.168.0.2:1234",
							},
						}},
						Labels: map[string]string{
							"__scheme__":       "http",
							"__metrics_path__": "/non-standard-path",
							"job":              "backend",
							"instance":         "backend-02",
							"mesh":             "demo",
							"dataplane":        "backend-02",
							"env":              "intg",
							"envs":             ",intg,",
							"service":          "backend",
							"services":         ",backend,",
						},
					},
					"/meshes/demo/dataplanes/web-01": &observability_proto.MonitoringAssignment{
						Name: "/meshes/demo/dataplanes/web-01",
						Targets: []*observability_proto.MonitoringAssignment_Target{{
							Labels: map[string]string{
								"__address__": "192.168.0.3:8765",
							},
						}},
						Labels: map[string]string{
							"__scheme__":       "http",
							"__metrics_path__": "/even-more-non-standard-path",
							"job":              "web",
							"instance":         "web-01",
							"mesh":             "demo",
							"dataplane":        "web-01",
							"env":              "test",
							"envs":             ",test,",
							"service":          "web",
							"services":         ",web,",
						},
					},
				}),
			}),
		)
	})
})
