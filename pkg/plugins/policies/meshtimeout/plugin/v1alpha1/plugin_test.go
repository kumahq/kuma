package v1alpha1

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshTimeout", func() {
	type sidecarTestCase struct {
		resources         []core_xds.Resource
		toRules           core_rules.ToRules
		fromRules         core_rules.FromRules
		expectedListeners []string
		expectedClusters  []string
	}
	DescribeTable("should generate proper Envoy config", func(given sidecarTestCase) {
		// given
		resourceSet := core_xds.NewResourceSet()
		for _, res := range given.resources {
			r := res
			resourceSet.Add(&r)
		}

		context := test_xds.CreateSampleMeshContext()
		proxy := xds.Proxy{
			Dataplane: builders.Dataplane().
				WithName("backend").
				WithMesh("default").
				WithAddress("127.0.0.1").
				AddOutboundsToServices("other-service", "second-service").
				WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http").
				Build(),
			Policies: xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
					api.MeshTimeoutType: {
						Type:      api.MeshTimeoutType,
						ToRules:   given.toRules,
						FromRules: given.fromRules,
					},
				},
			},
			Routing: core_xds.Routing{
				OutboundTargets: core_xds.EndpointMap{
					"other-service": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "http",
						},
					}},
					"other-service-_0_": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "http",
						},
					}},
					"second-service": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "tcp",
						},
					}},
				},
			},
		}

		// when
		plugin := NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())

		// then
		for i, expectedListener := range given.expectedListeners {
			Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[i].Resource)).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", expectedListener)))
		}
		for i, expectedCluster := range given.expectedClusters {
			Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[i].Resource)).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", expectedCluster)))
		}
	},
		Entry("http outbound route", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpOutboundListener(),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
				{
					Name:     "outbound-split",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service-_0_"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							ConnectionTimeout: test.ParseDuration("10s"),
							IdleTimeout:       test.ParseDuration("1h"),
							Http: &api.Http{
								RequestTimeout:        test.ParseDuration("5s"),
								StreamIdleTimeout:     test.ParseDuration("1s"),
								MaxStreamDuration:     test.ParseDuration("10m"),
								MaxConnectionDuration: test.ParseDuration("10m"),
							},
						},
					},
				},
			},
			expectedListeners: []string{"http_outbound_listener.golden.yaml"},
			expectedClusters: []string{
				"http_outbound_cluster.golden.yaml",
				"http_outbound_split_cluster.golden.yaml",
			},
		}),
		Entry("tcp outbound route", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:   "outbound",
					Origin: generator.OriginOutbound,
					Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 10002, core_xds.SocketAddressProtocolTCP).
						Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(TcpProxyDeprecated(
								"127.0.0.1:10002",
								envoy_common.NewCluster(
									envoy_common.WithService("backend"),
									envoy_common.WithWeight(100),
								),
							)),
						)).
						MustBuild(),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							ConnectionTimeout: test.ParseDuration("10s"),
							IdleTimeout:       test.ParseDuration("30s"),
						},
					},
				},
			},
			expectedClusters:  []string{"basic_tcp_cluster.golden.yaml"},
			expectedListeners: []string{"basic_tcp_listener.golden.yaml"},
		}),
		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: httpInboundListenerWith(),
				},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundServicePort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    80,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								ConnectionTimeout: test.ParseDuration("10s"),
								IdleTimeout:       test.ParseDuration("1h"),
								Http: &api.Http{
									RequestTimeout:        test.ParseDuration("5s"),
									StreamIdleTimeout:     test.ParseDuration("1s"),
									MaxStreamDuration:     test.ParseDuration("10m"),
									MaxConnectionDuration: test.ParseDuration("10m"),
								},
							},
						},
					},
				},
			},
			expectedClusters:  []string{"basic_inbound_cluster.golden.yaml"},
			expectedListeners: []string{"basic_inbound_listener.golden.yaml"},
		}),
		Entry("basic inbound route without defaults", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: httpInboundListenerWith(),
				},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundServicePort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    80,
					}: []*core_rules.Rule{},
				},
			},
			expectedClusters:  []string{"basic_without_defaults_inbound_cluster.golden.yaml"},
			expectedListeners: []string{"basic_without_defaults_inbound_listener.golden.yaml"},
		}),
		Entry("outbound with defaults when http conf missing", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpOutboundListener(),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "other-service",
							},
						},
						Conf: api.Conf{
							ConnectionTimeout: test.ParseDuration("10s"),
							IdleTimeout:       test.ParseDuration("1h"),
						},
					},
				},
			},
			expectedClusters:  []string{"outbound_with_defaults_cluster.golden.yaml"},
			expectedListeners: []string{"outbound_with_defaults_listener.golden.yaml"},
		}),
		Entry("default inbound conf when no from section specified", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: httpInboundListenerWith(),
				},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundServicePort)),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpOutboundListener(),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "other-service",
							},
						},
						Conf: api.Conf{
							ConnectionTimeout: test.ParseDuration("10s"),
							IdleTimeout:       test.ParseDuration("1h"),
						},
					},
				},
			},
			expectedClusters:  []string{"default_inbound_cluster.golden.yaml", "modified_outbound_cluster.golden.yaml"},
			expectedListeners: []string{"default_inbound_listener.golden.yaml", "modified_outbound_listener.golden.yaml"},
		}),
		Entry("default outbound conf when no to section specified", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: httpInboundListenerWith(),
				},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundServicePort)),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpOutboundListener(),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    80,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								ConnectionTimeout: test.ParseDuration("10s"),
								IdleTimeout:       test.ParseDuration("1h"),
								Http: &api.Http{
									RequestTimeout:        test.ParseDuration("5s"),
									StreamIdleTimeout:     test.ParseDuration("1s"),
									MaxStreamDuration:     test.ParseDuration("10m"),
									MaxConnectionDuration: test.ParseDuration("10m"),
								},
							},
						},
					},
				},
			},
			expectedClusters:  []string{"modified_inbound_cluster.golden.yaml", "default_outbound_cluster.golden.yaml"},
			expectedListeners: []string{"modified_inbound_listener.golden.yaml", "default_outbound_listener.golden.yaml"},
		}),
		Entry("default outbound conf when no to section specified", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: httpInboundListenerWith(),
				},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundServicePort)),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpOutboundListener(),
				},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			expectedClusters:  []string{"original_inbound_cluster.golden.yaml", "original_outbound_cluster.golden.yaml"},
			expectedListeners: []string{"original_inbound_listener.golden.yaml", "original_outbound_listener.golden.yaml"},
		}),
		Entry("timeouts per http route", sidecarTestCase{
			resources: []core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpOutboundListenerWithSeveralRoutes(),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "other-service",
							},
							{
								Key:   core_rules.RuleMatchesHashTag,
								Value: "qiQ6EM62EfIBogYTOW3r8RUBRaRsY8B+t8G7DE5BNB8=", // '[{"path":{"value":"/","type":"PathPrefix"}}]'
							},
						},
						Conf: api.Conf{
							Http: &api.Http{
								RequestTimeout:    test.ParseDuration("99s"),
								StreamIdleTimeout: test.ParseDuration("999s"),
							},
						},
					},
					{
						Subset: core_rules.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "other-service",
							},
							{
								Key:   core_rules.RuleMatchesHashTag,
								Value: "Lv6cpFf/JzQZSvl97nnZZFjFcZQbqoejHncFutEisJQ=", // '[{"path":{"value":"/another-backend","type":"Exact"}},{"method":"GET"}]'
							},
						},
						Conf: api.Conf{
							Http: &api.Http{
								RequestTimeout:    test.ParseDuration("88s"),
								StreamIdleTimeout: test.ParseDuration("888s"),
							},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_listener_with_different_timeouts_per_route.yaml"},
		}),
	)

	type gatewayTestCase struct {
		toRules core_rules.ToRules
		routes  []*core_mesh.MeshGatewayRouteResource
	}
	DescribeTable("should generate proper Envoy config", func(given gatewayTestCase) {
		resources := xds_context.NewResources()
		resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
			Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
		}
		resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
			Items: append([]*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()}, given.routes...),
		}

		xdsCtx := test_xds.CreateSampleMeshContextWith(resources)
		proxy := xds.Proxy{
			APIVersion: "v3",
			Dataplane:  samples.GatewayDataplane(),
			Policies: xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
					api.MeshTimeoutType: {
						Type:    api.MeshTimeoutType,
						ToRules: given.toRules,
					},
				},
			},
			Routing: core_xds.Routing{
				OutboundTargets: core_xds.EndpointMap{
					"backend": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "http",
						},
					}},
					"other-service": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "http",
						},
					}},
				},
			},
			RuntimeExtensions: map[string]interface{}{},
		}
		Expect(gateway_plugin.NewPlugin().(core_plugins.ProxyPlugin).Apply(context.Background(), xdsCtx.Mesh, &proxy)).To(Succeed())
		gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
		generatedResources, err := gatewayGenerator.Generate(context.Background(), xdsCtx, &proxy)
		Expect(err).NotTo(HaveOccurred())

		// when
		plugin := NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(generatedResources, xdsCtx, &proxy)).To(Succeed())

		nameSplit := strings.Split(GinkgoT().Name(), " ")
		name := nameSplit[len(nameSplit)-1]

		// then
		Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", fmt.Sprintf("%s.gateway.listener.golden.yaml", name))))
		Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", fmt.Sprintf("%s.gateway.cluster.golden.yaml", name))))
		Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", fmt.Sprintf("%s.gateway.route.golden.yaml", name))))
	}, Entry("basic", gatewayTestCase{
		toRules: core_rules.ToRules{
			Rules: []*core_rules.Rule{
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.Conf{
						ConnectionTimeout: test.ParseDuration("10s"),
						IdleTimeout:       test.ParseDuration("1h"),
						Http: &api.Http{
							RequestTimeout:        test.ParseDuration("5s"),
							StreamIdleTimeout:     test.ParseDuration("1s"),
							MaxStreamDuration:     test.ParseDuration("10m"),
							MaxConnectionDuration: test.ParseDuration("10m"),
						},
					},
				},
			},
		},
	}), Entry("no-default-idle-timeout", gatewayTestCase{
		toRules: core_rules.ToRules{
			Rules: []*core_rules.Rule{
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.Conf{
						ConnectionTimeout: test.ParseDuration("10s"),
						IdleTimeout:       test.ParseDuration("1h"),
						Http: &api.Http{
							RequestTimeout:        test.ParseDuration("5s"),
							MaxStreamDuration:     test.ParseDuration("10m"),
							MaxConnectionDuration: test.ParseDuration("10m"),
						},
					},
				},
			},
		},
	}), Entry("no-route-level-timeouts", gatewayTestCase{
		routes: []*core_mesh.MeshGatewayRouteResource{
			builders.GatewayRoute().
				WithName("sample-gateway-route").
				WithGateway("sample-gateway").
				WithExactMatchHttpRoute("/", "backend", "other-service").
				Build(),
		},
		toRules: core_rules.ToRules{
			Rules: []*core_rules.Rule{
				{
					Subset: core_rules.MeshService("backend"),
					Conf: api.Conf{
						Http: &api.Http{
							RequestTimeout: test.ParseDuration("24s"),
						},
					},
				},
			},
		},
	}))
})

