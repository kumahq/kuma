package v1alpha1_test

import (
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/plugin/v1alpha1"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshPassthrough", func() {
	type sidecarTestCase struct {
		resources         []*core_xds.Resource
		singleItemRules   core_rules.SingleItemRules
		expectedListeners []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			context := xds_samples.SampleContext()
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
			for i, expected := range given.expectedListeners {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[i].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("basic listener", sidecarTestCase{
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
							AppendMatch: []api.Match{
								{
									Type:     api.MatchType("Domain"),
									Value:    "api.example.com",
									Port:     pointer.To[int](443),
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "example.com",
									Port:     pointer.To[int](443),
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "*.example.com",
									Port:     pointer.To[int](443),
									Protocol: api.ProtocolType("tls"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "example.com",
									Port:     pointer.To[int](8080),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "other.com",
									Port:     pointer.To[int](8080),
									Protocol: api.ProtocolType("http"),
								},
								{
									Type:     api.MatchType("Domain"),
									Value:    "http2.com",
									Port:     pointer.To[int](8080),
									Protocol: api.ProtocolType("http2"),
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
									Value:    "192.168.0.1",
									Port:     pointer.To[int](9091),
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
							},
						},
					},
				},
			},
			expectedListeners: []string{"basic_listener.golden.yaml"},
		}),
	)
})
