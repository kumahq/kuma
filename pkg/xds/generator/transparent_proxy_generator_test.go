package generator_test

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/core/xds/types"
	. "github.com/kumahq/kuma/v2/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	"github.com/kumahq/kuma/v2/pkg/xds/generator"
)

var _ = Describe("TransparentProxyGenerator", func() {
	type testCase struct {
		proxy            *model.Proxy
		meshServicesMode mesh_proto.Mesh_MeshServices_Mode
		expected         string
		tlsMode          *mesh_proto.CertificateAuthorityBackend_Mode
	}

	strictInboundPortsProxy := func(inboundPorts []uint32) *model.Proxy {
		inbounds := make([]*mesh_proto.Dataplane_Networking_Inbound, len(inboundPorts))
		for i, port := range inboundPorts {
			inbounds[i] = &mesh_proto.Dataplane_Networking_Inbound{Port: port}
		}

		return &model.Proxy{
			Metadata: &model.DataplaneMetadata{Features: map[string]bool{
				types.FeatureUnifiedResourceNaming: true,
				types.FeatureStrictInboundPorts:    true,
			}},
			Id: *model.BuildProxyId("", "side-car"),
			Dataplane: &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{Version: "v1"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: inbounds,
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
				TrafficLogs: map[model.ServiceName]*core_mesh.TrafficLogResource{
					"some-service": {Spec: &mesh_proto.TrafficLog{Conf: &mesh_proto.TrafficLog_Conf{Backend: "file"}}},
				},
			},
			InternalAddresses: DummyInternalAddresses,
		}
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// given
			gen := &generator.TransparentProxyGenerator{}
			var mtls *mesh_proto.Mesh_Mtls
			if given.tlsMode != nil {
				mtls = &mesh_proto.Mesh_Mtls{
					EnabledBackend: "backend",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{
							Name: "backend",
							Type: "builtin",
							Mode: *given.tlsMode,
						},
					},
				}
			}
			xdsCtx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{
							MeshServices: &mesh_proto.Mesh_MeshServices{
								Mode: given.meshServicesMode,
							},
							Mtls: mtls,
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
		Entry("transparent_proxying=true,unified_naming=true", testCase{
			proxy: &model.Proxy{
				Metadata: &model.DataplaneMetadata{Features: map[string]bool{types.FeatureUnifiedResourceNaming: true}},
				Id:       *model.BuildProxyId("", "side-car"),
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
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			expected:         "05.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,unified_naming=true,inbound_filter,strict", testCase{
			proxy:            strictInboundPortsProxy([]uint32{8080}),
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			tlsMode:          mesh_proto.CertificateAuthorityBackend_STRICT.Enum(),
			expected:         "06.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,unified_naming=true,inbound_filter,permissive", testCase{
			proxy:            strictInboundPortsProxy([]uint32{8080, 9000}),
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			tlsMode:          mesh_proto.CertificateAuthorityBackend_PERMISSIVE.Enum(),
			expected:         "07.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,unified_naming=true,inbound_filter,no tls", testCase{
			proxy:            strictInboundPortsProxy([]uint32{8080}),
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			tlsMode:          mesh_proto.CertificateAuthorityBackend_PERMISSIVE.Enum(),
			expected:         "08.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,unified_naming=true,inbound_filter,strict,gateway", testCase{
			proxy: &model.Proxy{
				Metadata: &model.DataplaneMetadata{Features: map[string]bool{
					types.FeatureUnifiedResourceNaming: true,
					types.FeatureStrictInboundPorts:    true,
				}},
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{
								Tags: map[string]string{
									"app": "test-gateway",
								},
								Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
							},
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								IpFamilyMode:         mesh_proto.Dataplane_Networking_TransparentProxying_DualStack,
								RedirectPortOutbound: 15001,
								RedirectPortInbound:  15006,
							},
						},
					},
				},
				APIVersion:        envoy_common.APIV3,
				Policies:          model.MatchedPolicies{},
				InternalAddresses: DummyInternalAddresses,
			},
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			tlsMode:          mesh_proto.CertificateAuthorityBackend_STRICT.Enum(),
			expected:         "09.envoy.golden.yaml",
		}),
	)
})
