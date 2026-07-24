//nolint:staticcheck // SA1019 Test file: tests backward compatibility with deprecated core_rules.Rule
package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/plugin/xds"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

// passthroughListener builds an outbound passthrough listener with the default catch-all TCP
// chain the TransparentProxyGenerator would have produced, pointing at the given cluster.
func passthroughListener(name, address, cluster, stat string) *core_xds.Resource {
	return &core_xds.Resource{
		Name:   name,
		Origin: metadata.OriginTransparent,
		Resource: NewListenerBuilder(envoy_common.APIV3, name).
			Configure(OutboundListener(address, 15001, core_xds.SocketAddressProtocolTCP)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(TCPProxy(stat, plugins_xds.NewSplitBuilder().WithClusterName(cluster).WithWeight(100).Build())),
			)).MustBuild(),
	}
}

// clusterNames returns the names of all clusters currently in the resource set.
func clusterNames(rs *core_xds.ResourceSet) map[string]bool {
	out := map[string]bool{}
	for name := range rs.Resources(envoy_resource.ClusterType) {
		out[name] = true
	}
	return out
}

var _ = Describe("MeshPassthrough plugin edge cases", func() {
	// Reproduces the state the TransparentProxyGenerator leaves behind when passthrough is
	// enabled on the mesh: both IPv4 and IPv6 passthrough listeners plus their catch-all clusters.
	applyMatched := func(matches []api.Match) *core_xds.ResourceSet {
		rs := core_xds.NewResourceSet()
		rs.Add(passthroughListener("outbound:passthrough:ipv4", "0.0.0.0", "outbound:passthrough:ipv4", "outbound_passthrough_ipv4"))
		rs.Add(passthroughListener("outbound:passthrough:ipv6", "::", "outbound:passthrough:ipv6", "outbound_passthrough_ipv6"))
		c4, err := xds.CreateCluster(envoy_common.APIV3, "outbound:passthrough:ipv4", "tcp")
		Expect(err).ToNot(HaveOccurred())
		c6, err := xds.CreateCluster(envoy_common.APIV3, "outbound:passthrough:ipv6", "tcp")
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&core_xds.Resource{Name: "outbound:passthrough:ipv4", Origin: metadata.OriginTransparent, Resource: c4})
		rs.Add(&core_xds.Resource{Name: "outbound:passthrough:ipv6", Origin: metadata.OriginTransparent, Resource: c6})

		ctx := *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build()
		proxy := xds_builders.Proxy().
			WithApiVersion(envoy_common.APIV3).
			WithDataplane(builders.Dataplane().WithName("dp").WithMesh("default").WithAddress("127.0.0.1").
				WithTransparentProxying(15006, 15001, "ipv4").
				AddInbound(builders.Inbound().WithAddress("127.0.0.1").WithPort(17777).WithService("backend"))).
			WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshPassthroughType, core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						PassthroughMode: pointer.To[api.PassthroughMode]("Matched"),
						AppendMatch:     &matches,
					},
				}},
			})).
			Build()

		Expect(plugin.NewPlugin().(core_plugins.PolicyPlugin).Apply(rs, ctx, proxy)).To(Succeed())
		return rs
	}

	Describe("Matched mode with matches for a single IP family", func() {
		// This documents INTENDED behavior, not a bug. In "Matched" mode the default passthrough
		// clusters for BOTH families are removed and only the family that has matches gets a
		// reconfigured listener. The other family (here IPv6) keeps its default catch-all chain,
		// which now points at the deleted outbound:passthrough:ipv6 cluster. Envoy accepts the
		// listener (cluster references resolve lazily) and resets connections on that chain — i.e.
		// unmatched IPv6 traffic is BLOCKED. This is exactly the mechanism "None" mode uses to
		// block all passthrough, and the None-mode e2e test relies on it. Changing it here would
		// diverge Matched from None and let unmatched IPv6 traffic escape.
		It("blocks the unmatched family via the same cluster-removal mechanism as None mode", func() {
			rs := applyMatched([]api.Match{
				{Type: "IP", Value: "10.0.0.1", Protocol: "tcp", Port: pointer.To[uint32](80)},
			})

			clusters := clusterNames(rs)
			// Both default passthrough clusters are removed (unmatched IPv6 is blocked)...
			Expect(clusters).ToNot(HaveKey("outbound:passthrough:ipv6"))
			Expect(clusters).ToNot(HaveKey("outbound:passthrough:ipv4"))
			// ...only the matched IPv4 chain's cluster remains.
			Expect(clusters).To(HaveKey("meshpassthrough_tcp_10.0.0.1_80"))
			// Both passthrough listeners are still present.
			Expect(rs.Resources(envoy_resource.ListenerType)).To(HaveKey("outbound:passthrough:ipv6"))
			Expect(rs.Resources(envoy_resource.ListenerType)).To(HaveKey("outbound:passthrough:ipv4"))
		})
	})
})
