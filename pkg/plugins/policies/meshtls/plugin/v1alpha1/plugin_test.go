package v1alpha1_test

import (
	"fmt"
	"os"
	"path"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func getResource(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshTLS", func() {
	type testCase struct {
		caseName    string
		meshBuilder *builders.MeshBuilder
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			mesh := given.meshBuilder
			context := *xds_builders.Context().
				WithMeshBuilder(mesh).
				Build()
			resourceSet := core_xds.NewResourceSet()
			secretsTracker := envoy_common.NewSecretsTracker("default", nil)
			resources := getResources(secretsTracker, mesh)
			resourceSet.Add(resources...)

			policy := getPolicy(given.caseName)

			proxy := xds_builders.Proxy().
				WithSecretsTracker(secretsTracker).
				WithApiVersion(envoy_common.APIV3).
				WithOutbounds(xds_types.Outbounds{&xds_types.Outbound{
					LegacyOutbound: builders.Outbound().
						WithService("outgoing").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				}}).
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						WithTransparentProxying(15006, 15001, "ipv4").
						AddOutbound(
							builders.Outbound().
								WithAddress("127.0.0.1").
								WithPort(27777).
								WithService("outgoing"),
						).
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17777).
								WithService("backend"),
						).
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17778).
								WithService("frontend"),
						),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshTLSType, getFromRules(policy.Spec.From))).
				Build()

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

			// then
			Expect(getResource(resourceSet, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listeners.golden.yaml", given.caseName)))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.clusters.golden.yaml", given.caseName)))
		},
		Entry("strict with no mTLS on the mesh", testCase{
			caseName:    "strict-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("permissive with no mTLS on the mesh", testCase{
			caseName:    "permissive-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("strict with permissive mTLS on the mesh", testCase{
			caseName:    "strict-with-permissive-mtls",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
		}),
		Entry("permissive with permissive mTLS on the mesh", testCase{
			caseName:    "permissive-with-permissive-mtls",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
		}),
	)
})

func getResources(secretsTracker core_xds.SecretsTracker, mesh *builders.MeshBuilder) []*core_xds.Resource {
	return []*core_xds.Resource{
		{
			Name:   "inbound:127.0.0.1:17777",
			Origin: generator.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false)).
					Configure(
						listeners.HttpInboundRoutes(
							"backend",
							envoy_common.Routes{
								{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								},
							},
						),
					),
				)).MustBuild(),
		},
		{
			Name:   "inbound:127.0.0.1:17778",
			Origin: generator.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
				)).MustBuild(),
		},
		{
			Name:   "outgoing",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outgoing").
				Configure(clusters.ClientSideMTLS(secretsTracker, mesh.Build(), "outgoing", true, nil)).
				MustBuild(),
		},
	}
}

func getPolicy(caseName string) *api.MeshTLSResource {
	// setup
	meshTLS := api.NewMeshTLSResource()

	// when
	contents, err := os.ReadFile(path.Join("testdata", caseName+".policy.yaml"))
	Expect(err).ToNot(HaveOccurred())
	err = core_model.FromYAML(contents, &meshTLS.Spec)
	Expect(err).ToNot(HaveOccurred())

	meshTLS.SetMeta(&test_model.ResourceMeta{
		Name: "name",
		Mesh: core_model.DefaultMesh,
	})
	// and
	verr := meshTLS.Validate()
	Expect(verr).ToNot(HaveOccurred())

	return meshTLS
}

func getFromRules(froms []api.From) core_rules.FromRules {
	var rules []*core_rules.Rule

	for _, from := range froms {
		rules = append(rules, &core_rules.Rule{
			Subset: core_rules.Subset{},
			Conf:   from.Default,
		})
	}

	return core_rules.FromRules{
		Rules: map[core_rules.InboundListener]core_rules.Rules{
			{
				Address: "127.0.0.1",
				Port:    17777,
			}: rules,
			{
				Address: "127.0.0.1",
				Port:    17778,
			}: rules,
		},
	}
}
