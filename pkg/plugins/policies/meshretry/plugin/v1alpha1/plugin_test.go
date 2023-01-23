package v1alpha1

import (
	"fmt"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshRetry", func() {
	type testCase struct {
		resources        []core_xds.Resource
		toRules          core_xds.ToRules
		expectedListener string
	}
	DescribeTable("should generate proper Envoy config", func(given testCase) {
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
				AddOutboundsToServices("http-service", "grpc-service", "tcp-service").
				WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http").
				Build(),
			Policies: xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
					api.MeshRetryType: {
						Type:    api.MeshRetryType,
						ToRules: given.toRules,
					},
				},
			},
			Routing: core_xds.Routing{
				OutboundTargets: core_xds.EndpointMap{
					"http-service": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "http",
						},
					}},
					"tcp-service": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "tcp",
						},
					}},
					"grpc-service": []core_xds.Endpoint{{
						Tags: map[string]string{
							"kuma.io/protocol": "grpc",
						},
					}},
				},
			},
		}

		// when
		plugin := NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())

		// then
		Expect(getResourceYaml(resourceSet.ListOf(envoy_resource.ListenerType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", given.expectedListener)))
	},
		Entry("http retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListener(10001),
			}},
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
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
								RetriableResponseHeaders: &[]api.HTTPHeaderMatch{
									{
										Type:  pointer.To(api.HeaderMatchRegularExpression),
										Name:  "x-retry-regex",
										Value: ".*",
									},
									{
										Type:  pointer.To(api.HeaderMatchExact),
										Name:  "x-retry-exact",
										Value: "exact-value",
									},
								},
								RetriableRequestHeaders: &[]api.HTTPHeaderMatch{
									{
										Type:  pointer.To(api.HeaderMatchPrefix),
										Name:  "x-retry-prefix",
										Value: "prefix-",
									},
								},
							},
						},
					},
				},
			},
			expectedListener: "http_retry_listener.golden.yaml",
		}),
		Entry("grpc retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpListener(10002),
			}},
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
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
			expectedListener: "grpc_retry_listener.golden.yaml",
		}),
		Entry("tcp retry", testCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: tcpListener(10003),
			}},
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							TCP: &api.TCP{
								MaxConnectAttempt: pointer.To[uint32](21),
							},
						},
					},
				},
			},
			expectedListener: "tcp_retry_listener.golden.yaml",
		}),
	)
})

func getResourceYaml(list core_xds.ResourceList) []byte {
	actualListener, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualListener
}

func httpListener(port uint32) envoy_common.NamedResource {
	return NewListenerBuilder(envoy_common.APIV3).
		Configure(OutboundListener(fmt.Sprintf("outbound:127.0.0.1:%d", port), "127.0.0.1", port, core_xds.SocketAddressProtocolTCP)).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
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
	return NewListenerBuilder(envoy_common.APIV3).
		Configure(OutboundListener(fmt.Sprintf("outbound:127.0.0.1:%d", port), "127.0.0.1", port, core_xds.SocketAddressProtocolTCP)).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
			Configure(TcpProxy(
				fmt.Sprintf("outbound:127.0.0.1:%d", port),
				envoy_common.NewCluster(
					envoy_common.WithService("backend"),
					envoy_common.WithWeight(100),
				),
			)),
		)).
		MustBuild()
}
