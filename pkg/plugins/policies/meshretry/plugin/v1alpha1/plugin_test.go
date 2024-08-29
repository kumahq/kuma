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
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
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
	backendMeshServiceIdentifier := core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name:      "backend",
			Mesh:      "default",
			Namespace: "backend-ns",
			Zone:      "zone-1",
		},
		ResourceType: "MeshService",
		SectionName:  "",
	}

	backendMeshExternalServiceIdentifier := core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name:      "backend",
			Mesh:      "default",
			Namespace: "kuma-system",
			Zone:      "zone-1",
		},
		ResourceType: "MeshExternalService",
		SectionName:  "",
	}

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
			WithMeshBuilder(samples.MeshDefaultBuilder()).
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
				WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http")).
			WithOutbounds(xds_types.Outbounds{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Port: builders.FirstOutboundPort,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "http-service",
					},
				}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Port: builders.FirstOutboundPort + 1,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "grpc-service",
					},
				}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Port: builders.FirstOutboundPort + 2,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "tcp-service",
					},
				}},
			}).
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
				Resource: httpListenerWithSimpleRoute(10001),
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
		Entry("http retry 0 numRetries", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListenerWithSimpleRoute(10001),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							HTTP: &api.HTTP{
								NumRetries:    pointer.To[uint32](0),
								PerTryTimeout: test.ParseDuration("2s"),
							},
						},
					},
				},
			},
			goldenFilePrefix: "http_0_numretries",
		}),
		Entry("grpc retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListenerWithSimpleRoute(10002),
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
		Entry("grpc retry 0 numRetries", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListenerWithSimpleRoute(10002),
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
								NumRetries:    pointer.To[uint32](0),
								PerTryTimeout: test.ParseDuration("12s"),
							},
						},
					},
				},
			},
			goldenFilePrefix: "grpc_0_numretries",
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
		Entry("retry per http route", testCase{
			resources: []core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: httpListenerWithSeveralRoutes(10001),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "http-service",
							},
							{
								Key:   core_rules.RuleMatchesHashTag,
								Value: "9Zuf5Tg79OuZcQITwBbQykxAk2u4fRKrwYn3//AL4Yo=", // '[{"path":{"value":"/","type":"PathPrefix"}}]'
							},
						},
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
					{
						Subset: core_rules.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "http-service",
							},
							{
								Key:   core_rules.RuleMatchesHashTag,
								Value: "U8NGexJyQPtOd+lzwvsjLMysuDL6MmTJPSRX4C43niU=", // '[{"path":{"value":"/another-backend","type":"Exact"}},{"method":"GET"}]'
							},
						},
						Conf: api.Conf{
							HTTP: &api.HTTP{
								NumRetries:    pointer.To[uint32](6),
								PerTryTimeout: test.ParseDuration("77s"),
								BackOff: &api.BackOff{
									BaseInterval: test.ParseDuration("88s"),
									MaxInterval:  test.ParseDuration("999s"),
								},
								RateLimitedBackOff: &api.RateLimitedBackOff{
									MaxInterval: test.ParseDuration("11s"),
									ResetHeaders: &[]api.ResetHeader{
										{
											Name:   "x-retry-after-http",
											Format: "UnixTimestamp",
										},
									},
								},
								RetryOn: &[]api.HTTPRetryOn{
									"499",
								},
								RetriableResponseHeaders: &[]common_api.HeaderMatch{
									{
										Type:  pointer.To(common_api.HeaderMatchRegularExpression),
										Name:  "x-retry-regex",
										Value: ".*",
									},
								},
								RetriableRequestHeaders: &[]common_api.HeaderMatch{
									{
										Type:  pointer.To(common_api.HeaderMatchPrefix),
										Name:  "x-retry-prefix",
										Value: "prefix-another",
									},
								},
								HostSelection: &[]api.Predicate{
									{
										PredicateType: "OmitPreviousHosts",
									},
									{
										PredicateType:   "OmitPreviousPriorities",
										UpdateFrequency: 5,
									},
									{
										PredicateType: "OmitHostsWithTags",
										Tags: map[string]string{
											"another-test": "another-test",
										},
									},
								},
								HostSelectionMaxAttempts: pointer.To(int64(99)),
							},
						},
					},
				},
			},
			goldenFilePrefix: "retry_per_http_route",
		}),
		Entry("http retry", testCase{
			resources: []core_xds.Resource{{
				Name:           "outbound",
				Origin:         generator.OriginOutbound,
				Resource:       httpListenerWithSimpleRoute(10001),
				ResourceOrigin: &backendMeshServiceIdentifier,
				Protocol:       core_mesh.ProtocolHTTP,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
					backendMeshServiceIdentifier: {
						Conf: []interface{}{
							api.Conf{
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
			goldenFilePrefix: "http-real-mesh-service",
		}),
		Entry("http retry mesh external service", testCase{
			resources: []core_xds.Resource{{
				Name:           "outbound",
				Origin:         generator.OriginOutbound,
				Resource:       httpListenerWithSimpleRoute(10001),
				ResourceOrigin: &backendMeshExternalServiceIdentifier,
				Protocol:       core_mesh.ProtocolHTTP,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
					backendMeshExternalServiceIdentifier: {
						Conf: []interface{}{
							api.Conf{
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
			goldenFilePrefix: "http-real-mesh-external-service",
		}),
	)

	type gatewayTestCase struct {
		rules            core_rules.GatewayRules
		goldenFilePrefix string
		gateways         []*core_mesh.MeshGatewayResource
		gatewayRoutes    []*core_mesh.MeshGatewayRouteResource
		meshhttproutes   core_rules.GatewayRules
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
				WithPolicies(xds_builders.MatchedPolicies().
					WithGatewayPolicy(api.MeshRetryType, given.rules).
					WithGatewayPolicy(meshhttproute_api.MeshHTTPRouteType, given.meshhttproutes)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			httpRoutePlugin := meshhttproute_plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(httpRoutePlugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

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
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.Rules{
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
			},
			gateways:      []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			gatewayRoutes: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
		}),
		Entry("tcp retry", gatewayTestCase{
			goldenFilePrefix: "gateway.tcp",
			rules: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.Rules{
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
			},
			gateways:      []*core_mesh.MeshGatewayResource{samples.GatewayTCPResource()},
			gatewayRoutes: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayTCPRoute()},
		}),
		Entry("http retry with MeshHTTPRoute", gatewayTestCase{
			goldenFilePrefix: "per-route-configuration.gateway.http",
			rules: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.Rules{
						{Address: "192.168.0.1", Port: 8080}: {
							{
								Subset: core_rules.Subset{
									{
										Key:   core_rules.RuleMatchesHashTag,
										Value: "L2t9uuHxXPXUg5ULwRirUaoxN4BU/zlqyPK8peSWm2g=",
									},
								},
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
			},
			gateways: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			meshhttproutes: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
							{
								Subset: core_rules.MeshSubset(),
								Conf: meshhttproute_api.PolicyDefault{
									Rules: []meshhttproute_api.Rule{
										{
											Matches: []meshhttproute_api.Match{{
												Path: &meshhttproute_api.PathMatch{
													Type:  meshhttproute_api.Exact,
													Value: "/",
												},
											}},
											Default: meshhttproute_api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										},
										{
											Matches: []meshhttproute_api.Match{{
												Path: &meshhttproute_api.PathMatch{
													Type:  meshhttproute_api.Exact,
													Value: "/another-route",
												},
												Method: pointer.To[meshhttproute_api.Method]("GET"),
											}},
											Default: meshhttproute_api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
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

func httpListenerWithSeveralRoutes(port uint32) envoy_common.NamedResource {
	return httpListener(port, AddFilterChainConfigurer(samples.MeshHttpOutboundWithSeveralRoutes("http-service")))
}

func httpListenerWithSimpleRoute(port uint32) envoy_common.NamedResource {
	return httpListener(port, AddFilterChainConfigurer(samples.MeshHttpOutboudWithSingleRoute("backend")))
}

func httpListener(port uint32, route FilterChainBuilderOpt) envoy_common.NamedResource {
	return NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", port, core_xds.SocketAddressProtocolTCP).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(HttpConnectionManager(fmt.Sprintf("outbound:127.0.0.1:%d", port), false)).
			Configure(route))).
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
