package v1alpha1_test

import (
	"fmt"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
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

var _ = Describe("MeshPassthrough", func() {
	type testCase struct {
		resources               []*core_xds.Resource
		singleItemRules         core_rules.SingleItemRules
		meshPassthroughDisabled bool
		listenersGolden         string
		clustersGolden          string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			mesh := samples.MeshDefaultBuilder()
			if given.meshPassthroughDisabled {
				mesh.WithoutPassthrough()
			}
			context := *xds_builders.Context().
				WithMeshBuilder(mesh).
				Build()
			proxy := xds_builders.Proxy().
				WithApiVersion(envoy_common.APIV3).
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						WithTransparentProxying(15006, 15001, "ipv4").
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17777).
								WithService("backend"),
						),
				).
				WithPolicies(
					xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshPassthroughType, given.singleItemRules),
				).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

			// then
			Expect(getResource(resourceSet, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s", given.listenersGolden)))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s", given.clustersGolden)))
		},
		Entry("basic listener", testCase{
			resources: []*core_xds.Resource{
				{
					Name:   "outbound:passthrough:ipv4",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").
						Configure(OutboundListener("0.0.0.0", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv4", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv4").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
				{
					Name:   "outbound:passthrough:ipv6",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv6").
						Configure(OutboundListener("::", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv6", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv6").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							PassthroughMode: pointer.To[api.PassthroughMode](api.PassthroughMode("Matched")),
							AppendMatch: []api.Match{
								{
									Type:     api.MatchType("Domain"),
									Value:    "api.example.com",
									Port:     pointer.To(443),
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "example.com",
									Port:     pointer.To(443),
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "*.example.com",
									Port:     pointer.To(443),
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "example.com",
									Port:     pointer.To(8080),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "other.com",
									Port:     pointer.To(8080),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "grpcdomain.com",
									Port:     pointer.To(19000),
									Protocol: api.ProtocolType("grpc"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "http2.com",
									Port:     pointer.To(8080),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "http2.com",
									Port:     pointer.To(8080),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "*.example.com",
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "192.168.19.1",
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "192.168.19.1",
									Port:     pointer.To(10000),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "192.168.0.1",
									Port:     pointer.To(9091),
									Protocol: api.ProtocolType("tcp"),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "192.168.0.1/24",
									Protocol: api.ProtocolType("tcp"),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "192.168.0.1/30",
									Protocol: api.ProtocolType("tcp"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "trace-svc.datadog-agent.svc.cluster.local",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8126),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "trace-svc.datadog-agent.svc",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8126),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "172.16.0.0/12",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8126),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "cluster.test.local.dev",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8005),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "cluster-telemetry.test.local.dev",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8006),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "192.168.0.0/16",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8126),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "240.0.0.0/4",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(8126),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "www.google.com",
									Protocol: api.ProtocolType("http"),
									Port:     pointer.To(80),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "10.42.0.8",
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "b6e5:a45e:70ae:e77f:d24e:5023:375d:20a6",
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "9942:9abf:d0e0:f2da:2290:333b:e590:f497",
									Port:     pointer.To(9091),
									Protocol: api.ProtocolType("tcp"),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "b0ce:f616:4e74:28f7:427c:b969:8016:6344/64",
									Protocol: api.ProtocolType("tcp"),
								},
								{
									Type:     api.MatchType("CIDR"),
									Value:    "b0ce:f616:4e74:28f7:427c:b969:8016:6344/96",
									Protocol: api.ProtocolType("tcp"),
								},
							},
						},
					},
				},
			},
			listenersGolden: "basic.listener.golden.yaml",
			clustersGolden:  "basic.clusters.golden.yaml",
		}),
		Entry("only ipv4 rules", testCase{
			resources: []*core_xds.Resource{
				{
					Name:   "outbound:passthrough:ipv4",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").
						Configure(OutboundListener("0.0.0.0", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv4", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv4").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
				{
					Name:   "outbound:passthrough:ipv6",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv6").
						Configure(OutboundListener("::", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv6", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv6").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							AppendMatch: []api.Match{
								{
									Type:     api.MatchType("IP"),
									Value:    "192.168.0.0",
									Port:     pointer.To(80),
									Protocol: api.ProtocolType("tcp"),
								},
							},
						},
					},
				},
			},
			listenersGolden: "only-ipv4-rules.listener.golden.yaml",
			clustersGolden:  "only-ipv4-rules.clusters.golden.yaml",
		}),
		Entry("simple policy", testCase{
			resources: []*core_xds.Resource{
				{
					Name:   "outbound:passthrough:ipv4",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").
						Configure(OutboundListener("0.0.0.0", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv4", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv4").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
				{
					Name:   "outbound:passthrough:ipv6",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv6").
						Configure(OutboundListener("::", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv6", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv6").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							AppendMatch: []api.Match{
								{
									Type:     api.MatchType("Domain"),
									Value:    "api.example.com",
									Port:     pointer.To(80),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("IP"),
									Value:    "192.168.0.0",
									Port:     pointer.To(80),
									Protocol: api.ProtocolType("tcp"),
								},
							},
						},
					},
				},
			},
			listenersGolden: "simple.listener.golden.yaml",
			clustersGolden:  "simple.clusters.golden.yaml",
		}),
		Entry("disabled on policy but enabled on mesh", testCase{
			resources: []*core_xds.Resource{
				{
					Name:   "outbound:passthrough:ipv4",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").
						Configure(OutboundListener("0.0.0.0", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv4", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv4").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
				{
					Name:     "outbound:passthrough:ipv4",
					Origin:   generator.OriginTransparent,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").MustBuild(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							PassthroughMode: pointer.To[api.PassthroughMode](api.PassthroughMode("None")),
						},
					},
				},
			},
			listenersGolden: "disabled_on_policy.listeners.golden.yaml",
			clustersGolden:  "disabled_on_policy.clusters.golden.yaml",
		}),
		Entry("enabled on policy but disabled on mesh", testCase{
			resources: []*core_xds.Resource{
				{
					Name:   "outbound:passthrough:ipv4",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").
						Configure(OutboundListener("0.0.0.0", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv4", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv4").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							PassthroughMode: pointer.To[api.PassthroughMode](api.PassthroughMode("All")),
							AppendMatch: []api.Match{
								{
									Type:     api.MatchType("Domain"),
									Value:    "api.example.com",
									Port:     pointer.To(443),
									Protocol: api.ProtocolType("tls"),
								},
							},
						},
					},
				},
			},
			meshPassthroughDisabled: true,
			listenersGolden:         "enabled_on_policy.listeners.golden.yaml",
			clustersGolden:          "enabled_on_policy.clusters.golden.yaml",
		}),
		Entry("enabled on policy and on mesh", testCase{
			resources: []*core_xds.Resource{
				{
					Name:   "outbound:passthrough:ipv4",
					Origin: generator.OriginTransparent,
					Resource: NewListenerBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").
						Configure(OutboundListener("0.0.0.0", 15001, core_xds.SocketAddressProtocolTCP)).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TCPProxy("outbound_passthrough_ipv4", []envoy_common.Split{
								plugins_xds.NewSplitBuilder().WithClusterName("outbound:passthrough:ipv4").WithWeight(100).Build(),
							}...)),
						)).MustBuild(),
				},
				{
					Name:     "outbound:passthrough:ipv4",
					Origin:   generator.OriginTransparent,
					Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outbound:passthrough:ipv4").MustBuild(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							PassthroughMode: pointer.To[api.PassthroughMode](api.PassthroughMode("All")),
							AppendMatch: []api.Match{
								{
									Type:     api.MatchType("Domain"),
									Value:    "api.example.com",
									Port:     pointer.To(443),
									Protocol: api.ProtocolType("tls"),
								},
							},
						},
					},
				},
			},
			meshPassthroughDisabled: false,
			listenersGolden:         "enabled_on_policy_and_mesh.listeners.golden.yaml",
			clustersGolden:          "enabled_on_policy_and_mesh.clusters.golden.yaml",
		}),
	)
})
