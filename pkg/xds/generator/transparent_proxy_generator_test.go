package generator_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("TransparentProxyGenerator", func() {

	type testCase struct {
		proxy    *model.Proxy
		expected string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// given
			gen := &generator.TransparentProxyGenerator{}
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{
							Logging: &mesh_proto.Logging{
								Backends: []*mesh_proto.LoggingBackend{
									{
										Name: "file",
										Type: mesh_proto.LoggingFileType,
										Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
											Path: "/var/log",
										}),
									},
								},
							},
						},
					},
				},
			}

			// when
			rs, err := gen.Generate(ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "transparent-proxy", given.expected)))
		},
		Entry("transparent_proxying=false", testCase{
			proxy: &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
				},
				APIVersion: envoy_common.APIV3,
			},
			expected: "01.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true", testCase{
			proxy: &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPortOutbound: 15001,
								RedirectPortInbound:  15006,
							},
						},
					},
				},
				APIVersion: envoy_common.APIV3,
				Policies: model.MatchedPolicies{
					TrafficLogs: map[model.ServiceName]*core_mesh.TrafficLogResource{ // to show that is not picked
						"some-service": {
							Spec: &mesh_proto.TrafficLog{
								Conf: &mesh_proto.TrafficLog_Conf{Backend: "file"},
							},
						},
					},
				},
			},
			expected: "02.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true with logs", testCase{
			proxy: &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPortOutbound: 15001,
								RedirectPortInbound:  15006,
							},
						},
					},
				},
				APIVersion: envoy_common.APIV3,
				Policies: model.MatchedPolicies{
					TrafficLogs: map[model.ServiceName]*core_mesh.TrafficLogResource{ // to show that is is not picked
						"pass_through": {
							Spec: &mesh_proto.TrafficLog{
								Conf: &mesh_proto.TrafficLog_Conf{Backend: "file"},
							},
						},
					},
				},
			},
			expected: "03.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true ipv6", testCase{
			proxy: &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPortOutbound:  15001,
								RedirectPortInbound:   15006,
								RedirectPortInboundV6: 15010,
							},
						},
					},
				},
				APIVersion: envoy_common.APIV3,
				Policies: model.MatchedPolicies{
					TrafficLogs: map[model.ServiceName]*core_mesh.TrafficLogResource{ // to show that is not picked
						"some-service": {
							Spec: &mesh_proto.TrafficLog{
								Conf: &mesh_proto.TrafficLog_Conf{Backend: "file"},
							},
						},
					},
				},
			},
			expected: "04.envoy.golden.yaml",
		}),
	)
})
