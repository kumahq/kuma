package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/mads/generator"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/util/proto"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("MonitoringAssignmentsGenerator", func() {

	Describe("Generate()", func() {

		type testCase struct {
			meshes     []*mesh_core.MeshResource
			dataplanes []*mesh_core.DataplaneResource
			expected   []*core_xds.Resource
		}

		DescribeTable("should generate proper MonitoringAssignment resources",
			func(given testCase) {
				// setup
				generator := MonitoringAssignmentsGenerator{}
				// when
				resources, err := generator.Generate(Args{
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
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
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
				expected: []*core_xds.Resource{},
			}),
			Entry("Dataplane inside a Mesh without Prometheus enabled", testCase{
				meshes: []*mesh_core.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
							Mesh: "demo",
						},
					},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
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
				meshes: []*mesh_core.MeshResource{
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
							Name: "gateway-01",
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Gateway: &mesh_proto.Dataplane_Networking_Gateway{
									Tags: map[string]string{
										"service": "gateway",
										"region":  "eu",
									},
								},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name:    "/meshes/demo/dataplanes/gateway-01",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/demo/dataplanes/gateway-01",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": ":1234",
								},
							}},
							Labels: map[string]string{
								"__scheme__":       "http",
								"__metrics_path__": "/non-standard-path",
								"job":              "gateway",
								"instance":         "gateway-01",
								"mesh":             "demo",
								"dataplane":        "gateway-01",
								"region":           "eu",
								"regions":          ",eu,",
								"service":          "gateway",
								"services":         ",gateway,",
							},
						},
					},
				},
			}),
			Entry("Dataplane with multiple inbound interfaces", testCase{
				meshes: []*mesh_core.MeshResource{
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
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
										Tags: map[string]string{
											"service": "backend",
											"env":     "prod",
											"version": "v1",
										},
									},
									{
										Address:     "192.168.0.2",
										Port:        443,
										ServicePort: 8443,
										Tags: map[string]string{
											"service": "backend-https",
											"env":     "prod",
											"version": "v2",
										},
									},
								},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name:    "/meshes/demo/dataplanes/backend-01",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/demo/dataplanes/backend-01",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": "192.168.0.1:1234",
								},
							}},
							Labels: map[string]string{
								"__scheme__":       "http",
								"__metrics_path__": "/non-standard-path",
								"job":              "backend",
								"instance":         "backend-01",
								"mesh":             "demo",
								"dataplane":        "backend-01",
								"env":              "prod",
								"envs":             ",prod,",
								"service":          "backend",
								"services":         ",backend,backend-https,", // must have multiple values
								"version":          "v1",
								"versions":         ",v1,v2,", // must have multiple values
							},
						},
					},
				},
			}),
			Entry("Dataplane with a user-defined plural tag", testCase{
				meshes: []*mesh_core.MeshResource{
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
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"service":  "backend",
										"version":  "v1",
										"versions": "v1+v1.0.1",
									},
								}},
							},
						},
					},
				},
				expected: []*core_xds.Resource{
					{
						Name:    "/meshes/demo/dataplanes/backend-01",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/demo/dataplanes/backend-01",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": "192.168.0.1:1234",
								},
							}},
							Labels: map[string]string{
								"__scheme__":       "http",
								"__metrics_path__": "/non-standard-path",
								"job":              "backend",
								"instance":         "backend-01",
								"mesh":             "demo",
								"dataplane":        "backend-01",
								"service":          "backend",
								"services":         ",backend,",
								"version":          "v1",
								"versions":         "v1+v1.0.1", // must have user-defined value
								"versionss":        ",v1+v1.0.1,",
							},
						},
					},
				},
			}),
			Entry("Dataplane with unsafe characters in tag's name and value", testCase{
				meshes: []*mesh_core.MeshResource{
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
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        80,
									ServicePort: 8080,
									Tags: map[string]string{
										"service":         "backend",
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
						Name:    "/meshes/demo/dataplanes/backend-01",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/demo/dataplanes/backend-01",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": "192.168.0.1:1234",
								},
							}},
							Labels: map[string]string{
								"__scheme__":       "http",
								"__metrics_path__": "/non-standard-path",
								"job":              "backend",
								"instance":         "backend-01",
								"mesh":             "demo",
								"dataplane":        "backend-01",
								"service":          "backend",
								"services":         ",backend,",
								"app_description":  "?!,.:;",   // tag name must be escaped
								"app_descriptions": ",?!,.:;,", // tag name must be escaped
								"com_company_tag":  "&*()-+",   // tag name must be escaped
								"com_company_tags": ",&*()-+,", // tag name must be escaped
							},
						},
					},
				},
			}),
			Entry("multiple Meshes and Dataplanes", testCase{
				meshes: []*mesh_core.MeshResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "default",
							Mesh: "default",
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
											Port: 2345,
											Path: "/another-non-standard-path",
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
							Name: "web-02",
							Mesh: "demo",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
									Port:        443,
									ServicePort: 8443,
									Tags: map[string]string{
										"service": "web",
										"env":     "intg",
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
						Name:    "/meshes/default/dataplanes/backend-01",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/default/dataplanes/backend-01",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": "192.168.0.1:1234",
								},
							}},
							Labels: map[string]string{
								"__scheme__":       "http",
								"__metrics_path__": "/non-standard-path",
								"job":              "backend",
								"instance":         "backend-01",
								"mesh":             "default",
								"dataplane":        "backend-01",
								"env":              "prod",
								"envs":             ",prod,",
								"service":          "backend",
								"services":         ",backend,",
							},
						},
					},
					{
						Name:    "/meshes/demo/dataplanes/web-02",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/demo/dataplanes/web-02",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": "192.168.0.2:8765",
								},
							}},
							Labels: map[string]string{
								"__scheme__":       "http",
								"__metrics_path__": "/even-more-non-standard-path",
								"job":              "web",
								"instance":         "web-02",
								"mesh":             "demo",
								"dataplane":        "web-02",
								"env":              "intg",
								"envs":             ",intg,",
								"service":          "web",
								"services":         ",web,",
							},
						},
					},
				},
			}),
			Entry("should include `k8s_kuma_io_namespace` and `k8s_kuma_io_name` labels on Kubernetes", testCase{
				meshes: []*mesh_core.MeshResource{
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
							Name: "backend-5c89f4d995-85znn.my-namespace",
							NameExtensions: core_model.ResourceNameExtensions{
								"k8s.kuma.io/namespace": "my-namespace",
								"k8s.kuma.io/name":      "backend-5c89f4d995-85znn",
							},
							Mesh: "demo",
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
				expected: []*core_xds.Resource{
					{
						Name:    "/meshes/demo/dataplanes/backend-5c89f4d995-85znn.my-namespace",
						Version: "",
						Resource: &observability_proto.MonitoringAssignment{
							Name: "/meshes/demo/dataplanes/backend-5c89f4d995-85znn.my-namespace",
							Targets: []*observability_proto.MonitoringAssignment_Target{{
								Labels: map[string]string{
									"__address__": "192.168.0.1:1234",
								},
							}},
							Labels: map[string]string{
								"__scheme__":            "http",
								"__metrics_path__":      "/non-standard-path",
								"job":                   "backend",
								"instance":              "backend-5c89f4d995-85znn.my-namespace",
								"k8s_kuma_io_name":      "backend-5c89f4d995-85znn",
								"k8s_kuma_io_namespace": "my-namespace",
								"mesh":                  "demo",
								"dataplane":             "backend-5c89f4d995-85znn.my-namespace",
								"service":               "backend",
								"services":              ",backend,",
							},
						},
					},
				},
			}),
		)
	})
})
