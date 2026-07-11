package v1alpha1_test

import (
	"path"
	"path/filepath"
	"time"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshratelimit/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test"
	test_matchers "github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v3/pkg/test/xds/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var _ = Describe("MeshRateLimit", func() {
	type sidecarTestCase struct {
		resources         []*core_xds.Resource
		fromRules         core_rules.FromRules
		expectedListeners []string
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
				WithPolicies(
					xds_builders.MatchedPolicies().
						WithFromPolicy(api.MeshRateLimitType, given.fromRules),
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
		Entry("basic listener: 2 inbounds one http and second tcp", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:   "inbound:127.0.0.1:17777",
					Origin: metadata.OriginInbound,
					Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								HttpInboundRoutes(
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
					Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
									OnRateLimit: &api.OnRateLimit{
										Status: pointer.To(uint32(444)),
										Headers: &api.HeaderModifier{
											Add: &[]api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header",
													Value: "test-value",
												},
												{
													Name:  "x-kuma-rate-limit",
													Value: "other-value",
												},
											},
											Set: &[]api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header-set",
													Value: "test-value",
												},
											},
										},
									},
								},
							},
						},
					}},
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
								TCP: &api.LocalTCP{
									ConnectionRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
									OnRateLimit: &api.OnRateLimit{
										Status: pointer.To(uint32(444)),
										Headers: &api.HeaderModifier{
											Add: &[]api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header",
													Value: "test-value",
												},
												{
													Name:  "x-kuma-rate-limit",
													Value: "other-value",
												},
											},
											Set: &[]api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header-set",
													Value: "test-value",
												},
											},
										},
									},
								},
							},
						},
					}},
					{Address: "127.0.0.1", Port: 17778}: {{
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
								TCP: &api.LocalTCP{
									ConnectionRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
			},
			expectedListeners: []string{"basic_listener_1.golden.yaml", "basic_listener_2.golden.yaml"},
		}),
		Entry("tcp rate limiter is disabled", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17778",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								TCP: &api.LocalTCP{
									Disabled:       pointer.To(true),
									ConnectionRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17778}: {{
						Conf: api.Conf{
							Local: &api.Local{
								TCP: &api.LocalTCP{
									Disabled:       pointer.To(true),
									ConnectionRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
			},
			expectedListeners:    []string{"tcp_disabled.golden.yaml"},
		}),
		Entry("http rate limiter is disabled", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
						Configure(
							HttpInboundRoutes(
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
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									Disabled:    pointer.To(true),
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									Disabled:    pointer.To(true),
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
								},
							},
						},
					}},
				},
			},
			expectedListeners:    []string{"http_disabled.golden.yaml"},
		}),
		Entry("tcp rate limiter is not configured", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17778",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								TCP: &api.LocalTCP{
									ConnectionRate: nil,
								},
							},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17778}: {{
						Conf: api.Conf{
							Local: &api.Local{
								TCP: &api.LocalTCP{
									ConnectionRate: nil,
								},
							},
						},
					}},
				},
			},
			expectedListeners:    []string{"tcp_disabled.golden.yaml"},
		}),
		Entry("http rate limiter is not configured", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
						Configure(
							HttpInboundRoutes(
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
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: nil,
								},
							},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: nil,
								},
							},
						},
					}},
				},
			},
			expectedListeners:    []string{"http_disabled.golden.yaml"},
		}),
		Entry("inbound listener with catch-all and rules[].matches[].spiffeID", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
						Configure(
							HttpInboundRoutes(
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
			}},
			fromRules: core_rules.FromRules{
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {
						{
							Conf: api.Conf{
								Local: &api.Local{
									HTTP: &api.LocalHTTP{
										RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
									},
								},
							},
						},
						{
							Match: &common_api.Match{
								SpiffeID: &common_api.SpiffeIDMatch{
									Type:  common_api.PrefixMatchType,
									Value: "spiffe://default/ns/clients/",
								},
							},
							Conf: api.Conf{
								Local: &api.Local{
									HTTP: &api.LocalHTTP{
										RequestRate: &api.Rate{Num: 5, Interval: *test.ParseDuration("1s")},
										OnRateLimit: &api.OnRateLimit{
											Status: pointer.To(uint32(429)),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedListeners:    []string{"inbound_matches_spiffeid_and_catchall.listener.golden.yaml"},
		}),
	)

	It("should generate proper Envoy config for zone egress listener with rules[].matches[].sni", func() {
		name := naming.ContextualZoneEgressListenerName("ze-port")
		resourceSet := core_xds.NewResourceSet()
		resourceSet.Add(&core_xds.Resource{
			Name:   name,
			Origin: metadata.OriginEgress,
			Resource: NewListenerBuilder(envoy_common.APIV3, name).
				Configure(InboundListener("10.20.30.40", 10002, core_xds.SocketAddressProtocolTCP, true)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
					Configure(MatchTransportProtocol("tls")).
					Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
					Configure(HttpConnectionManager("mes-http", false, nil, true)).
					Configure(AddFilterChainConfigurer(samples.MeshHttpOutboudWithSingleRoute("mes-http"))),
				)).
				MustBuild(),
		})

		meshRateLimit := api.NewMeshRateLimitResource()
		meshRateLimit.SetMeta(&test_model.ResourceMeta{
			Mesh:   "default",
			Name:   "zone-egress-rate-limit",
			Labels: map[string]string{},
		})
		meshRateLimit.Spec = &api.MeshRateLimit{
			TargetRef: &common_api.TargetRef{
				Kind:        common_api.Dataplane,
				SectionName: pointer.To("ze-port"),
			},
			Rules: &[]api.Rule{{
				Matches: &[]common_api.Match{{
					SNI: &common_api.SNIMatch{
						Type:  common_api.SNIExactMatchType,
						Value: "sni.extsvc.default.zone-1.aws-aurora.8443",
					},
				}},
				Default: api.Conf{
					Local: &api.Local{
						HTTP: &api.LocalHTTP{
							RequestRate: &api.Rate{Num: 50, Interval: *test.ParseDuration("5s")},
						},
					},
				},
			}},
		}
		Expect(meshRateLimit.Validate()).To(Succeed())

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
			WithPolicies(xds_builders.MatchedPolicies().
				With(func(policies *core_xds.MatchedPolicies) {
					policies.Dynamic[api.MeshRateLimitType] = core_xds.TypedMatchingPolicies{
						DataplanePolicies: []core_model.Resource{meshRateLimit},
					}
				})).
			Build()

		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(p.Apply(resourceSet, xds_samples.SampleContext(), proxy)).To(Succeed())
		Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[0].Resource)).To(
			test_matchers.MatchGoldenYAML(path.Join("testdata", "zoneegress_matches_sni.listener.golden.yaml")),
		)
	})

	It("should merge catch-all and SNI-specific listener rate limits on zone Egress", func() {
		name := naming.ContextualZoneEgressListenerName("ze-port")
		resourceSet := core_xds.NewResourceSet()
		resourceSet.Add(&core_xds.Resource{
			Name:   name,
			Origin: metadata.OriginEgress,
			Resource: NewListenerBuilder(envoy_common.APIV3, name).
				Configure(InboundListener("10.20.30.40", 10002, core_xds.SocketAddressProtocolTCP, true)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
					Configure(MatchTransportProtocol("tls")).
					Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
					Configure(HttpConnectionManager("mes-http", false, nil, true)).
					Configure(AddFilterChainConfigurer(samples.MeshHttpOutboudWithSingleRoute("mes-http"))),
				)).
				MustBuild(),
		})

		meshRateLimit := api.NewMeshRateLimitResource()
		meshRateLimit.SetMeta(&test_model.ResourceMeta{
			Mesh:   "default",
			Name:   "zone-egress-rate-limit-precedence",
			Labels: map[string]string{},
		})
		meshRateLimit.Spec = &api.MeshRateLimit{
			TargetRef: &common_api.TargetRef{
				Kind:        common_api.Dataplane,
				SectionName: pointer.To("ze-port"),
			},
			Rules: &[]api.Rule{
				{
					Default: api.Conf{
						Local: &api.Local{
							HTTP: &api.LocalHTTP{
								RequestRate: &api.Rate{
									Num:      10,
									Interval: *test.ParseDuration("1s"),
								},
								OnRateLimit: &api.OnRateLimit{
									Status: pointer.To(uint32(429)),
									Headers: &api.HeaderModifier{
										Add: &[]api.HeaderKeyValue{{
											Name:  "x-rate-limit-scope",
											Value: "common",
										}},
									},
								},
							},
						},
					},
				},
				{
					Matches: &[]common_api.Match{{
						SNI: &common_api.SNIMatch{
							Type:  common_api.SNIExactMatchType,
							Value: "sni.extsvc.default.zone-1.aws-aurora.8443",
						},
					}},
					Default: api.Conf{
						Local: &api.Local{
							HTTP: &api.LocalHTTP{
								RequestRate: &api.Rate{
									Num:      50,
									Interval: *test.ParseDuration("5s"),
								},
							},
						},
					},
				},
			},
		}
		Expect(meshRateLimit.Validate()).To(Succeed())

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
			WithPolicies(xds_builders.MatchedPolicies().
				With(func(policies *core_xds.MatchedPolicies) {
					policies.Dynamic[api.MeshRateLimitType] = core_xds.TypedMatchingPolicies{
						DataplanePolicies: []core_model.Resource{meshRateLimit},
					}
				})).
			Build()

		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(p.Apply(resourceSet, xds_samples.SampleContext(), proxy)).To(Succeed())

		listener, ok := resourceSet.ListOf(envoy_resource.ListenerType)[0].Resource.(*envoy_listener.Listener)
		Expect(ok).To(BeTrue())

		rateLimit := routeRateLimitFromZoneProxyListener(listener)
		Expect(rateLimit.GetTokenBucket().GetMaxTokens()).To(BeEquivalentTo(50))
		Expect(rateLimit.GetTokenBucket().GetFillInterval().AsDuration()).To(Equal(5 * time.Second))
		Expect(rateLimit.GetStatus().GetCode()).To(BeEquivalentTo(429))
		Expect(rateLimit.GetResponseHeadersToAdd()).To(HaveLen(1))
		Expect(rateLimit.GetResponseHeadersToAdd()[0].GetHeader().GetKey()).To(Equal("x-rate-limit-scope"))
		Expect(rateLimit.GetResponseHeadersToAdd()[0].GetHeader().GetValue()).To(Equal("common"))
	})

	It("should generate correct configuration for ExternalService with ZoneEgress", func() {
		// given
		rs := core_xds.NewResourceSet()

		// listener that matches
		listener, err := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 10002, core_xds.SocketAddressProtocolTCP, true).
			WithOverwriteName("test_listener").
			Configure(
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "external-service-1_mesh-1").Configure(
					MatchTransportProtocol("tls"),
					MatchServerNames("external-service-1{mesh=mesh-1}"),
					HttpConnectionManager("external-service-1", false, nil, true),
					AddFilterChainConfigurer(httpOutboundRoute("external-service-1")),
				)),
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "external-service-2_mesh-1").Configure(
					MatchTransportProtocol("tls"),
					MatchServerNames("external-service-2{mesh=mesh-1}"),
					TCPProxy("external-service-2", []envoy_common.Split{
						plugins_xds.NewSplitBuilder().WithClusterName("external-service-2").WithWeight(100).Build(),
					}...),
				)),
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "external-service-1_mesh-2").Configure(
					MatchTransportProtocol("tls"),
					MatchServerNames("external-service-1{mesh=mesh-2}"),
					TCPProxy("external-service-1", []envoy_common.Split{
						plugins_xds.NewSplitBuilder().WithClusterName("external-service-1").WithWeight(100).Build(),
					}...),
				)),
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "external-service-2_mesh-2").Configure(
					MatchTransportProtocol("tls"),
					MatchServerNames("external-service-2{mesh=mesh-2}"),
					HttpConnectionManager("external-service-2", false, nil, true),
					AddFilterChainConfigurer(httpOutboundRoute("external-service-2")),
				)),
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "internal-service-1_mesh-1").Configure(
					MatchTransportProtocol("tls"),
					MatchServerNames("internal-service-1{mesh=mesh-1}"),
					TCPProxy("internal-service-1"),
				)),
			).
			Build()
		Expect(err).ToNot(HaveOccurred())
		rs.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   metadata.OriginEgress,
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
			APIVersion: envoy_common.APIV3,
			ZoneEgressProxy: &core_xds.ZoneEgressProxy{
				ZoneEgressResource: &core_mesh.ZoneEgressResource{
					Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "mesh-1"},
					Spec: &mesh_proto.ZoneEgress{
						Networking: &mesh_proto.ZoneEgress_Networking{
							Address: "192.168.0.1",
							Port:    10002,
						},
					},
				},
				ZoneIngresses: []*core_mesh.ZoneIngressResource{},
				MeshResourcesList: []*core_xds.MeshResources{
					{
						Mesh: builders.Mesh().WithName("mesh-1").WithEnabledMTLSBackend("ca-1").WithBuiltinMTLSBackend("ca-1").Build(),
						ExternalServices: []*core_mesh.ExternalServiceResource{
							{
								Meta: &test_model.ResourceMeta{
									Mesh: "mesh-1",
									Name: "es-1",
								},
								Spec: &mesh_proto.ExternalService{
									Tags: map[string]string{
										"kuma.io/service":  "external-service-1",
										"kuma.io/protocol": "http",
									},
									Networking: &mesh_proto.ExternalService_Networking{
										Address: "externalservice-1.org",
									},
								},
							},
						},
						Dynamic: core_xds.ExternalServiceDynamicPolicies{
							"external-service-1": {
								api.MeshRateLimitType: core_xds.TypedMatchingPolicies{
									FromRules: core_rules.FromRules{
										Rules: map[core_rules.InboundListener]core_rules.Rules{
											{
												Address: "192.168.0.1", Port: 10002,
											}: {
												{
													Subset: subsetutils.MeshSubset(),
													Conf: api.Conf{
														Local: &api.Local{
															HTTP: &api.LocalHTTP{
																RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
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
					},
					{
						Mesh: builders.Mesh().WithName("mesh-2").WithEnabledMTLSBackend("ca-2").WithBuiltinMTLSBackend("ca-2").Build(),
						ExternalServices: []*core_mesh.ExternalServiceResource{
							{
								Meta: &test_model.ResourceMeta{
									Mesh: "mesh-2",
									Name: "es-1",
								},
								Spec: &mesh_proto.ExternalService{
									Tags: map[string]string{
										"kuma.io/service":  "external-service-1",
										"kuma.io/protocol": "tcp",
									},
									Networking: &mesh_proto.ExternalService_Networking{
										Address: "externalservice-1.org",
									},
								},
							},
							{
								Meta: &test_model.ResourceMeta{
									Mesh: "mesh-2",
									Name: "es-2",
								},
								Spec: &mesh_proto.ExternalService{
									Tags: map[string]string{
										"kuma.io/service":  "external-service-2",
										"kuma.io/protocol": "http",
									},
									Networking: &mesh_proto.ExternalService_Networking{
										Address: "externalservice-2.org",
									},
								},
							},
						},
						Dynamic: core_xds.ExternalServiceDynamicPolicies{
							"external-service-1": {
								api.MeshRateLimitType: core_xds.TypedMatchingPolicies{
									FromRules: core_rules.FromRules{
										Rules: map[core_rules.InboundListener]core_rules.Rules{
											{
												Address: "192.168.0.1", Port: 10002,
											}: {
												{
													Subset: subsetutils.MeshSubset(),
													Conf: api.Conf{
														Local: &api.Local{
															TCP: &api.LocalTCP{
																ConnectionRate: &api.Rate{Num: 22, Interval: *test.ParseDuration("22s")},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"external-service-2": {
								api.MeshRateLimitType: core_xds.TypedMatchingPolicies{
									FromRules: core_rules.FromRules{
										Rules: map[core_rules.InboundListener]core_rules.Rules{
											{
												Address: "192.168.0.1", Port: 10002,
											}: {
												{
													Subset: subsetutils.MeshSubset(),
													Conf: api.Conf{
														Local: &api.Local{
															HTTP: &api.LocalHTTP{
																RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
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
					},
				},
			},
		}

		// when
		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		err = p.Apply(rs, *ctx, proxy)
		Expect(err).ToNot(HaveOccurred())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())
		bytes, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(test_matchers.MatchGoldenYAML(path.Join("testdata", "basic_egress.golden.yaml")))
	})
})

func httpOutboundRoute(serviceName string) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	prefixMatch := meshhttproute_api.Match{
		Path: &meshhttproute_api.PathMatch{
			Type:  meshhttproute_api.PathPrefix,
			Value: "/",
		},
	}
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		VirtualHostName: serviceName,
		RouteConfigName: envoy_names.GetOutboundRouteName(serviceName),
		Routes: []meshhttproute_xds.OutboundRoute{{
			Split: []envoy_common.Split{
				plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
			},
			Name:  string(meshhttproute_api.HashMatches([]meshhttproute_api.Match{prefixMatch})),
			Match: prefixMatch,
		}},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				serviceName: true,
			},
		},
	}
}

func routeRateLimitFromZoneProxyListener(listener *envoy_listener.Listener) *envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit {
	Expect(listener.GetFilterChains()).ToNot(BeEmpty())
	Expect(listener.GetFilterChains()[0].GetFilters()).ToNot(BeEmpty())

	hcm := &envoy_hcm.HttpConnectionManager{}
	Expect(util_proto.UnmarshalAnyTo(listener.GetFilterChains()[0].GetFilters()[0].GetTypedConfig(), hcm)).To(Succeed())
	Expect(hcm.GetRouteConfig().GetVirtualHosts()).ToNot(BeEmpty())
	Expect(hcm.GetRouteConfig().GetVirtualHosts()[0].GetRoutes()).ToNot(BeEmpty())

	rateLimitAny := hcm.GetRouteConfig().GetVirtualHosts()[0].GetRoutes()[0].GetTypedPerFilterConfig()["envoy.filters.http.local_ratelimit"]
	Expect(rateLimitAny).ToNot(BeNil())

	rateLimit := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{}
	Expect(util_proto.UnmarshalAnyTo(rateLimitAny, rateLimit)).To(Succeed())
	return rateLimit
}
