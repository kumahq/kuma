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
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("TracingProxyGenerator", func() {

	type testCase struct {
		ctx      xds_context.Context
		proxy    *core_xds.Proxy
		expected string
	}

	DescribeTable("should not generate Envoy xDS resources unless tracing is present",
		func(given testCase) {
			// setup
			gen := &generator.TracingProxyGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(rs).To(BeNil())
		},
		Entry("Mesh has no Tracing configuration", testCase{
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "demo.backend-01"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
				},
				APIVersion: envoy_common.APIV3,
			},
		}),
	)

	DescribeTable("should generate Envoy xDS resources if tracing backend is present",
		func(given testCase) {
			// given
			gen := &generator.TracingProxyGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "tracing", given.expected)))
		},
		Entry("should create cluster for Datadog", testCase{
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "demo.backend-01"),
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
				APIVersion: envoy_common.APIV3,
				Policies: core_xds.MatchedPolicies{
					TrafficTrace: &core_mesh.TrafficTraceResource{
						Spec: &mesh_proto.TrafficTrace{
							Conf: &mesh_proto.TrafficTrace_Conf{
								Backend: "datadog",
							},
						},
					},
				},
			},
			ctx: xds_context.Context{Mesh: xds_context.MeshContext{
				Resource: &core_mesh.MeshResource{Spec: &mesh_proto.Mesh{
					Tracing: &mesh_proto.Tracing{
						Backends: []*mesh_proto.TracingBackend{
							{
								Name: "datadog",
								Type: mesh_proto.TracingDatadogType,
								Conf: util_proto.MustToStruct(&mesh_proto.DatadogTracingBackendConfig{
									Address: "localhost",
									Port:    2304,
								}),
							},
						},
					},
				}},
			}},
			expected: "datadog.envoy-config.golden.yaml",
		}),
		Entry("should create cluster for Zipkin", testCase{
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "demo.backend-01"),
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
				APIVersion: envoy_common.APIV3,
				Policies: core_xds.MatchedPolicies{
					TrafficTrace: &core_mesh.TrafficTraceResource{
						Spec: &mesh_proto.TrafficTrace{
							Conf: &mesh_proto.TrafficTrace_Conf{
								Backend: "zipkin",
							},
						},
					},
				},
			},
			ctx: xds_context.Context{Mesh: xds_context.MeshContext{
				Resource: &core_mesh.MeshResource{Spec: &mesh_proto.Mesh{
					Tracing: &mesh_proto.Tracing{
						Backends: []*mesh_proto.TracingBackend{
							{
								Name: "zipkin",
								Type: mesh_proto.TracingZipkinType,
								Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
									Url: "http://zipkin.us:9090/v2/spans",
								}),
							},
						},
					},
				}},
			}},
			expected: "zipkin.envoy-config.golden.yaml",
		}),
		Entry("should create cluster for Zipkin with tls sni", testCase{
			proxy: &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("", "demo.https-backend-01"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "https-backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				APIVersion: envoy_common.APIV3,
				Policies: core_xds.MatchedPolicies{
					TrafficTrace: &core_mesh.TrafficTraceResource{
						Spec: &mesh_proto.TrafficTrace{
							Conf: &mesh_proto.TrafficTrace_Conf{
								Backend: "zipkin",
							},
						},
					},
				},
			},
			ctx: xds_context.Context{Mesh: xds_context.MeshContext{
				Resource: &core_mesh.MeshResource{Spec: &mesh_proto.Mesh{
					Tracing: &mesh_proto.Tracing{
						Backends: []*mesh_proto.TracingBackend{
							{
								Name: "zipkin",
								Type: mesh_proto.TracingZipkinType,
								Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
									Url: "https://zipkin.us:9090/v2/spans",
								}),
							},
						},
					},
				}},
			}},
			expected: "zipkin.envoy-config-https.golden.yaml",
		}),
	)
})