func getResourceYaml(list core_xds.ResourceList) []byte {
	actualResource, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualResource
}

func httpOutboundListener() envoy_common.NamedResource {
	return createListener(
		NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 10001, core_xds.SocketAddressProtocolTCP),
		AddFilterChainConfigurer(&meshhttproute_xds.HttpOutboundRouteConfigurer{
			Service: "backend",
			Routes: []meshhttproute_xds.OutboundRoute{{
				Split: []envoy_common.Split{
					plugins_xds.NewSplitBuilder().WithClusterName("backend").WithWeight(100).Build(),
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
					"web": true,
				},
			},
		}))
}

func httpOutboundListenerWithSeveralRoutes() envoy_common.NamedResource {
	return createListener(
		NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 10001, core_xds.SocketAddressProtocolTCP),
		AddFilterChainConfigurer(&meshhttproute_xds.HttpOutboundRouteConfigurer{
			Service: "other-service",
			Routes: []meshhttproute_xds.OutboundRoute{
				{
					Split: []envoy_common.Split{
						plugins_xds.NewSplitBuilder().WithClusterName("other-service").WithWeight(100).Build(),
					},
					Matches: []meshhttproute_api.Match{
						{
							Path: &meshhttproute_api.PathMatch{
								Type:  meshhttproute_api.Exact,
								Value: "/another-backend",
							},
							Method: pointer.To[meshhttproute_api.Method]("GET"),
						},
					},
				},
				{
					Split: []envoy_common.Split{
						plugins_xds.NewSplitBuilder().WithClusterName("other-service").WithWeight(100).Build(),
					},
					Matches: []meshhttproute_api.Match{
						{
							Path: &meshhttproute_api.PathMatch{
								Type:  meshhttproute_api.PathPrefix,
								Value: "/",
							},
						},
					},
				},
			},
			DpTags: map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
		}))
}

func httpInboundListenerWith() envoy_common.NamedResource {
	return createListener(
		NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 80, core_xds.SocketAddressProtocolTCP),
		HttpInboundRoutes(
			"backend",
			envoy_common.Routes{{
				Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
					envoy_common.WithService("backend"),
					envoy_common.WithWeight(100),
				)},
			}},
		))
}

func createListener(builder *ListenerBuilder, route FilterChainBuilderOpt) envoy_common.NamedResource {
	return builder.
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(HttpConnectionManager(builder.GetName(), false)).
			Configure(route),
		)).MustBuild()
}
