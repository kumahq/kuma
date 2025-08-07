package generator_test

import (
	"context"
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
			xdsCtx := xds_context.Context{
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
			rs, err := gen.Generate(context.Background(), nil, xdsCtx, given.proxy)

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
								IpFamilyMode:         mesh_proto.Dataplane_Networking_TransparentProxying_DualStack,
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
				InternalAddresses: DummyInternalAddresses,
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
								IpFamilyMode:         mesh_proto.Dataplane_Networking_TransparentProxying_DualStack,
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
		Entry("transparent_proxying=true ipv6 default port", testCase{
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
		Entry("transparent_proxying=true ipv6 port customized", testCase{
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
								RedirectPortInboundV6: 15066,
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
			expected: "05.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true ipv6 disabled", testCase{
			proxy: &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								IpFamilyMode:         mesh_proto.Dataplane_Networking_TransparentProxying_IPv4,
								RedirectPortOutbound: 15001,
								RedirectPortInbound:  15006,
								// this value here is actually invalid, it should be always 0 when IpFamilyMode is ipv4
								// we will assert this value should be ignored even it's set in this case
								RedirectPortInboundV6: 15066,
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
			expected: "06.envoy.golden.yaml",
		}),
	)

	Describe("TransparentProxyGenerator GetIPv6InboundRedirectPort", func() {
		It("should use ipv4 redirect port for ipv6", func() {
			p := createDataplaneProxy(0, mesh_proto.Dataplane_Networking_TransparentProxying_DualStack)

			ipv6RedirectPort := generator.GetIPv6InboundRedirectPort(p)

			Expect(ipv6RedirectPort).To(Equal(uint32(15006)))
		})

		It("should use user customized ipv6 redirect port when ipv6 not disabled", func() {
			p := createDataplaneProxy(15088, mesh_proto.Dataplane_Networking_TransparentProxying_DualStack)

			ipv6RedirectPort := generator.GetIPv6InboundRedirectPort(p)

			Expect(ipv6RedirectPort).To(Equal(uint32(15088)))
		})

		It("should get ipv6 redirect port as 0 when ipv6 disabled", func() {
			p := createDataplaneProxy(15088, mesh_proto.Dataplane_Networking_TransparentProxying_IPv4)

			ipv6RedirectPort := generator.GetIPv6InboundRedirectPort(p)

			Expect(ipv6RedirectPort).To(Equal(uint32(0)))
		})

		It("should get ipv6 redirect port as 0 when ipv6 disabled with unspecified ipFamilyMode", func() {
			p := createDataplaneProxy(0, mesh_proto.Dataplane_Networking_TransparentProxying_UnSpecified)

			ipv6RedirectPort := generator.GetIPv6InboundRedirectPort(p)

			Expect(ipv6RedirectPort).To(Equal(uint32(0)))
		})
	})
})

func createDataplaneProxy(ipv6InboundRedirectPort uint32, ipFamilyMode mesh_proto.Dataplane_Networking_TransparentProxying_IpFamilyMode) *model.Proxy {
	return &model.Proxy{
		Id: *model.BuildProxyId("", "sidecar"),
		Dataplane: &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Version: "v1",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
						RedirectPortOutbound:  15001,
						RedirectPortInbound:   15006,
						RedirectPortInboundV6: ipv6InboundRedirectPort,
						IpFamilyMode:          ipFamilyMode,
					},
				},
			},
		},
		APIVersion: envoy_common.APIV3,
	}
}
