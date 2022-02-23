package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads/generator"
	. "github.com/kumahq/kuma/pkg/mads/v1/generator"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("MonitoringAssignmentsGenerator", func() {

	Describe("Generate()", func() {

		type testCase struct {
			meshes     []*core_mesh.MeshResource
			dataplanes []*core_mesh.DataplaneResource
			expected   []*core_xds.Resource
		}

		DescribeTable("should generate proper MonitoringAssignment resources",
			func(given testCase) {
				// setup
				gen := MonitoringAssignmentsGenerator{}
				// when
				resources, err := gen.Generate(generator.Args{
					Meshes:     given.meshes,
					Dataplanes: given.dataplanes,
				})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(resources).To(Equal(given.expected))
			},
			Entry("no Meshes, no Dataplanes", testCase{
				expected: []*core_xds.Resource{},
			}),
			Entry("Dataplane without Mesh", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								}},
							},
						},
					},
				},
				expected: []*core_xds.Resource{},
			}),
			Entry("Dataplane inside a Mesh without Prometheus enabled", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								}},
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
				expected: []*core_xds.Resource{},
			}),
			Entry("Dataplane without inbound interfaces", testCase{
				meshes: []*core_mesh.MeshResource{
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
					{
						Meta: &test_model.ResourceMeta{
							Name: "gateway-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Gateway: &mesh_proto.Dataplane_Networking_Gateway{
									Tags: map[string]string{
										"kuma.io/service": "gateway",
										"region":          "eu",
									},
								},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name: "/meshes/demo/dataplanes/gateway-01",
						Resource: &observability_v1.MonitoringAssignment{
							Service: "gateway",
							Mesh:    "demo",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "gateway-01",
								Address:     ":1234",
								Scheme:      "http",
								MetricsPath: "/non-standard-path",
								Labels: map[string]string{
									"region":           "eu",
									"regions":          ",eu,",
									"kuma_io_service":  "gateway",
									"kuma_io_services": ",gateway,",
								},
							}},
						},
					},
				},
			}),
			Entry("Dataplane with multiple inbound interfaces", testCase{
				meshes: []*core_mesh.MeshResource{
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
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
										Tags: map[string]string{
											"kuma.io/service": "backend",
											"env":             "prod",
											"version":         "v1",
										},
									},
									{
										Address:     "192.168.0.2",
										Port:        443,
										ServicePort: 8443,
										Tags: map[string]string{
											"kuma.io/service": "backend-https",
											"env":             "prod",
											"version":         "v2",
										},
									},
								},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name: "/meshes/demo/dataplanes/backend-01",
						Resource: &observability_v1.MonitoringAssignment{
							Mesh:    "demo",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								Scheme:      "http",
								MetricsPath: "/non-standard-path",
								Labels: map[string]string{
									"env":              "prod",
									"envs":             ",prod,",
									"kuma_io_service":  "backend",
									"kuma_io_services": ",backend,backend-https,", // must have multiple values
									"version":          "v1",
									"versions":         ",v1,v2,", // must have multiple values
								},
							}},
						},
					},
				},
			}),
			Entry("Dataplane with a user-defined plural tag", testCase{
				meshes: []*core_mesh.MeshResource{
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
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"version":         "v1",
										"versions":        "v1+v1.0.1",
									},
								}},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name: "/meshes/demo/dataplanes/backend-01",
						Resource: &observability_v1.MonitoringAssignment{
							Mesh:    "demo",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Scheme:      "http",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/non-standard-path",
								Labels: map[string]string{
									"kuma_io_service":  "backend",
									"kuma_io_services": ",backend,",
									"version":          "v1",
									"versions":         "v1+v1.0.1", // must have user-defined value
									"versionss":        ",v1+v1.0.1,",
								},
							}},
						},
					},
				},
			}),
			Entry("Dataplane with unsafe characters in tag's name and value", testCase{
				meshes: []*core_mesh.MeshResource{
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
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"app:description": "?!,.:;",
										"com.company/tag": "&*()-+",
									},
								}},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name: "/meshes/demo/dataplanes/backend-01",
						Resource: &observability_v1.MonitoringAssignment{
							Mesh:    "demo",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Scheme:      "http",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/non-standard-path",
								Labels: map[string]string{
									"kuma_io_service":  "backend",
									"kuma_io_services": ",backend,",
									"app_description":  "?!,.:;",   // tag name must be escaped
									"app_descriptions": ",?!,.:;,", // tag name must be escaped
									"com_company_tag":  "&*()-+",   // tag name must be escaped
									"com_company_tags": ",&*()-+,", // tag name must be escaped
								},
							}},
						},
					},
				},
			}),
			Entry("multiple Meshes and Dataplanes", testCase{
				meshes: []*core_mesh.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
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
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Mesh{
							Metrics: &mesh_proto.Metrics{
								EnabledBackend: "prometheus-1",
								Backends: []*mesh_proto.MetricsBackend{
									{
										Name: "prometheus-1",
										Type: mesh_proto.MetricsPrometheusType,
										Conf: proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 2345,
											Path: "/another-non-standard-path",
										}),
									},
								},
							},
						},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "default",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"kuma.io/service": "backend",
										"env":             "prod",
									},
								}},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "web-02",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        443,
									ServicePort: 8443,
									Tags: map[string]string{
										"kuma.io/service": "web",
										"env":             "intg",
									},
								}},
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
				expected: []*core_xds.Resource{
					{
						Name: "/meshes/default/dataplanes/backend-01",
						Resource: &observability_v1.MonitoringAssignment{
							Service: "backend",
							Mesh:    "default",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-01",
								Address:     "192.168.0.1:1234",
								Scheme:      "http",
								MetricsPath: "/non-standard-path",
								Labels: map[string]string{
									"env":              "prod",
									"envs":             ",prod,",
									"kuma_io_service":  "backend",
									"kuma_io_services": ",backend,",
								},
							}},
						},
					},
					{
						Name: "/meshes/demo/dataplanes/web-02",
						Resource: &observability_v1.MonitoringAssignment{
							Mesh:    "demo",
							Service: "web",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "web-02",
								Address:     "192.168.0.2:8765",
								Scheme:      "http",
								MetricsPath: "/even-more-non-standard-path",
								Labels: map[string]string{
									"env":              "intg",
									"envs":             ",intg,",
									"kuma_io_service":  "web",
									"kuma_io_services": ",web,",
								},
							}},
						},
					},
				},
			}),
			Entry("should include `k8s_kuma_io_namespace` and `k8s_kuma_io_name` labels on Kubernetes", testCase{
				meshes: []*core_mesh.MeshResource{
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
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-5c89f4d995-85znn.my-namespace",
							NameExtensions: core_model.ResourceNameExtensions{
								"k8s.kuma.io/namespace": "my-namespace",
								"k8s.kuma.io/name":      "backend-5c89f4d995-85znn",
							},
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								}},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name: "/meshes/demo/dataplanes/backend-5c89f4d995-85znn.my-namespace",
						Resource: &observability_v1.MonitoringAssignment{
							Mesh:    "demo",
							Service: "backend",
							Targets: []*observability_v1.MonitoringAssignment_Target{{
								Name:        "backend-5c89f4d995-85znn.my-namespace",
								Scheme:      "http",
								Address:     "192.168.0.1:1234",
								MetricsPath: "/non-standard-path",
								Labels: map[string]string{
									"k8s_kuma_io_name":      "backend-5c89f4d995-85znn",
									"k8s_kuma_io_namespace": "my-namespace",
									"kuma_io_service":       "backend",
									"kuma_io_services":      ",backend,",
								},
							}},
						},
					},
				},
			}),
		)
	})
})
