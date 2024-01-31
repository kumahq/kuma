package v1alpha1_test

import (
	"context"
	"path"
	"path/filepath"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

var _ = Describe("MeshRateLimit", func() {
	type sidecarTestCase struct {
		resources            []*core_xds.Resource
		fromRules            core_rules.FromRules
		inboundRateLimitsMap core_xds.InboundRateLimitsMap
		expectedListeners    []string
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
					Origin: generator.OriginInbound,
					Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:17777", false)).
							Configure(
								HttpInboundRoutes(
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
					Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
						)).MustBuild(),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{Num: 100, Interval: *test.ParseDuration("10s")},
									OnRateLimit: &api.OnRateLimit{
										Status: pointer.To(uint32(444)),
										Headers: &api.HeaderModifier{
											Add: []api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header",
													Value: "test-value",
												},
												{
													Name:  "x-kuma-rate-limit",
													Value: "other-value",
												},
											},
											Set: []api.HeaderKeyValue{
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
						Subset: core_rules.Subset{},
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
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"basic_listener_1.golden.yaml", "basic_listener_2.golden.yaml"},
		}),
		Entry("tcp rate limiter is disabled", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17778",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_rules.Subset{},
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
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"tcp_disabled.golden.yaml"},
		}),
		Entry("http rate limiter is disabled", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)).
						Configure(
							HttpInboundRoutes(
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
						Subset: core_rules.Subset{},
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
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"http_disabled.golden.yaml"},
		}),
		Entry("tcp rate limiter is not configured", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17778",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17778}: {{
						Subset: core_rules.Subset{},
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
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"tcp_disabled.golden.yaml"},
		}),
		Entry("http rate limiter is not configured", sidecarTestCase{
			resources: []*core_xds.Resource{{
				Name:   "inbound:127.0.0.1:17777",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)).
						Configure(
							HttpInboundRoutes(
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
						Subset: core_rules.Subset{},
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
			inboundRateLimitsMap: core_xds.InboundRateLimitsMap{},
			expectedListeners:    []string{"http_disabled.golden.yaml"},
		}),
	)

	It("should generate correct configuration for ExternalService with ZoneEgress", func() {
		// given
		rs := core_xds.NewResourceSet()

		// listener that matches
		listener, err := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 10002, core_xds.SocketAddressProtocolTCP).
			WithOverwriteName("test_listener").
			Configure(
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "external-service-1_mesh-1").Configure(
					MatchTransportProtocol("tls"),
					MatchServerNames("external-service-1{mesh=mesh-1}"),
					HttpConnectionManager("external-service-1", false),
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
					HttpConnectionManager("external-service-2", false),
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
			Origin:   egress.OriginEgress,
			Resource: listener,
		})

		// mesh with enabled mTLS and egress
		ctx := xds_builders.Context().
			WithMesh(builders.Mesh().
				WithName("mesh-1").
				WithBuiltinMTLSBackend("builtin-1").
				WithEnabledMTLSBackend("builtin-1").
				WithEgressRoutingEnabled()).
			Build()

		proxy := &core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
			ZoneEgressProxy: &core_xds.ZoneEgressProxy{
				ZoneEgressResource: &mesh.ZoneEgressResource{
					Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "mesh-1"},
					Spec: &mesh_proto.ZoneEgress{
						Networking: &mesh_proto.ZoneEgress_Networking{
							Address: "192.168.0.1",
							Port:    10002,
						},
					},
				},
				ZoneIngresses: []*mesh.ZoneIngressResource{},
				MeshResourcesList: []*core_xds.MeshResources{
					{
						Mesh: builders.Mesh().WithName("mesh-1").WithEnabledMTLSBackend("ca-1").WithBuiltinMTLSBackend("ca-1").Build(),
						ExternalServices: []*mesh.ExternalServiceResource{
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
													Subset: core_rules.MeshSubset(),
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
						ExternalServices: []*mesh.ExternalServiceResource{
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
													Subset: core_rules.MeshSubset(),
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
													Subset: core_rules.MeshSubset(),
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

	It("should generate proper Envoy config for MeshGateway Dataplanes", func() {
		// given
		gatewayRules := core_rules.GatewayRules{
			ToRules: core_rules.GatewayToRules{
				ByListener: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "192.168.0.1", Port: 8080}: {{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Local: &api.Local{
								HTTP: &api.LocalHTTP{
									RequestRate: &api.Rate{
										Num:      100,
										Interval: v1.Duration{Duration: 10 * time.Second},
									},
									OnRateLimit: &api.OnRateLimit{
										Status: pointer.To(uint32(444)),
										Headers: &api.HeaderModifier{
											Add: []api.HeaderKeyValue{
												{
													Name:  "x-kuma-rate-limit-header",
													Value: "test-value",
												},
												{
													Name:  "x-kuma-rate-limit",
													Value: "other-value",
												},
											},
										},
									},
								},
							},
						},
					}},
				},
			},
		}

		resources := xds_context.NewResources()
		resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
			Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
		}
		resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
			Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
		}

		xdsCtx := xds_samples.SampleContextWith(resources)
		proxy := xds_builders.Proxy().
			WithDataplane(samples.GatewayDataplaneBuilder()).
			WithPolicies(
				xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshRateLimitType, gatewayRules),
			).
			Build()
		for n, p := range core_plugins.Plugins().ProxyPlugins() {
			Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
		}
		gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
		generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
		Expect(err).NotTo(HaveOccurred())

		// when
		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

		// then
		Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())
		Expect(util_proto.ToYAML(generatedResources.ListOf(envoy_resource.RouteType)[0].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", "gateway_basic_routes.golden.yaml")))
		Expect(util_proto.ToYAML(generatedResources.ListOf(envoy_resource.ListenerType)[0].Resource)).To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", "gateway_basic_listener.golden.yaml")))
	})
})

func httpOutboundRoute(serviceName string) *meshhttproute_xds.HttpOutboundRouteConfigurer {
	return &meshhttproute_xds.HttpOutboundRouteConfigurer{
		Service: serviceName,
		Routes: []meshhttproute_xds.OutboundRoute{{
			Split: []envoy_common.Split{
				plugins_xds.NewSplitBuilder().WithClusterName(serviceName).WithWeight(100).Build(),
			},
			Matches: []meshhttproute_api.Match{
				{
					Path: &meshhttproute_api.PathMatch{
						Type:  meshhttproute_api.PathPrefix,
						Value: "/",
					},
				},
			},
		}},
		DpTags: map[string]map[string]bool{
			"kuma.io/service": {
				serviceName: true,
			},
		},
	}
}
