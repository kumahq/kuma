package bootstrap

import (
	envoy_config_bootstrap_v2 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Bootstrap Tracing", func() {

	type testCase struct {
		backend      *mesh_proto.TracingBackend
		expectedYAML string
	}

	DescribeTable("should enrich bootstrap config with tracing",
		func(given testCase) {
			// given
			bootstrap := &envoy_config_bootstrap_v2.Bootstrap{}

			// when
			err := AddTracingConfig(bootstrap, given.backend)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			actual, err := util_proto.ToYAML(bootstrap)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(actual).To(MatchYAML(given.expectedYAML))
		},
		Entry("infer version from /api/v1/spans path", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-us",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url: "http://zipkin:9090/api/v1/spans",
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-us
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-us
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-us
                      collectorEndpoint: /api/v1/spans
`,
		}),
		Entry("infer version from /api/v2/spans path", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-eu",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url: "http://zipkin:9090/api/v2/spans",
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-eu
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-eu
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-eu
                      collectorEndpoint: /api/v2/spans
                      collectorEndpointVersion: HTTP_JSON
`,
		}),
		Entry("explicit httpJsonV1 version config in backend", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-eu",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url:        "http://zipkin:9090/api/v2/spans",
					ApiVersion: "httpJsonV1",
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-eu
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-eu
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-eu
                      collectorEndpoint: /api/v2/spans
`,
		}),
		Entry("explicit httpJson version config in backend", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-eu",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url:        "http://zipkin:9090/some/path",
					ApiVersion: "httpJson",
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-eu
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-eu
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-eu
                      collectorEndpoint: /some/path
                      collectorEndpointVersion: HTTP_JSON
`,
		}),
		Entry("explicit httpProto version config in backend", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-eu",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url:        "http://zipkin:9090/some/path",
					ApiVersion: "httpProto",
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-eu
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-eu
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-eu
                      collectorEndpoint: /some/path
                      collectorEndpointVersion: HTTP_PROTO
`,
		}),
		Entry("version defaults to httpJson", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-eu",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url: "http://zipkin:9090/some/path",
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-eu
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-eu
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-eu
                      collectorEndpoint: /some/path
                      collectorEndpointVersion: HTTP_JSON
`,
		}),
		Entry("traceId128bit on", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin-eu",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url:           "http://zipkin:9090/api/v2/spans",
					TraceId128Bit: true,
				}),
			},
			expectedYAML: `
                staticResources:
                  clusters:
                  - connectTimeout: 10s
                    loadAssignment:
                      clusterName: zipkin-eu
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: zipkin
                                portValue: 9090
                    name: zipkin-eu
                    type: STRICT_DNS
                tracing:
                  http:
                    name: envoy.zipkin
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
                      collectorCluster: zipkin-eu
                      collectorEndpoint: /api/v2/spans
                      collectorEndpointVersion: HTTP_JSON
                      traceId128bit: true
`,
		}),
	)
})
