package mesh_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("MeshResource", func() {

	Describe("HasPrometheusMetricsEnabled", func() {

		type testCase struct {
			mesh     *MeshResource
			expected bool
		}

		DescribeTable("should correctly determine whether Prometheus metrics has been enabled on that Mesh",
			func(given testCase) {
				Expect(given.mesh.HasPrometheusMetricsEnabled()).To(Equal(given.expected))
			},
			Entry("mesh == nil", testCase{
				mesh:     nil,
				expected: false,
			}),
			Entry("mesh.metrics == nil", testCase{
				mesh:     NewMeshResource(),
				expected: false,
			}),
			Entry("mesh.metrics.prometheus == nil", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Metrics: &mesh_proto.Metrics{},
					},
				},
				expected: false,
			}),
			Entry("mesh.metrics.prometheus != nil", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Metrics: &mesh_proto.Metrics{
							EnabledBackend: "prometheus-1",
							Backends: []*mesh_proto.MetricsBackend{
								{
									Name: "prometheus-1",
									Type: mesh_proto.MetricsPrometheusType,
								},
							},
						},
					},
				},
				expected: true,
			}),
		)
	})

	Describe("GetTracingBackend", func() {

		type testCase struct {
			mesh     *MeshResource
			name     string
			expected string
		}

		DescribeTable("should return tracing backend",
			func(given testCase) {
				// when
				backend := given.mesh.GetTracingBackend(given.name)

				// then
				if given.expected == "" {
					Expect(backend).To(BeNil())
				} else {
					Expect(backend.Name).To(Equal(given.expected))
				}
			},
			Entry("two backends and name that exists", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Tracing: &mesh_proto.Tracing{
							DefaultBackend: "zipkin-us",
							Backends: []*mesh_proto.TracingBackend{
								{
									Name: "zipkin-us",
									Type: mesh_proto.TracingZipkinType,
									Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
										Url: "http://zipkin.us:8080/v1/spans",
									}),
								},
								{
									Name: "zipkin-eu",
									Type: mesh_proto.TracingZipkinType,
									Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
										Url: "http://zipkin.eu:8080/v1/spans",
									}),
								},
							},
						},
					},
				},
				name:     "zipkin-eu",
				expected: "zipkin-eu",
			}),
			Entry("nil when backend does not exist", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Tracing: &mesh_proto.Tracing{
							DefaultBackend: "zipkin-us",
							Backends: []*mesh_proto.TracingBackend{
								{
									Name: "zipkin-us",
									Type: mesh_proto.TracingZipkinType,
									Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
										Url: "http://zipkin.us:8080/v1/spans",
									}),
								},
							},
						},
					},
				},
				name:     "non-existing-backend",
				expected: "",
			}),
			Entry("default backend when name is not specified", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Tracing: &mesh_proto.Tracing{
							DefaultBackend: "zipkin-eu",
							Backends: []*mesh_proto.TracingBackend{
								{
									Name: "zipkin-us",
									Type: mesh_proto.TracingZipkinType,
									Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
										Url: "http://zipkin.us:8080/v1/spans",
									}),
								},
								{
									Name: "zipkin-eu",
									Type: mesh_proto.TracingZipkinType,
									Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
										Url: "http://zipkin.eu:8080/v1/spans",
									}),
								},
							},
						},
					},
				},
				name:     "",
				expected: "zipkin-eu",
			}),
			Entry("nil when name and default backend are not specified", testCase{
				mesh: &MeshResource{
					Spec: &mesh_proto.Mesh{
						Tracing: &mesh_proto.Tracing{
							Backends: []*mesh_proto.TracingBackend{
								{
									Name: "zipkin-us",
									Type: mesh_proto.TracingZipkinType,
									Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
										Url: "http://zipkin.us:8080/v1/spans",
									}),
								},
							},
						},
					},
				},
				name:     "",
				expected: "",
			}),
		)
	})

	Describe("should return logging backends", func() {
		It("should return logging backends if not empty", func() {
			mesh := &MeshResource{
				Spec: &mesh_proto.Mesh{
					Logging: &mesh_proto.Logging{
						Backends: []*mesh_proto.LoggingBackend{
							{
								Name: "logstash-1",
								Type: "logstash",
							},
							{
								Name: "file-1",
								Type: "file",
							},
						},
					},
				},
			}
			backends := mesh.GetLoggingBackends()
			Expect(backends).To(Equal("logstash/logstash-1, file/file-1"))
		})
		It("should return default logging backend if logging backends is empty", func() {
			mesh := &MeshResource{
				Spec: &mesh_proto.Mesh{
					Logging: &mesh_proto.Logging{
						DefaultBackend: "default-backend",
						Backends:       []*mesh_proto.LoggingBackend{},
					},
				},
			}
			backends := mesh.GetLoggingBackends()
			Expect(backends).To(Equal(""))
		})
	})

	Describe("should return tracing backends", func() {
		It("should return tracing backends if not empty", func() {
			mesh := &MeshResource{
				Spec: &mesh_proto.Mesh{
					Tracing: &mesh_proto.Tracing{
						Backends: []*mesh_proto.TracingBackend{
							{
								Name: "zipkin-us",
								Type: mesh_proto.TracingZipkinType,
								Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
									Url: "http://zipkin.us:8080/v1/spans",
								}),
							},
							{
								Name: "zipkin-eu",
								Type: mesh_proto.TracingZipkinType,
								Conf: proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
									Url: "http://zipkin.eu:8080/v1/spans",
								}),
							},
						},
					},
				},
			}

			backends := mesh.GetTracingBackends()
			Expect(backends).To(Equal("zipkin/zipkin-us, zipkin/zipkin-eu"))
		})
		It("should return default tracing backend if tracing backends is empty", func() {
			mesh := &MeshResource{
				Spec: &mesh_proto.Mesh{
					Tracing: &mesh_proto.Tracing{
						Backends: []*mesh_proto.TracingBackend{},
					},
				},
			}
			backends := mesh.GetTracingBackends()
			Expect(backends).To(Equal(""))
		})
	})
	Describe("ParseDuration", func() {

		type testCase struct {
			input  string
			output time.Duration
		}

		DescribeTable("should return the correct duration",
			func(given testCase) {
				duration, _ := ParseDuration(given.input)
				Expect(given.output).To(Equal(duration))
			},
			Entry("should return 0 if seconds is 0", testCase{
				input:  "0s",
				output: 0,
			}),
			Entry("should return minute", testCase{
				input:  "5m",
				output: 5 * time.Minute,
			}),
			Entry("should return day", testCase{
				input:  "4d",
				output: 4 * 24 * time.Hour,
			}),
			Entry("should return year", testCase{
				input:  "5y",
				output: 5 * 365 * 24 * time.Hour,
			}),
		)
	})
})
