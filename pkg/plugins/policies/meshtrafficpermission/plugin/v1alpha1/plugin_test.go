package v1alpha1_test

import (
	"path"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	policies_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	meshtrafficpermission "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
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
			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener,
			})

			// listener that is originated from inbound proxy generator but won't match: proves the
			// fail-closed default-deny fallback (mTLS inbound + zero MeshTrafficPermission rules still
			// gets an empty-policy, deny-all envoy.filters.network.rbac filter; see apply.golden.yaml)
			listener2, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8081, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener2").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener2", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener2.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener2,
			})

			// listener that matches but is not originated from inbound proxy generator
			listener3, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8082, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener3").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener3", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener3.GetName(),
				Origin:   "not-inbound-origin",
				Resource: listener3,
			})

			// listener that matches but it does not have mTLS
			listener4, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8083, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener4").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.HttpConnectionManager("test_listener", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener4.GetName(),
				Origin:   metadata.OriginInbound,
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
										Subset: []subsetutils.Tag{
											{Key: mesh_proto.ServiceTag, Value: "frontend"},
										},
										Conf: policies_api.Conf{
											Action: pointer.To[policies_api.Action]("Allow"),
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

		It("should ignore legacy 'from' MTP and default-deny under WorkloadIdentity", func() {
			// given
			rs := core_xds.NewResourceSet()
			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener,
			})

			proxy := xds_builders.Proxy().
				WithDataplane(builders.Dataplane().WithName("dp1").WithMesh("mesh-1").WithServices("backend")).
				WithWorkloadIdentity(&core_xds.WorkloadIdentity{}).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithFromPolicy(policies_api.MeshTrafficPermissionType, core_rules.FromRules{
							Rules: map[core_rules.InboundListener]core_rules.Rules{
								{Address: "192.168.0.1", Port: 8080}: {
									{
										Subset: []subsetutils.Tag{{Key: mesh_proto.ServiceTag, Value: "frontend"}},
										Conf:   policies_api.Conf{Action: pointer.To[policies_api.Action]("Allow")},
									},
								},
							},
						}),
				).
				Build()

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			Expect(p.Apply(rs, *ctx, proxy)).To(Succeed())

			// then
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			bytes, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply-workload-identity-ignores-from.golden.yaml")))
		})

		It("should enrich matching listener with RBAC filter using matching api", func() {
			// given
			rs := core_xds.NewResourceSet()
			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			// listener that matches
			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener,
			})

			// listener that is originated from inbound proxy generator but won't match
			listener2, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8081, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener2").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener2", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener2.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener2,
			})

			// listener that matches but is not originated from inbound proxy generator
			listener3, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8082, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener3").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener3", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener3.GetName(),
				Origin:   "not-inbound-origin",
				Resource: listener3,
			})

			listener4, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8083, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener4").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.HttpConnectionManager("test_listener4", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener4.GetName(),
				Origin:   metadata.OriginInbound,
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
							InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
								{
									Address: "192.168.0.1", Port: 8080,
								}: {
									{
										Conf: policies_api.RuleConf{
											Deny: &[]common_api.Match{
												{
													SpiffeID: &common_api.SpiffeIDMatch{
														Type:  common_api.ExactMatchType,
														Value: "spiffe://trust-domain.mesh/ns/backend/v1",
													},
												},
											},
											AllowWithShadowDeny: &[]common_api.Match{
												{
													SpiffeID: &common_api.SpiffeIDMatch{
														Type:  common_api.ExactMatchType,
														Value: "spiffe://trust-domain.mesh/ns/backend/v2",
													},
												},
											},
											Allow: &[]common_api.Match{
												{
													SpiffeID: &common_api.SpiffeIDMatch{
														Type:  common_api.PrefixMatchType,
														Value: "spiffe://trust-domain.mesh/ns/backend/",
													},
												},
											},
										},
										Origin: common.Origin{
											Resource: &test_model.ResourceMeta{
												Mesh: "default",
												Name: "mtp-1",
												Labels: map[string]string{
													mesh_proto.ZoneTag:          "zone-1",
													mesh_proto.KubeNamespaceTag: "ns-1",
												},
											},
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
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "matching-api.golden.yaml")))
		})

		It("should add TLS inspector to inbound listener when rules match on SNI", func() {
			// given
			rs := core_xds.NewResourceSet()
			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener,
			})

			proxy := xds_builders.Proxy().
				WithDataplane(builders.Dataplane().WithName("dp1").WithMesh("mesh-1").WithServices("backend")).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithFromPolicy(policies_api.MeshTrafficPermissionType, core_rules.FromRules{
							InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
								{Address: "192.168.0.1", Port: 8080}: {
									{
										Conf: policies_api.RuleConf{
											Allow: &[]common_api.Match{
												{SNI: &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: "sni.example"}},
											},
										},

										Origin: common.Origin{
											Resource: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "mtp-1"},
										},
									},
								},
							},
						}),
				).
				Build()

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			Expect(p.Apply(rs, *ctx, proxy)).To(Succeed())

			// then
			gotListener := rs.ListOf(envoy_resource.ListenerType)[0].Resource.(*envoy_listener.Listener)
			Expect(gotListener.ListenerFilters).To(HaveLen(1))
			Expect(gotListener.ListenerFilters[0].Name).To(Equal("envoy.filters.listener.tls_inspector"))
		})

		It("should not add TLS inspector to inbound listener when no SNI matcher is used", func() {
			// given
			rs := core_xds.NewResourceSet()
			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP, true).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil), nil, nil, false, false)).
					Configure(listeners.HttpConnectionManager("test_listener", false, nil, true)))).
				Build()
			Expect(err).ToNot(HaveOccurred())
			rs.Add(&core_xds.Resource{
				Name:     listener.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: listener,
			})

			proxy := xds_builders.Proxy().
				WithDataplane(builders.Dataplane().WithName("dp1").WithMesh("mesh-1").WithServices("backend")).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithFromPolicy(policies_api.MeshTrafficPermissionType, core_rules.FromRules{
							InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
								{Address: "192.168.0.1", Port: 8080}: {
									{
										Conf: policies_api.RuleConf{
											Allow: &[]common_api.Match{
												{SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://mesh-1/svc"}},
											},
										},

										Origin: common.Origin{
											Resource: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "mtp-1"},
										},
									},
								},
							},
						}),
				).
				Build()

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			Expect(p.Apply(rs, *ctx, proxy)).To(Succeed())

			// then
			gotListener := rs.ListOf(envoy_resource.ListenerType)[0].Resource.(*envoy_listener.Listener)
			Expect(gotListener.ListenerFilters).To(BeEmpty())
		})
	})

	Context("for DPP with ZoneEgress listener", func() {
		buildZEListener := func(address string, port uint32, name string) func() *core_xds.Resource {
			return func() *core_xds.Resource {
				listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, address, port, core_xds.SocketAddressProtocolTCP, true).
					WithOverwriteName(name).
					Configure(
						listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "mes-1").Configure(
							listeners.MatchTransportProtocol("tls"),
							listeners.MatchServerNames("sni.mes-1.default.zone-1.aws.8443"),
							listeners.DownstreamTlsContext(&envoy_tls.DownstreamTlsContext{}),
							listeners.TCPProxy("mes-1"),
						)),
						listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, "mes-2").Configure(
							listeners.MatchTransportProtocol("tls"),
							listeners.MatchServerNames("sni.mes-2.default.zone-1.aws.8444"),
							listeners.DownstreamTlsContext(&envoy_tls.DownstreamTlsContext{}),
							listeners.HttpConnectionManager("mes-2", false, nil, false),
						)),
					).
					Build()
				Expect(err).ToNot(HaveOccurred())
				return &core_xds.Resource{
					Name:     name,
					Origin:   metadata.OriginEgress,
					Resource: listener,
				}
			}
		}

		It("should apply deny-by-default RBAC when no MTP selects the ZE listener", func() {
			// given
			rs := core_xds.NewResourceSet()
			rs.Add(buildZEListener("192.168.0.1", 10002, "ze-listener")())

			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			proxy := xds_builders.Proxy().
				WithDataplane(builders.Dataplane().WithName("dp1").WithMesh("mesh-1").WithServices("backend")).
				WithWorkloadIdentity(&core_xds.WorkloadIdentity{}).
				Build()

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			Expect(p.Apply(rs, *ctx, proxy)).To(Succeed())

			// then
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			bytes, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply-ze-listener-no-mtp.golden.yaml")))
		})

		It("should apply MTP RBAC rules to all ZE listener filter chains", func() {
			// given
			rs := core_xds.NewResourceSet()
			rs.Add(buildZEListener("192.168.0.1", 10002, "ze-listener")())

			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder().WithName("mesh-1")).
				Build()

			proxy := xds_builders.Proxy().
				WithDataplane(builders.Dataplane().WithName("dp1").WithMesh("mesh-1").WithServices("backend")).
				WithWorkloadIdentity(&core_xds.WorkloadIdentity{}).
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithFromPolicy(policies_api.MeshTrafficPermissionType, core_rules.FromRules{
							InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
								{Address: "192.168.0.1", Port: 10002}: {
									{
										Conf: policies_api.RuleConf{
											Allow: &[]common_api.Match{
												{
													SNI: &common_api.SNIMatch{
														Type:  common_api.SNIExactMatchType,
														Value: "sni.mes-1.default.zone-1.aws.8443",
													},
												},
											},
										},
										Origin: common.Origin{
											Resource: &test_model.ResourceMeta{
												Mesh: "mesh-1",
												Name: "mtp-1",
											},
										},
									},
								},
							},
						}),
				).
				Build()

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			Expect(p.Apply(rs, *ctx, proxy)).To(Succeed())

			// then
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			bytes, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply-ze-listener-with-mtp.golden.yaml")))
		})
	})

})
