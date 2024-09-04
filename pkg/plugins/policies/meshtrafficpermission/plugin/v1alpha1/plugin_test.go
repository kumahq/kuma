package v1alpha1_test

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	meshtrafficpermission "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

var _ = Describe("RBAC", func() {
	Context("for Dataplane", func() {
		It("should enrich matching listener with RBAC filter", func() {
			// given
			rs := core_xds.NewResourceSet()
			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			// listener that matches
			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil)).
					Configure(listeners.HttpConnectionManager("test_listener", false)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   generator.OriginInbound,
				Resource: listener,
			})

			// listener that is originated from inbound proxy generator but won't match
			listener2, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8081, core_xds.SocketAddressProtocolTCP).
				WithOverwriteName("test_listener2").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil)).
					Configure(listeners.HttpConnectionManager("test_listener2", false)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener2.GetName(),
				Origin:   generator.OriginInbound,
				Resource: listener2,
			})

			// listener that matches but is not originated from inbound proxy generator
			listener3, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8082, core_xds.SocketAddressProtocolTCP).
				WithOverwriteName("test_listener3").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil)).
					Configure(listeners.HttpConnectionManager("test_listener3", false)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener3.GetName(),
				Origin:   "not-inbound-origin",
				Resource: listener3,
			})

			// listener that matches but it does not have mTLS
			listener4, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8083, core_xds.SocketAddressProtocolTCP).
				WithOverwriteName("test_listener4").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.HttpConnectionManager("test_listener", false)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener4.GetName(),
				Origin:   generator.OriginInbound,
				Resource: listener4,
			})

			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("dp1").
						WithMesh("mesh-1").
						WithServices("backend"),
				).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithFromPolicy(policies_api.MeshTrafficPermissionType, core_rules.FromRules{
							Rules: map[core_rules.InboundListener]core_rules.Rules{
								{
									Address: "192.168.0.1", Port: 8080,
								}: {
									{
										Subset: []core_rules.Tag{
											{Key: mesh_proto.ServiceTag, Value: "frontend"},
										},
										Conf: policies_api.Conf{
											Action: "Allow",
										},
									},
								},
							},
						}),
				).
				Build()
			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			err = p.Apply(rs, *ctx, proxy)
			Expect(err).ToNot(HaveOccurred())

			// then
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			bytes, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply.golden.yaml")))
		})
	})

	Context("for ZoneEgress", func() {
		It("should enrich matching listener with RBAC filter", func() {
			// given
			rs := core_xds.NewResourceSet()

			// listener that matches
			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 10002, core_xds.SocketAddressProtocolTCP).
				WithOverwriteName("test_listener").
				Configure(
					listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-1_mesh-1").Configure(
						listeners.MatchTransportProtocol("tls"),
						listeners.MatchServerNames("external-service-1{mesh=mesh-1}"),
						listeners.HttpConnectionManager("external-service-1", false),
					)),
					listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-2_mesh-1").Configure(
						listeners.MatchTransportProtocol("tls"),
						listeners.MatchServerNames("external-service-2{mesh=mesh-1}"),
						listeners.TCPProxy("external-service-2"),
					)),
					listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "external-service-1_mesh-2").Configure(
						listeners.MatchTransportProtocol("tls"),
						listeners.MatchServerNames("external-service-1{mesh=mesh-2}"),
						listeners.TCPProxy("external-service-1"),
					)),
					listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "internal-service-1_mesh-1").Configure(
						listeners.MatchTransportProtocol("tls"),
						listeners.MatchServerNames("internal-service-1{mesh=mesh-1}"),
						listeners.TCPProxy("internal-service-1"),
					)),
				).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   egress.OriginEgress,
				Resource: listener,
			})

			// mesh with enabled mTLS and egress
			ctx := xds_builders.Context().
				WithMeshBuilder(builders.Mesh().
					WithName("mesh-1").
					WithBuiltinMTLSBackend("builtin-1").
					WithEnabledMTLSBackend("builtin-1").
					WithEgressRoutingEnabled()).
				Build()

			proxy := &core_xds.Proxy{
				APIVersion: envoy.APIV3,
				ZoneEgressProxy: &core_xds.ZoneEgressProxy{
					ZoneEgressResource: &mesh.ZoneEgressResource{
						Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "mesh-1"},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "192.168.0.1",
								Port:    10002,
							},
						},
					},
					ZoneIngresses: []*mesh.ZoneIngressResource{},
					MeshResourcesList: []*core_xds.MeshResources{
						{
							Mesh: builders.Mesh().WithName("mesh-1").WithEnabledMTLSBackend("ca-1").WithBuiltinMTLSBackend("ca-1").Build(),
							ExternalServices: []*mesh.ExternalServiceResource{
								{
									Meta: &test_model.ResourceMeta{
										Mesh: "mesh-1",
										Name: "es-1",
									},
									Spec: &mesh_proto.ExternalService{
										Tags: map[string]string{
											"kuma.io/service": "external-service-1",
										},
										Networking: &mesh_proto.ExternalService_Networking{
											Address: "externalservice-1.org",
										},
									},
								},
							},
							Dynamic: core_xds.ExternalServiceDynamicPolicies{
								"external-service-1": {
									policies_api.MeshTrafficPermissionType: core_xds.TypedMatchingPolicies{
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{
													Address: "192.168.0.1", Port: 10002,
												}: {
													{
														Subset: core_rules.MeshService("frontend"),
														Conf:   policies_api.Conf{Action: policies_api.Allow},
													},
												},
											},
										},
									},
								},
								"example-mes": {
									policies_api.MeshTrafficPermissionType: core_xds.TypedMatchingPolicies{
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{
													Address: "192.168.0.1", Port: 10002,
												}: {
													{
														Subset: core_rules.MeshSubset(),
														Conf:   policies_api.Conf{Action: policies_api.Allow},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Mesh: builders.Mesh().WithName("mesh-2").WithEnabledMTLSBackend("ca-2").WithBuiltinMTLSBackend("ca-2").Build(),
							ExternalServices: []*mesh.ExternalServiceResource{
								{
									Meta: &test_model.ResourceMeta{
										Mesh: "mesh-2",
										Name: "es-1",
									},
									Spec: &mesh_proto.ExternalService{
										Tags: map[string]string{
											"kuma.io/service": "external-service-1",
										},
										Networking: &mesh_proto.ExternalService_Networking{
											Address: "externalservice-1.org",
										},
									},
								},
							},
							Dynamic: core_xds.ExternalServiceDynamicPolicies{
								"external-service-1": {
									policies_api.MeshTrafficPermissionType: core_xds.TypedMatchingPolicies{
										FromRules: core_rules.FromRules{
											Rules: map[core_rules.InboundListener]core_rules.Rules{
												{
													Address: "192.168.0.1", Port: 10002,
												}: {
													{
														Subset: core_rules.MeshSubset(),
														Conf:   policies_api.Conf{Action: policies_api.Allow},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			err = p.Apply(rs, *ctx, proxy)
			Expect(err).ToNot(HaveOccurred())

			// then
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			bytes, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply-egress.golden.yaml")))
		})
	})
})
