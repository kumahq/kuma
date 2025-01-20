package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshProxyPatch", func() {
	type testCase struct {
		resources        []core_xds.Resource
		rules            core_rules.SingleItemRules
		expectedClusters []string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			context := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshProxyPatchType, given.rules)).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, proxy)).To(Succeed())
			policies_xds.ResourceArrayShouldEqual(resources.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("add and patch a cluster", testCase{
			resources: []core_xds.Resource{
				{
					Name:   "echo-http",
					Origin: generator.OriginOutbound,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "echo-http").
						MustBuild(),
				},
			},
			rules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							AppendModifications: []api.Modification{
								{
									Cluster: &api.ClusterMod{
										Operation: api.ModOpAdd,
										Value: pointer.To(`
name: new-cluster
connectTimeout: 5s
`),
									},
								},
								{
									Cluster: &api.ClusterMod{
										Operation: api.ModOpPatch,
										Match: &api.ClusterMatch{
											Name: pointer.To("echo-http"),
										},
										Value: pointer.To(`
connectTimeout: 100s
`),
									},
								},
							},
						},
					},
				},
			},
			expectedClusters: []string{
				`
name: echo-http
connectTimeout: 100s
`,
				`
name: new-cluster
connectTimeout: 5s`,
			},
		}),
	)
})
