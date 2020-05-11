package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/xds/generator"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	util_proto "github.com/Kong/kuma/pkg/util/proto"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
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
		Entry("both Mesh and Datalane have no Prometheus configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
					},
				},
			},
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
				},
			},
		}),
		Entry("Datalane has Prometheus configuration while Mesh doesn't", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
					},
				},
			},
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: mesh_proto.Dataplane{
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
		Entry("both Mesh and Datalane do have Prometheus configuration but dataplane metadata is unknown", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: mesh_proto.Mesh{
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
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: mesh_proto.Dataplane{
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
		Entry("both Mesh and Datalane do have Prometheus configuration but Admin API is not enabled on that dataplane", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: mesh_proto.Mesh{
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
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: mesh_proto.Dataplane{
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
			resp, err := model.ResourceList(rs).ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should support a Dataplane without custom metrics configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: mesh_proto.Mesh{
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
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort: 9902,
				},
			},
			expected: `
            resources:
            - name: kuma:envoy:admin
              resource:
                '@type': type.googleapis.com/envoy.api.v2.Cluster
                connectTimeout: 5s
                loadAssignment:
                  clusterName: kuma:envoy:admin
                  endpoints:
                  - lbEndpoints:
                    - endpoint:
                        address:
                          socketAddress:
                            address: 127.0.0.1
                            portValue: 9902
                name: kuma:envoy:admin
                altStatName: kuma_envoy_admin
                type: STATIC
            - name: kuma:metrics:prometheus
              resource:
                '@type': type.googleapis.com/envoy.api.v2.Listener
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 0.0.0.0
                    portValue: 1234
                filterChains:
                - filters:
                  - name: envoy.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                      httpFilters:
                      - name: envoy.router
                      routeConfig:
                        virtualHosts:
                        - domains:
                          - '*'
                          name: envoy_admin
                          routes:
                          - match:
                              prefix: /non-standard-path
                            route:
                              cluster: kuma:envoy:admin
                              prefixRewrite: /stats/prometheus
                      statPrefix: kuma_metrics_prometheus
                name: kuma:metrics:prometheus
`,
		}),
		Entry("should support a Dataplane with custom metrics configuration", testCase{
			ctx: xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: mesh_proto.Mesh{
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
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: mesh_proto.Dataplane{
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
				},
			},
			expected: `
            resources:
            - name: kuma:envoy:admin
              resource:
                '@type': type.googleapis.com/envoy.api.v2.Cluster
                connectTimeout: 5s
                loadAssignment:
                  clusterName: kuma:envoy:admin
                  endpoints:
                  - lbEndpoints:
                    - endpoint:
                        address:
                          socketAddress:
                            address: 127.0.0.1
                            portValue: 9902
                name: kuma:envoy:admin
                altStatName: kuma_envoy_admin
                type: STATIC
            - name: kuma:metrics:prometheus
              resource:
                '@type': type.googleapis.com/envoy.api.v2.Listener
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 0.0.0.0
                    portValue: 8765
                filterChains:
                - filters:
                  - name: envoy.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                      httpFilters:
                      - name: envoy.router
                      routeConfig:
                        virtualHosts:
                        - domains:
                          - '*'
                          name: envoy_admin
                          routes:
                          - match:
                              prefix: /even-more-non-standard-path
                            route:
                              cluster: kuma:envoy:admin
                              prefixRewrite: /stats/prometheus
                      statPrefix: kuma_metrics_prometheus
                name: kuma:metrics:prometheus
`,
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
						Resource: &mesh_core.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "demo",
							},
							Spec: mesh_proto.Mesh{
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
				proxy := &model.Proxy{
					Id: model.ProxyId{Name: "demo.backend-01"},
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name: "backend-01",
							Mesh: "demo",
						},
					},
					Metadata: &core_xds.DataplaneMetadata{
						AdminPort: 9902,
					},
				}
				Expect(util_proto.FromYAML([]byte(given.dataplane), &proxy.Dataplane.Spec)).To(Succeed())

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
                    port: 8080
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
                    service: db
                  - port: 59200
                    service: elastic
                metrics:
                  type: prometheus
                  conf:
                    port: 54321
`,
			}),
		)
	})
})
