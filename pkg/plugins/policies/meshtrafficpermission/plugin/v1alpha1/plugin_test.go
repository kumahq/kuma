package v1alpha1_test

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	meshtrafficpermission "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

<<<<<<< HEAD
var _ = Describe("Apply", func() {

	It("should enrich matching listener with RBAC filter", func() {
		// given
		rs := core_xds.NewResourceSet()

		// listener that matches
		listener, err := listeners.NewListenerBuilder(envoy.APIV3).
			Configure(listeners.InboundListener("test_listener", "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP)).
			Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3).
				Configure(listeners.HttpConnectionManager("test_listener", false)))).
			Build()
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   generator.OriginInbound,
			Resource: listener,
=======
var _ = Describe("RBAC", func() {
	Context("for Dataplane", func() {
		It("should enrich matching listener with RBAC filter", func() {
			// given
			rs := core_xds.NewResourceSet()
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: samples.MeshMTLSBuilder().WithName("mesh-1").Build(),
				},
			}

			// listener that matches
			listener, err := listeners.NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP).
				WithOverwriteName("test_listener").
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil))).
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
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil))).
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
					Configure(listeners.ServerSideMTLS(ctx.Mesh.Resource, envoy.NewSecretsTracker(ctx.Mesh.Resource.Meta.GetName(), nil))).
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

			proxy := &core_xds.Proxy{
				Dataplane: &mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "mesh-1"},
				},
				Policies: core_xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						policies_api.MeshTrafficPermissionType: {
							FromRules: core_rules.FromRules{
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
							},
						},
					},
				},
			}

			// when
			p := meshtrafficpermission.NewPlugin().(plugins.PolicyPlugin)
			err = p.Apply(rs, ctx, proxy)
			Expect(err).ToNot(HaveOccurred())

			// then
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			bytes, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply.golden.yaml")))
>>>>>>> 0e0489feb (fix(MeshTrafficPermission): support permissive mtls (#8171))
		})

		// listener that is originated from inbound proxy generator but won't match
		listener2, err := listeners.NewListenerBuilder(envoy.APIV3).
			Configure(listeners.InboundListener("test_listener2", "192.168.0.1", 8081, core_xds.SocketAddressProtocolTCP)).
			Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3).
				Configure(listeners.HttpConnectionManager("test_listener2", false)))).
			Build()
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&core_xds.Resource{
			Name:     listener2.GetName(),
			Origin:   generator.OriginInbound,
			Resource: listener2,
		})

		// listener that matches but is not originated from inbound proxy generator
		listener3, err := listeners.NewListenerBuilder(envoy.APIV3).
			Configure(listeners.InboundListener("test_listener3", "192.168.0.1", 8082, core_xds.SocketAddressProtocolTCP)).
			Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy.APIV3).
				Configure(listeners.HttpConnectionManager("test_listener3", false)))).
			Build()
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&core_xds.Resource{
			Name:     listener3.GetName(),
			Origin:   "not-inbound-origin",
			Resource: listener3,
		})

		// mesh with enabled mTLS
		ctx := xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &mesh.MeshResource{
					Meta: &test_model.ResourceMeta{Name: "mesh-1", Mesh: core_model.NoMesh},
					Spec: &mesh_proto.Mesh{
						Mtls: &mesh_proto.Mesh_Mtls{
							EnabledBackend: "builtin-1",
							Backends: []*mesh_proto.CertificateAuthorityBackend{
								{
									Name: "builtin-1",
									Type: "builtin",
								},
							},
						},
					},
				},
			},
		}

		proxy := &core_xds.Proxy{
			Dataplane: &mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "mesh-1"},
			},
			Policies: core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					policies_api.MeshTrafficPermissionType: {
						FromRules: core_xds.FromRules{
							Rules: map[core_xds.InboundListener]core_xds.Rules{
								{
									Address: "192.168.0.1", Port: 8080,
								}: {
									{
										Subset: []core_xds.Tag{
											{Key: mesh_proto.ServiceTag, Value: "frontend"},
										},
										Conf: policies_api.Conf{
											Action: "Allow",
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
		err = p.Apply(rs, ctx, proxy)
		Expect(err).ToNot(HaveOccurred())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())
		bytes, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "apply.golden.yaml")))
	})
})
