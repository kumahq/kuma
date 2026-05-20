package v1alpha1_test

import (
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	core_matchers "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/matchers"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshproxypatch/plugin/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v2/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v2/pkg/test/xds/samples"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v2/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/metadata"
)

func zoneEgressOnlyDataplane() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("zone-egress-1").
		WithAddress("192.168.0.10").
		WithoutInbounds().
		WithLabels(map[string]string{mesh_proto.ListenerZoneEgressLabel: "enabled"}).
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
				{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
					Address: "192.168.0.10",
					Port:    10002,
					Name:    "ze-port",
				},
			}
		})
}

func zoneIngressOnlyDataplane() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("zone-ingress-1").
		WithAddress("192.168.0.11").
		WithoutInbounds().
		WithLabels(map[string]string{mesh_proto.ListenerZoneIngressLabel: "enabled"}).
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
				{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
					Address: "192.168.0.11",
					Port:    10001,
					Name:    "zi-port",
				},
			}
		})
}

func mixedInboundAndZoneEgressDataplane() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("backend").
		WithAddress("192.168.0.1").
		AddInbound(builders.Inbound().
			WithService("backend").
			WithAddress("192.168.0.1").
			WithPort(17777)).
		WithLabels(map[string]string{mesh_proto.ListenerZoneEgressLabel: "enabled"}).
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
				{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
					Address: "192.168.0.1",
					Port:    10002,
					Name:    "ze-port",
				},
			}
		})
}

func zoneEgressListenerResource() core_xds.Resource {
	name := naming.ContextualZoneEgressListenerName("ze-port")
	return core_xds.Resource{
		Name:   name,
		Origin: metadata.OriginEgress,
		Resource: NewListenerBuilder(envoy_common.APIV3, name).
			Configure(InboundListener("192.168.0.10", 10002, core_xds.SocketAddressProtocolTCP, false)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
				Configure(HttpConnectionManager("mes-http", false, nil, true)),
			)).MustBuild(),
	}
}

func zoneIngressListenerResource() core_xds.Resource {
	name := naming.ContextualZoneIngressListenerName("zi-port")
	return core_xds.Resource{
		Name:   name,
		Origin: metadata.OriginIngress,
		Resource: NewListenerBuilder(envoy_common.APIV3, name).
			Configure(InboundListener("192.168.0.11", 10001, core_xds.SocketAddressProtocolTCP, false)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("backend{mesh=default}")),
			)).MustBuild(),
	}
}

func mixedInboundAndZoneEgressResources() []core_xds.Resource {
	inbound := core_xds.Resource{
		Name:   "inbound:192.168.0.1:17777",
		Origin: metadata.OriginInbound,
		Resource: NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 17777, core_xds.SocketAddressProtocolTCP, false).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(HttpConnectionManager("192.168.0.1:17777", false, nil, true)),
			)).MustBuild(),
	}
	egressName := naming.ContextualZoneEgressListenerName("ze-port")
	egress := core_xds.Resource{
		Name:   egressName,
		Origin: metadata.OriginEgress,
		Resource: NewListenerBuilder(envoy_common.APIV3, egressName).
			Configure(InboundListener("192.168.0.1", 10002, core_xds.SocketAddressProtocolTCP, false)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
				Configure(HttpConnectionManager("mes-http", false, nil, true)),
			)).MustBuild(),
	}
	return []core_xds.Resource{inbound, egress}
}

func newMeshProxyPatch(name string, targetRef *common_api.TargetRef, mods []api.Modification) *api.MeshProxyPatchResource {
	return &api.MeshProxyPatchResource{
		Meta: &test_model.ResourceMeta{Mesh: "default", Name: name},
		Spec: &api.MeshProxyPatch{
			TargetRef: targetRef,
			Default: api.Conf{
				AppendModifications: &mods,
			},
		},
	}
}

