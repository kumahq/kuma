package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	plugin_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
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
)

var _ = Describe("MeshRetry", func() {
	type testCase struct {
		resources        []core_xds.Resource
		toRules          core_rules.ToRules
		goldenFilePrefix string
	}

	DescribeTable("should generate proper Envoy config", func(given testCase) {
		// given
		resourceSet := core_xds.NewResourceSet()
		for _, res := range given.resources {
			r := res
			resourceSet.Add(&r)
		}

		context := *xds_builders.Context().
			WithMesh(samples.MeshDefaultBuilder()).
			WithResources(xds_context.NewResources()).
			AddServiceProtocol("http-service", core_mesh.ProtocolHTTP).
			AddServiceProtocol("tcp-service", core_mesh.ProtocolTCP).
			AddServiceProtocol("grpc-service", core_mesh.ProtocolGRPC).
			AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
			Build()

		proxy := xds_builders.Proxy().
			WithDataplane(builders.Dataplane().
				WithName("backend").
				WithMesh("default").
				WithAddress("127.0.0.1").
				AddOutboundsToServices("http-service", "grpc-service", "tcp-service").
				WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http")).
			WithRouting(
				xds_builders.Routing().
					WithOutboundTargets(xds_builders.EndpointMap().
						AddEndpoint("http-service", xds_samples.HttpEndpointBuilder()).
						AddEndpoint("tcp-service", xds_samples.TcpEndpointBuilder()).
						AddEndpoint("grpc-service", xds_samples.GrpcEndpointBuilder()),
					),
			).
			WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshRetryType, given.toRules)).
			Build()

		// when
		plugin := plugin_v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

		// then
		Expect(getResourceYaml(resourceSet.ListOf(envoy_resource.ListenerType))).
			To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.listeners.golden.yaml", given.goldenFilePrefix))))
	},
		Entry("http retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListener(10001),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							HTTP: &api.HTTP{
								NumRetries:    pointer.To[uint32](1),
								PerTryTimeout: test.ParseDuration("2s"),
								BackOff: &api.BackOff{
									BaseInterval: test.ParseDuration("3s"),
									MaxInterval:  test.ParseDuration("4s"),
								},
								RateLimitedBackOff: &api.RateLimitedBackOff{
									MaxInterval: test.ParseDuration("5s"),
									ResetHeaders: &[]api.ResetHeader{
										{
											Name:   "retry-after-http",
											Format: "Seconds",
										},
										{
											Name:   "x-retry-after-http",
											Format: "UnixTimestamp",
										},
									},
								},
								RetryOn: &[]api.HTTPRetryOn{
									api.All5xx,
									api.GatewayError,
									api.Reset,
									api.Retriable4xx,
									api.ConnectFailure,
									api.EnvoyRatelimited,
									api.RefusedStream,
									api.Http3PostConnectFailure,
									api.HttpMethodConnect,
									api.HttpMethodDelete,
									api.HttpMethodGet,
									api.HttpMethodHead,
									api.HttpMethodOptions,
									api.HttpMethodPatch,
									api.HttpMethodPost,
									api.HttpMethodPut,
									api.HttpMethodTrace,
									"429",
								},
								RetriableResponseHeaders: &[]common_api.HeaderMatch{
									{
										Type:  pointer.To(common_api.HeaderMatchRegularExpression),
										Name:  "x-retry-regex",
										Value: ".*",
									},
									{
										Type:  pointer.To(common_api.HeaderMatchExact),
										Name:  "x-retry-exact",
										Value: "exact-value",
									},
								},
								RetriableRequestHeaders: &[]common_api.HeaderMatch{
									{
										Type:  pointer.To(common_api.HeaderMatchPrefix),
										Name:  "x-retry-prefix",
										Value: "prefix-",
									},
								},
								HostSelection: &[]api.Predicate{
									{
										PredicateType: "OmitPreviousHosts",
									},
									{
										PredicateType:   "OmitPreviousPriorities",
										UpdateFrequency: 2,
									},
									{
										PredicateType: "OmitHostsWithTags",
										Tags: map[string]string{
											"test": "test",
										},
									},
								},
								HostSelectionMaxAttempts: pointer.To(int64(2)),
							},
						},
					},
				},
			},
			goldenFilePrefix: "http",
		}),
		Entry("grpc retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListener(10002),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "grpc-service",
						}},
						Conf: api.Conf{
							GRPC: &api.GRPC{
								NumRetries:    pointer.To[uint32](11),
								PerTryTimeout: test.ParseDuration("12s"),
								BackOff: &api.BackOff{
									BaseInterval: test.ParseDuration("13s"),
									MaxInterval:  test.ParseDuration("14s"),
								},
								RateLimitedBackOff: &api.RateLimitedBackOff{
									MaxInterval: test.ParseDuration("15s"),
									ResetHeaders: &[]api.ResetHeader{
										{
											Name:   "retry-after-grpc",
											Format: "Seconds",
										},
										{
											Name:   "x-retry-after-grpc",
											Format: "UnixTimestamp",
										},
									},
								},
								RetryOn: &[]api.GRPCRetryOn{
									api.Canceled,
									api.DeadlineExceeded,
									api.Internal,
									api.ResourceExhausted,
									api.Unavailable,
								},
							},
						},
					},
				},
			},
			goldenFilePrefix: "grpc",
		}),
		Entry("tcp retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: tcpListener(10003),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							TCP: &api.TCP{
								MaxConnectAttempt: pointer.To[uint32](21),
							},
						},
					},
				},
			},
			goldenFilePrefix: "tcp",
		}),
	)

	type gatewayTestCase struct {
		rules            core_rules.GatewayRules
		goldenFilePrefix string
		gateways         []*core_mesh.MeshGatewayResource
		gatewayRoutes    []*core_mesh.MeshGatewayRouteResource
	}

	DescribeTable("should generate proper Envoy config for MeshGateways",
		func(given gatewayTestCase) {
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: given.gateways,
			}
			resources.MeshLocalResources[core_mesh.RetryType] = &core_mesh.RetryResourceList{
				Items: []*core_mesh.RetryResource{{
					Meta: &test_model.ResourceMeta{Name: "retry1"},
					Spec: &mesh_proto.Retry{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									mesh_proto.ServiceTag: "*",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									mesh_proto.ServiceTag: "*",
								},
							},
						},
						Conf: &mesh_proto.Retry_Conf{
							Http: &mesh_proto.Retry_Conf_Http{
								NumRetries:    util_proto.UInt32(5),
								PerTryTimeout: util_proto.Duration(16 * time.Second),
								BackOff: &mesh_proto.Retry_Conf_BackOff{
									BaseInterval: util_proto.Duration(25 * time.Millisecond),
									MaxInterval:  util_proto.Duration(250 * time.Millisecond),
								},
							},
							Tcp: &mesh_proto.Retry_Conf_Tcp{
								MaxConnectAttempts: 5,
							},
							Grpc: &mesh_proto.Retry_Conf_Grpc{
								NumRetries:    util_proto.UInt32(5),
								PerTryTimeout: util_proto.Duration(16 * time.Second),
								BackOff: &mesh_proto.Retry_Conf_BackOff{
									BaseInterval: util_proto.Duration(25 * time.Millisecond),
									MaxInterval:  util_proto.Duration(250 * time.Millisecond),
								},
							},
						},
					},
				}},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: given.gatewayRoutes,
			}

			xdsCtx := xds_samples.SampleContextWith(resources)
			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshRetryType, given.rules)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin_v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.listeners.golden.yaml", given.goldenFilePrefix))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.routes.golden.yaml", given.goldenFilePrefix))))
		},
		Entry("http retry", gatewayTestCase{
			goldenFilePrefix: "gateway.http",
			rules: core_rules.GatewayRules{
				ToRules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "192.168.0.1", Port: 8080}: {
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								HTTP: &api.HTTP{
									NumRetries:    pointer.To[uint32](1),
									PerTryTimeout: test.ParseDuration("2s"),
									BackOff: &api.BackOff{
										BaseInterval: test.ParseDuration("3s"),
										MaxInterval:  test.ParseDuration("4s"),
									},
									RateLimitedBackOff: &api.RateLimitedBackOff{
										MaxInterval: test.ParseDuration("5s"),
										ResetHeaders: &[]api.ResetHeader{
											{
												Name:   "retry-after-http",
												Format: "Seconds",
											},
											{
												Name:   "x-retry-after-http",
												Format: "UnixTimestamp",
											},
										},
									},
									RetryOn: &[]api.HTTPRetryOn{
										api.All5xx,
										api.GatewayError,
										api.Reset,
										api.Retriable4xx,
										api.ConnectFailure,
										api.EnvoyRatelimited,
										api.RefusedStream,
										api.Http3PostConnectFailure,
										api.HttpMethodConnect,
										api.HttpMethodDelete,
										api.HttpMethodGet,
										api.HttpMethodHead,
										api.HttpMethodOptions,
										api.HttpMethodPatch,
										api.HttpMethodPost,
										api.HttpMethodPut,
										api.HttpMethodTrace,
										"429",
									},
									RetriableResponseHeaders: &[]common_api.HeaderMatch{
										{
											Type:  pointer.To(common_api.HeaderMatchRegularExpression),
											Name:  "x-retry-regex",
											Value: ".*",
										},
										{
											Type:  pointer.To(common_api.HeaderMatchExact),
											Name:  "x-retry-exact",
											Value: "exact-value",
										},
									},
									RetriableRequestHeaders: &[]common_api.HeaderMatch{
										{
											Type:  pointer.To(common_api.HeaderMatchPrefix),
											Name:  "x-retry-prefix",
											Value: "prefix-",
										},
									},
									HostSelection: &[]api.Predicate{
										{
											PredicateType: "OmitPreviousHosts",
										},
										{
											PredicateType:   "OmitPreviousPriorities",
											UpdateFrequency: 2,
										},
										{
											PredicateType: "OmitHostsWithTags",
											Tags: map[string]string{
												"test": "test",
											},
										},
									},
									HostSelectionMaxAttempts: pointer.To(int64(2)),
								},
							},
						},
					},
				},
			},
			gateways:      []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			gatewayRoutes: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
		}),
		Entry("tcp retry", gatewayTestCase{
			goldenFilePrefix: "gateway.tcp",
			rules: core_rules.GatewayRules{
				ToRules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "192.168.0.1", Port: 8080}: {
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								TCP: &api.TCP{
									MaxConnectAttempt: pointer.To[uint32](21),
								},
							},
						},
					},
				},
			},
			gateways:      []*core_mesh.MeshGatewayResource{samples.GatewayTCPResource()},
			gatewayRoutes: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayTCPRoute()},
		}),
	)
})

func getResourceYaml(list core_xds.ResourceList) []byte {
	if len(list) == 0 {
		return []byte{}
	}
	actualListener, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualListener
}

func httpListener(port uint32) envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", port, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(HttpConnectionManager(fmt.Sprintf("outbound:127.0.0.1:%d", port), false)).
			Configure(HttpOutboundRoute(
				"backend",
				envoy_common.Routes{{
					Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
						envoy_common.WithService("backend"),
						envoy_common.WithWeight(100),
					)},
				}},
				map[string]map[string]bool{
					"kuma.io/service": {
						"web": true,
					},
				},
			)))).
		MustBuild()
}

func tcpListener(port uint32) envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", port, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(TcpProxyDeprecated(
				fmt.Sprintf("outbound:127.0.0.1:%d", port),
				envoy_common.NewCluster(
					envoy_common.WithService("backend"),
					envoy_common.WithWeight(100),
				),
			)),
		)).
		MustBuild()
}
