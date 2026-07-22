package v1alpha1_test

import (
	"path"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshfaultinjection/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test"
	test_matchers "github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v3/pkg/test/xds/samples"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var _ = Describe("MeshFaultInjection", func() {
	type sidecarTestCase struct {
		resources         []*core_xds.Resource
		fromRules         core_rules.FromRules
		expectedListeners []string
	}

	policyOrigin := func(policyName string) common.Origin {
		return common.Origin{
			Resource: &test_model.ResourceMeta{
				Mesh: "default",
				Name: policyName,
				Labels: map[string]string{
					mesh_proto.ZoneTag:          "zone-1",
					mesh_proto.KubeNamespaceTag: "ns-1",
				},
			},
		}
	}

	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			context := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						AddInbound(
							builders.Inbound().
								WithTags(map[string]string{mesh_proto.ProtocolTag: "http"}).
								WithAddress("127.0.0.1").
								WithPort(17777).
								WithService("backend"),
						).
						AddInbound(
							builders.Inbound().
								WithTags(map[string]string{mesh_proto.ProtocolTag: "tcp"}).
								WithAddress("127.0.0.1").
								WithPort(17778).
								WithService("frontend"),
						),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshFaultInjectionType, given.fromRules)).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

			// then
			for i, expected := range given.expectedListeners {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[i].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("basic listener: 2 inbounds one http and second tcp", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:   "inbound:127.0.0.1:17777",
					Origin: metadata.OriginInbound,
					Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
						Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								listeners.HttpInboundRoutes(
									envoy_names.GetInboundRouteName("backend"),
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
					Origin: metadata.OriginInbound,
					Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
						Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {
						{
							Subset: subsetutils.Subset{
								{
									Key:   "kuma.io/service",
									Value: "demo-client",
								},
							},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: int32(444),
											Percentage: intstr.FromString("12"),
										},
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("55s"),
											Percentage: intstr.FromString("55"),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "111Mbps",
											Percentage: intstr.FromString("62.9"),
										},
									},
								},
							},
						},
						{
							Subset: subsetutils.Subset{
								{
									Key:   "kuma.io/service",
									Value: "demo-client",
									Not:   true,
								},
							},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: 111,
											Percentage: intstr.FromInt32(11),
										},
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("22s"),
											Percentage: intstr.FromInt32(22),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "333Mbps",
											Percentage: intstr.FromString("33.3"),
										},
									},
								},
							},
						},
					},
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Http: &[]api.FaultInjectionConf{
								{
									Abort: &api.AbortConf{
										HttpStatus: int32(444),
										Percentage: intstr.FromString("12.1"),
									},
									Delay: &api.DelayConf{
										Value:      *test.ParseDuration("55s"),
										Percentage: intstr.FromInt(55),
									},
									ResponseBandwidth: &api.ResponseBandwidthConf{
										Limit:      "111Mbps",
										Percentage: intstr.FromString("62.9"),
									},
								},
							},
						},
					}},
				},
			},
			expectedListeners: []string{"basic_listener_1.golden.yaml", "basic_listener_2.golden.yaml"},
		}),
		Entry("basic listener: 2 inbounds one http and second tcp, rules api", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:   "inbound:127.0.0.1:17777",
					Origin: metadata.OriginInbound,
					Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
						Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								listeners.HttpInboundRoutes(
									envoy_names.GetInboundRouteName("backend"),
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
					Origin: metadata.OriginInbound,
					Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
						Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {
						{
							Match: &common_api.Match{
								SpiffeID: &common_api.SpiffeIDMatch{
									Type:  common_api.PrefixMatchType,
									Value: "spiffe://trust-domain.mesh/",
								},
							},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: int32(444),
											Percentage: intstr.FromString("12"),
										},
									},
									{
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("55s"),
											Percentage: intstr.FromString("55"),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "111Mbps",
											Percentage: intstr.FromString("62.9"),
										},
									},
								},
							},
							Origin: policyOrigin("mfi-1"),
						},
					},
					{Address: "127.0.0.1", Port: 17778}: {
						{
							Match: &common_api.Match{
								SpiffeID: &common_api.SpiffeIDMatch{
									Type:  common_api.PrefixMatchType,
									Value: "spiffe://trust-domain.mesh/",
								},
							},
							Conf: api.Conf{
								Http: &[]api.FaultInjectionConf{
									{
										Abort: &api.AbortConf{
											HttpStatus: int32(444),
											Percentage: intstr.FromString("12"),
										},
										Delay: &api.DelayConf{
											Value:      *test.ParseDuration("55s"),
											Percentage: intstr.FromString("55"),
										},
										ResponseBandwidth: &api.ResponseBandwidthConf{
											Limit:      "111Mbps",
											Percentage: intstr.FromString("62.9"),
										},
									},
								},
							},
							Origin: policyOrigin("mfi-1"),
						},
					},
				},
			},
			expectedListeners: []string{"basic_listener_1_rules.golden.yaml", "basic_listener_2_rules.golden.yaml"},
		}),
	)

	It("should generate proper Envoy config for zone egress listener with rules[].matches[].sni", func() {
		name := naming.ContextualZoneEgressListenerName("ze-port")
		resourceSet := core_xds.NewResourceSet()
		resourceSet.Add(&core_xds.Resource{
			Name:   name,
			Origin: metadata.OriginEgress,
			Resource: listeners.NewListenerBuilder(envoy_common.APIV3, name).
				Configure(listeners.InboundListener("10.20.30.40", 10002, core_xds.SocketAddressProtocolTCP, true)).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
					Configure(listeners.MatchTransportProtocol("tls")).
					Configure(listeners.MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
					Configure(listeners.HttpConnectionManager("mes-http", false, nil, true)).
					Configure(listeners.AddFilterChainConfigurer(samples.MeshHttpOutboudWithSingleRoute("mes-http"))),
				)).
				MustBuild(),
		})

		proxy := xds_builders.Proxy().
			WithDataplane(
				builders.Dataplane().
					WithName("zone-proxy-egress").
					WithMesh("default").
					WithAddress("10.20.30.40").
					With(func(d *core_mesh.DataplaneResource) {
						d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{{
							Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
							Address: "10.20.30.40",
							Port:    10002,
							Name:    "ze-port",
						}}
					}),
			).
			WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshFaultInjectionType, core_rules.FromRules{
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "10.20.30.40", Port: 10002}: {{
						Match: &common_api.Match{
							SNI: &common_api.SNIMatch{
								Type:  common_api.SNIExactMatchType,
								Value: "sni.extsvc.default.zone-1.aws-aurora.8443",
							},
						},
						Conf: api.Conf{
							Http: &[]api.FaultInjectionConf{{
								Abort: &api.AbortConf{
									HttpStatus: 503,
									Percentage: intstr.FromString("50"),
								},
							}},
						},
						Origin: policyOrigin("mfi-zone-egress"),
					}},
				},
			})).
			Build()

		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resourceSet, xds_samples.SampleContext(), proxy)).To(Succeed())
		Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[0].Resource)).To(
			test_matchers.MatchGoldenYAML(path.Join("testdata", "zoneegress_matches_sni.listener.golden.yaml")),
		)
	})
})
