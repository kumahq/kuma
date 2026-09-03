package generator_test

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/types"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("TransparentProxyGenerator", func() {
	type testCase struct {
		proxy    *model.Proxy
		expected string
	}

	withWorkloadIdentity := func(proxy *model.Proxy) *model.Proxy {
		proxy.WorkloadIdentity = &model.WorkloadIdentity{}
		return proxy
	}

	strictInboundPortsProxy := func(inboundPorts []uint32) *model.Proxy {
		inbounds := make([]*mesh_proto.Dataplane_Networking_Inbound, len(inboundPorts))
		for i, port := range inboundPorts {
			inbounds[i] = &mesh_proto.Dataplane_Networking_Inbound{Port: port}
		}

		return &model.Proxy{
			Metadata: &model.DataplaneMetadata{Features: map[string]bool{
				types.FeatureStrictInboundPorts: true,
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
			APIVersion:        envoy_common.APIV3,
			InternalAddresses: DummyInternalAddresses,
		}
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
						Spec: &mesh_proto.Mesh{},
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
				APIVersion:        envoy_common.APIV3,
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
			},
			expected: "04.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,inbound_filter,workload identity", testCase{
			proxy:    withWorkloadIdentity(strictInboundPortsProxy([]uint32{8080})),
			expected: "06.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,inbound_filter,no identity", testCase{
			proxy:    strictInboundPortsProxy([]uint32{8080}),
			expected: "08.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,inbound_filter,no identity,multiple ports", testCase{
			proxy:    strictInboundPortsProxy([]uint32{8080, 9000}),
			expected: "07.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,inbound_filter,workload identity,duplicate_ports", testCase{
			proxy:    withWorkloadIdentity(strictInboundPortsProxy([]uint32{8080, 8080})),
			expected: "10.envoy.golden.yaml",
		}),
		Entry("transparent_proxying=true,inbound_filter,workload identity,gateway", testCase{
			proxy: &model.Proxy{
				Metadata: &model.DataplaneMetadata{Features: map[string]bool{
					types.FeatureStrictInboundPorts: true,
				}},
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{},
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
				WorkloadIdentity:  &model.WorkloadIdentity{},
			},
			expected: "09.envoy.golden.yaml",
		}),
	)
})
