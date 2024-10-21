package reconcile_test

import (
	"context"
	"net"
	"time"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
	mads_cache "github.com/kumahq/kuma/pkg/mads/v1/cache"
	mads_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	meshmetrics_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	. "github.com/kumahq/kuma/pkg/mads/v1/reconcile"
	"github.com/kumahq/kuma/pkg/metrics"
	// to match custom policy resource type like you need to register them manually in tests
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/util/proto"
	v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
)

var _ = Describe("snapshotGenerator", func() {
	Describe("GenerateSnapshot()", func() {
		format.MaxLength = 0
		var resourceManager core_manager.ResourceManager
		var store core_store.ResourceStore
		node1Id := "one"
		snapshotWithTwoAssignments := mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
			"/meshes/demo/dataplanes/backend-02": &observability_v1.MonitoringAssignment{
				Mesh:    "demo",
				Service: "backend",
				Targets: []*observability_v1.MonitoringAssignment_Target{{
					Name:        "backend-02",
					Address:     "192.168.0.2:1234",
					Scheme:      "http",
					MetricsPath: "/metrics",
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
		})

		meshMetricSnapshot := mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
			"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
				Mesh:    "default",
				Service: "backend",
				Targets: []*observability_v1.MonitoringAssignment_Target{{
					Name:        "backend-01",
					Address:     "192.168.0.1:1234",
					MetricsPath: "/custom",
					Scheme:      "http",
					Labels:      map[string]string{},
				}},
			},
		})

		BeforeEach(func() {
			store = memory.NewStore()
			resourceManager = core_manager.NewResourceManager(store)
		})

		type testCase struct {
			meshes            []*core_mesh.MeshResource
			meshMetrics       []*v1alpha1.MeshMetricResource
			dataplanes        []*core_mesh.DataplaneResource
			expectedSnapshots map[string]v3.Snapshot
		}

		DescribeTable("",
			func(given testCase) {
				// setup
				zone := ""
				cacheExpirationTime := time.Millisecond
				meshContextBuilder := xds_context.NewMeshContextBuilder(
					resourceManager,
					server.MeshResourceTypes(),
					net.LookupIP,
					zone,
					vips.NewPersistence(resourceManager, config_manager.NewConfigManager(store), false),
					".mesh",
					80,
					xds_context.AnyToAnyReachableServicesGraphBuilder,
				)
				newMetrics, err := metrics.NewMetrics(zone)
				Expect(err).ToNot(HaveOccurred())
				cache, err := mesh.NewCache(cacheExpirationTime, meshContextBuilder, newMetrics)
				Expect(err).ToNot(HaveOccurred())

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
				for _, meshMetric := range given.meshMetrics {
					// when
					err := resourceManager.Create(ctx, meshMetric, core_store.CreateBy(core_model.MetaToResourceKey(meshMetric.GetMeta())))
					// then
					Expect(err).ToNot(HaveOccurred())
				}

				// given
				snapshotter := NewSnapshotGenerator(resourceManager, mads_generator.MonitoringAssignmentsGenerator{}, cache)
				// when
				snapshotPerClient, err := snapshotter.GenerateSnapshot(context.Background())
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				for c := range snapshotPerClient {
					// Cleanup the versions as they are UUIDs
					snapshotPerClient[c] = snapshotPerClient[c].WithVersion(mads_v1.MonitoringAssignmentType, "")
				}
				Expect(snapshotPerClient).To(Equal(given.expectedSnapshots))
			},
			Entry("no Meshes, no Dataplanes, no MeshMetrics", testCase{
				expectedSnapshots: map[string]v3.Snapshot{
					meshmetrics_generator.DefaultKumaClientId: mads_cache.NewSnapshot("", nil),
				},
			}),
			Entry("no Meshes with Prometheus enabled, no MeshMetrics", testCase{
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
				expectedSnapshots: map[string]v3.Snapshot{
					meshmetrics_generator.DefaultKumaClientId: mads_cache.NewSnapshot("", nil),
				},
			}),
			Entry("Mesh with Prometheus enabled but no Dataplanes, no MeshMetrics", testCase{
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
				expectedSnapshots: map[string]v3.Snapshot{
					meshmetrics_generator.DefaultKumaClientId: mads_cache.NewSnapshot("", nil),
				},
			}),
			Entry("Mesh with Prometheus enabled and some Dataplanes, no MeshMetrics", testCase{
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
				expectedSnapshots: map[string]v3.Snapshot{
					meshmetrics_generator.DefaultKumaClientId: snapshotWithTwoAssignments,
				},
			}),
			Entry("no Meshes with Prometheus enabled, MeshMetric with Prometheus enabled for all nodes", testCase{
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
						WithInboundOfTags(mesh_proto.ServiceTag, "backend", "env", "intg").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.Mesh,
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											// ClientId: no client id passed, should configure all nodes the same
											Port: 1234,
											Path: "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
								},
							},
						},
					},
				},
				expectedSnapshots: map[string]v3.Snapshot{
					meshmetrics_generator.DefaultKumaClientId: meshMetricSnapshot,
				},
			}),
			Entry("no Meshes with Prometheus enabled, MeshMetric with Prometheus enabled for node 1", testCase{
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
						WithInboundOfTags(mesh_proto.ServiceTag, "backend", "env", "intg").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.Mesh,
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port:     1234,
											Path:     "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
								},
							},
						},
					},
				},
				expectedSnapshots: map[string]v3.Snapshot{
					node1Id: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
					}),
				},
			}),
			Entry("no Meshes with Prometheus enabled, MeshMetric with Prometheus enabled for node 1 and default client for rest", testCase{
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
						WithInboundOfTags(mesh_proto.ServiceTag, "backend", "env", "intg").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.Mesh,
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port:     1234,
											Path:     "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											Port: 1234,
											Path: "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
								},
							},
						},
					},
				},
				expectedSnapshots: map[string]v3.Snapshot{
					node1Id: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
					}),
					meshmetrics_generator.DefaultKumaClientId: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
					}),
				},
			}),
			Entry("no Meshes with Prometheus enabled, MeshMetric with Prometheus for some dataplanes", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					builders.Dataplane().
						WithName("backend-01").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-01").
						WithServices("backend-01").
						WithAddress("192.168.0.1").
						Build(),
					builders.Dataplane().
						WithName("backend-02").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-02").
						WithAddress("192.168.0.2").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.MeshService,
								Name: "backend-01",
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port:     1234,
											Path:     "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
								},
							},
						},
					},
				},
				expectedSnapshots: map[string]v3.Snapshot{
					node1Id: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-01",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
					}),
				},
			}),
			Entry("no Meshes with Prometheus enabled, multiple MeshMetrics with Prometheus and merging", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					builders.Dataplane().
						WithName("backend-01").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-01").
						WithServices("backend-01").
						WithAddress("192.168.0.1").
						Build(),
					builders.Dataplane().
						WithName("backend-02").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-02").
						WithAddress("192.168.0.2").
						Build(),
					builders.Dataplane().
						WithName("backend-03").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-03").
						WithAddress("192.168.0.3").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.Mesh,
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port:     1234,
											Path:     "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "override",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.MeshService,
								Name: "backend-02",
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port:     5678,
											Path:     "/other",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.ProvidedTLS,
											},
										},
									},
								},
							},
						},
					},
				},
				expectedSnapshots: map[string]v3.Snapshot{
					node1Id: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-01",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
						"/meshes/default/dataplanes/backend-02": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-02",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-02",
								Address:     "192.168.0.2:5678",
								MetricsPath: "/other",
								Scheme:      "https",
								Labels:      map[string]string{},
							}},
						},
						"/meshes/default/dataplanes/backend-03": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-03",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-03",
								Address:     "192.168.0.3:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
					}),
				},
			}),
			Entry("no Meshes with Prometheus enabled, MeshMetric targeting Mesh with override for backend-02 with Prometheus metrics backend", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					builders.Dataplane().
						WithName("backend-01").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-01").
						WithServices("backend-01").
						WithAddress("192.168.0.1").
						Build(),
					builders.Dataplane().
						WithName("backend-02").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-02").
						WithAddress("192.168.0.2").
						Build(),
					builders.Dataplane().
						WithName("backend-03").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend-03").
						WithAddress("192.168.0.3").
						Build(),
				},
				meshMetrics: []*v1alpha1.MeshMetricResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.Mesh,
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											ClientId: &node1Id,
											Port:     1234,
											Path:     "/custom",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.Disabled,
											},
										},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "override",
							Mesh: "default",
						},
						Spec: &v1alpha1.MeshMetric{
							TargetRef: &common_api.TargetRef{
								Kind: common_api.MeshService,
								Name: "backend-02",
							},
							Default: v1alpha1.Conf{
								Backends: &[]v1alpha1.Backend{
									{
										Type: v1alpha1.PrometheusBackendType,
										Prometheus: &v1alpha1.PrometheusBackend{
											Port: 5678,
											Path: "/other",
											Tls: &v1alpha1.PrometheusTls{
												Mode: v1alpha1.ProvidedTLS,
											},
										},
									},
								},
							},
						},
					},
				},
				expectedSnapshots: map[string]v3.Snapshot{
					meshmetrics_generator.DefaultKumaClientId: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-02": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-02",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-02",
								Address:     "192.168.0.2:5678",
								MetricsPath: "/other",
								Scheme:      "https",
								Labels:      map[string]string{},
							}},
						},
					}),
					node1Id: mads_cache.NewSnapshot("", map[string]envoy_types.Resource{
						"/meshes/default/dataplanes/backend-01": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-01",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
						"/meshes/default/dataplanes/backend-03": &observability_v1.MonitoringAssignment{
							Mesh:    "default",
							Service: "backend-03",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-03",
								Address:     "192.168.0.3:1234",
								MetricsPath: "/custom",
								Scheme:      "http",
								Labels:      map[string]string{},
							}},
						},
					}),
				},
			}),
		)
	})
})