// MeshProxyPatch on a zone proxy Dataplane flows through the same
// SingleItemRules path as a regular Dataplane — the matcher (#16584)
// is what makes Dataplane-targeted policies reach the embedded zone
// proxy listeners. MeshProxyPatch itself is proxy-wide; scoping to a
// specific zone proxy DPP is done with computed labels on `targetRef`,
// exercised end-to-end by the universal envoyconfig suite.
var _ = Describe("MeshProxyPatch on zone proxy Dataplane", func() {
	type testCase struct {
		dp         *builders.DataplaneBuilder
		resources  []core_xds.Resource
		policies   []*api.MeshProxyPatchResource
		goldenFile string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			rs := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				rs.Add(&r)
			}

			dpp := given.dp.Build()
			meshResources := xds_context.NewResources()
			meshResources.MeshLocalResources[api.MeshProxyPatchType] = &api.MeshProxyPatchResourceList{
				Items: given.policies,
			}

			matched, err := core_matchers.MatchedPolicies(api.MeshProxyPatchType, dpp, meshResources)
			Expect(err).ToNot(HaveOccurred())

			xdsCtx := *xds_samples.SampleContextWith(meshResources).WithMeshBuilder(samples.MeshDefaultBuilder()).Build()
			proxy := xds_builders.Proxy().
				WithDataplane(given.dp).
				WithMetadata(&core_xds.DataplaneMetadata{}).
				Build()
			proxy.Policies.Dynamic = map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
				api.MeshProxyPatchType: matched,
			}

			Expect(plugin.NewPlugin().(core_plugins.PolicyPlugin).Apply(rs, xdsCtx, proxy)).To(Succeed())

			listenerYaml, err := util_yaml.GetResourcesToYaml(rs, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(listenerYaml).To(matchers.MatchGoldenYAML(filepath.Join("testdata", given.goldenFile+".listeners.golden.yaml")))

			clusterYaml, err := util_yaml.GetResourcesToYaml(rs, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterYaml).To(matchers.MatchGoldenYAML(filepath.Join("testdata", given.goldenFile+".clusters.golden.yaml")))
		},
		Entry("zone egress, scoped by listener-zoneegress label on Dataplane targetRef", testCase{
			dp:        zoneEgressOnlyDataplane(),
			resources: []core_xds.Resource{zoneEgressListenerResource()},
			policies: []*api.MeshProxyPatchResource{
				newMeshProxyPatch("by-label", &common_api.TargetRef{
					Kind:   common_api.Dataplane,
					Labels: pointer.To(map[string]string{mesh_proto.ListenerZoneEgressLabel: "enabled"}),
				}, []api.Modification{
					{Listener: &api.ListenerMod{
						Operation: api.ModOpPatch,
						Match:     &api.ListenerMatch{Name: pointer.To(naming.ContextualZoneEgressListenerName("ze-port"))},
						Value:     pointer.To("statPrefix: patched_egress\n"),
					}},
				}),
			},
			goldenFile: "zone-egress-listener-patch",
		}),
		Entry("zone ingress, scoped by listener-zoneingress label on Dataplane targetRef", testCase{
			dp:        zoneIngressOnlyDataplane(),
			resources: []core_xds.Resource{zoneIngressListenerResource()},
			policies: []*api.MeshProxyPatchResource{
				newMeshProxyPatch("by-label", &common_api.TargetRef{
					Kind:   common_api.Dataplane,
					Labels: pointer.To(map[string]string{mesh_proto.ListenerZoneIngressLabel: "enabled"}),
				}, []api.Modification{
					{Listener: &api.ListenerMod{
						Operation: api.ModOpPatch,
						Match:     &api.ListenerMatch{Name: pointer.To(naming.ContextualZoneIngressListenerName("zi-port"))},
						Value:     pointer.To("statPrefix: patched_ingress\n"),
					}},
				}),
			},
			goldenFile: "zone-ingress-listener-patch",
		}),
		Entry("mixed inbound + zone egress, Dataplane targetRef patches both listeners", testCase{
			dp:        mixedInboundAndZoneEgressDataplane(),
			resources: mixedInboundAndZoneEgressResources(),
			policies: []*api.MeshProxyPatchResource{
				newMeshProxyPatch("patch-both", &common_api.TargetRef{Kind: common_api.Dataplane}, []api.Modification{
					{Listener: &api.ListenerMod{
						Operation: api.ModOpPatch,
						Match:     &api.ListenerMatch{Name: pointer.To("inbound:192.168.0.1:17777")},
						Value:     pointer.To("statPrefix: patched_inbound\n"),
					}},
					{Listener: &api.ListenerMod{
						Operation: api.ModOpPatch,
						Match:     &api.ListenerMatch{Name: pointer.To(naming.ContextualZoneEgressListenerName("ze-port"))},
						Value:     pointer.To("statPrefix: patched_egress\n"),
					}},
					{Cluster: &api.ClusterMod{
						Operation: api.ModOpAdd,
						Value:     pointer.To("name: extra-cluster\nconnectTimeout: 7s\n"),
					}},
				}),
			},
			goldenFile: "mixed-patch-both",
		}),
	)
})
