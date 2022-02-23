package generator_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("PrometheusEndpointGenerator", func() {

	type testCase struct {
		ctx      xds_context.Context
		proxy    *core_xds.Proxy
		expected string
	}

	DescribeTable("should not generate Envoy xDS resources unless Prometheus metrics have been enabled Mesh-wide",
		func(given testCase) {
			// setup
			gen := &generator.PrometheusEndpointGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(rs).To(BeNil())
		},
		Entry("both Mesh and Dataplane have no Prometheus configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "demo.backend-01"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{},
				},
				APIVersion: envoy_common.APIV3,
			},
		}),
		Entry("Dataplane has Prometheus configuration while Mesh doesn't", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Metrics: &mesh_proto.MetricsBackend{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Port: 1234,
								Path: "/non-standard-path",
							}),
						},
					},
				},
			},
		}),
		Entry("both Mesh and Dataplane do have Prometheus configuration but dataplane metadata is unknown", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
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
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 1234,
											Path: "/non-standard-path",
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Metrics: &mesh_proto.MetricsBackend{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Port: 8765,
								Path: "/even-more-non-standard-path",
							}),
						},
					},
				},
				Metadata: nil, // dataplane metadata is unknown
			},
		}),
		Entry("both Mesh and Dataplane do have Prometheus configuration but Admin API is not enabled on that dataplane", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
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
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 1234,
											Path: "/non-standard-path",
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Metrics: &mesh_proto.MetricsBackend{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Port: 8765,
								Path: "/even-more-non-standard-path",
							}),
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{}, // dataplane was started without AdminPort
			},
		}),
	)

	DescribeTable("should generate Envoy xDS resources if Prometheus metrics have been enabled Mesh-wide",
		func(given testCase) {
			// setup
			gen := &generator.PrometheusEndpointGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "prometheus-endpoint", given.expected)))
		},
		Entry("should support a Dataplane without custom metrics configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
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
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port:     1234,
											Path:     "/non-standard-path",
											SkipMTLS: util_proto.Bool(false),
											Tags: map[string]string{
												"kuma.io/service": "dataplane-metrics",
											},
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
				},
			},
			expected: "default.envoy-config.golden.yaml",
		}),
		Entry("should support a Dataplane without metrics hijacker", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
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
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port:     1234,
											Path:     "/non-standard-path",
											SkipMTLS: util_proto.Bool(false),
											Tags: map[string]string{
												"kuma.io/service": "dataplane-metrics",
											},
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.1.6",
						},
					},
				},
			},
			expected: "default-without-hijacker.envoy-config.golden.yaml",
		}),
		Entry("should support a Dataplane with custom metrics configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
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
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 1234,
											Path: "/non-standard-path",
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
						Metrics: &mesh_proto.MetricsBackend{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Port: 8765,
								Path: "/even-more-non-standard-path",
							}),
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
				},
			},
			expected: "custom.envoy-config.golden.yaml",
		}),
		Entry("should support a Dataplane with mTLS on", testCase{
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
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
							Metrics: &mesh_proto.Metrics{
								EnabledBackend: "prometheus-1",
								Backends: []*mesh_proto.MetricsBackend{
									{
										Name: "prometheus-1",
										Type: mesh_proto.MetricsPrometheusType,
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port:     1234,
											Path:     "/non-standard-path",
											SkipMTLS: util_proto.Bool(false),
											Tags: map[string]string{
												"kuma.io/service": "dataplane-metrics",
											},
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
				},
			},
			expected: "default-mtls.envoy-config.golden.yaml",
		}),
		Entry("should support a Dataplane with mTLS on (skipMTLS not explicitly defined)", testCase{
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
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
							Metrics: &mesh_proto.Metrics{
								EnabledBackend: "prometheus-1",
								Backends: []*mesh_proto.MetricsBackend{
									{
										Name: "prometheus-1",
										Type: mesh_proto.MetricsPrometheusType,
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port: 1234,
											Path: "/non-standard-path",
											Tags: map[string]string{
												"kuma.io/service": "dataplane-metrics",
											},
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
				},
			},
			expected: "default-mtls.envoy-config.golden.yaml",
		}),
		Entry("should support a Dataplane with mTLS on but skipMTLS true", testCase{
			ctx: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
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
							Metrics: &mesh_proto.Metrics{
								EnabledBackend: "prometheus-1",
								Backends: []*mesh_proto.MetricsBackend{
									{
										Name: "prometheus-1",
										Type: mesh_proto.MetricsPrometheusType,
										Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
											Port:     1234,
											Path:     "/non-standard-path",
											SkipMTLS: util_proto.Bool(true),
											Tags: map[string]string{
												"kuma.io/service": "dataplane-metrics",
											},
										}),
									},
								},
							},
						},
					},
				},
			},
			proxy: &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
				},
			},
			expected: "default.envoy-config.golden.yaml",
		}),
	)

	Describe("should not generate Envoy xDS resources if Prometheus endpoint would otherwise overshadow a port that is already in use by the application or other Envoy listeners", func() {

		type testCase struct {
			dataplane string
		}

		DescribeTable("should not generate Envoy xDS resources if Prometheus endpoint would otherwise overshadow a port that is already in use by the application or other Envoy listeners",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					Mesh: xds_context.MeshContext{
						Resource: &core_mesh.MeshResource{
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
											Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
												Port: 1234,
												Path: "/non-standard-path",
											}),
										},
									},
								},
							},
						},
					},
				}
				proxy := &core_xds.Proxy{
					Id:         *core_xds.BuildProxyId("", "demo.backend-01"),
					APIVersion: envoy_common.APIV3,
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
						Spec: &mesh_proto.Dataplane{},
					},
					Metadata: &core_xds.DataplaneMetadata{
						AdminPort: 9902,
					},
				}
				Expect(util_proto.FromYAML([]byte(given.dataplane), proxy.Dataplane.Spec)).To(Succeed())

				// setup
				gen := &generator.PrometheusEndpointGenerator{}

				// when
				rs, err := gen.Generate(ctx, proxy)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(rs).To(BeNil())
			},
			Entry("should not overshadow inbound listener port", testCase{
				dataplane: `
                networking:
                  address: 192.168.0.1
                  inbound:
                  - port: 80
                    servicePort: 8080
                  outbound:
                  - port: 54321
                    service: db
                  - port: 59200
                    service: elastic
                metrics:
                  type: prometheus
                  conf:
                    port: 80
`,
			}),
			Entry("should not overshadow application port", testCase{
				dataplane: `
                networking:
                  address: 192.168.0.1
                  inbound:
                  - port: 80
                    servicePort: 8080
                  outbound:
                  - port: 54321
                    service: db
                  - port: 59200
                    service: elastic
                metrics:
                  type: prometheus
                  conf:
                    port: 80
`,
			}),
			Entry("should not overshadow outbound listener port", testCase{
				dataplane: `
                networking:
                  address: 192.168.0.1
                  inbound:
                  - port: 80
                    servicePort: 8080
                  outbound:
                  - port: 54321
                    address: 192.168.0.1
                    service: db
                metrics:
                  type: prometheus
                  conf:
                    port: 54321
`,
			}),
		)
	})
})
