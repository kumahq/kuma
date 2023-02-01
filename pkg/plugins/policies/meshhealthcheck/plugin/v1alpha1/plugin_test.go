package v1alpha1_test

import (
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	"github.com/kumahq/kuma/pkg/test"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshHealthCheck", func() {
	httpServiceTag := "echo-http"
	splitHttpServiceTag := "echo-http-_0_"
	tcpServiceTag := "echo-tcp"
	grpcServiceTag := "echo-grpc"
	type testCase struct {
		resources        []core_xds.Resource
		toRules          core_xds.ToRules
		expectedClusters []string
	}
	httpClusters := []core_xds.Resource{
		{
			Name:   "cluster-echo-http",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
				Configure(policies_xds.WithName(httpServiceTag)).
				MustBuild(),
		},
		{
			Name:   "cluster-echo-http-_0_",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
				Configure(policies_xds.WithName(splitHttpServiceTag)).
				MustBuild(),
		},
	}
	tcpCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-tcp",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
				Configure(policies_xds.WithName(tcpServiceTag)).
				MustBuild(),
		},
	}
	grpcCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-grpc",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
				Configure(policies_xds.WithName(grpcServiceTag)).
				MustBuild(),
		},
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			context := xds_context.Context{}
			proxy := xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Dataplane: samples.DataplaneBackendBuilder().
					AddOutbound(
						builders.Outbound().WithAddress("127.0.0.1").WithPort(27777).WithTags(map[string]string{
							mesh_proto.ServiceTag:  httpServiceTag,
							mesh_proto.ProtocolTag: "http",
						}),
					).
					AddOutbound(
						builders.Outbound().WithAddress("127.0.0.1").WithPort(27778).WithTags(map[string]string{
							mesh_proto.ServiceTag:  tcpServiceTag,
							mesh_proto.ProtocolTag: "tcp",
						}),
					).
					AddOutbound(
						builders.Outbound().WithAddress("127.0.0.1").WithPort(27779).WithTags(map[string]string{
							mesh_proto.ServiceTag:  grpcServiceTag,
							mesh_proto.ProtocolTag: "grpc",
						}),
					).
					AddOutbound(
						builders.Outbound().WithAddress("240.0.0.1").WithPort(27779).WithTags(map[string]string{
							mesh_proto.ServiceTag:  grpcServiceTag,
							mesh_proto.ProtocolTag: "grpc",
						}),
					).
					AddOutbound(
						builders.Outbound().WithAddress("127.0.0.1").WithPort(27780).WithTags(map[string]string{
							mesh_proto.ServiceTag:  splitHttpServiceTag,
							mesh_proto.ProtocolTag: "http",
						}),
					).
					Build(),
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshHealthCheckType: {
							Type:    api.MeshHealthCheckType,
							ToRules: given.toRules,
						},
					},
				},
				Routing: xds.Routing{
					OutboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
						httpServiceTag: {
							{
								Tags: map[string]string{mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP},
							},
						},
						splitHttpServiceTag: {
							{
								Tags: map[string]string{mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP},
							},
						},
						grpcServiceTag: {
							{
								Tags: map[string]string{mesh_proto.ProtocolTag: core_mesh.ProtocolGRPC},
							},
						},
						tcpServiceTag: {
							{
								Tags: map[string]string{mesh_proto.ProtocolTag: core_mesh.ProtocolTCP},
							},
						},
					},
				},
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, &proxy)).To(Succeed())

			for idx, expected := range given.expectedClusters {
				Expect(util_proto.ToYAML(resources.ListOf(envoy_resource.ClusterType)[idx].Resource)).
					To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("HTTP HealthCheck", testCase{
			resources: httpClusters,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval:                     test.ParseDuration("10s"),
							Timeout:                      test.ParseDuration("2s"),
							UnhealthyThreshold:           pointer.To[int32](3),
							HealthyThreshold:             pointer.To[int32](1),
							InitialJitter:                test.ParseDuration("13s"),
							IntervalJitter:               test.ParseDuration("15s"),
							IntervalJitterPercent:        pointer.To[int32](10),
							HealthyPanicThreshold:        pointer.To[intstr.IntOrString](intstr.FromString("62.9")),
							FailTrafficOnPanic:           pointer.To(true),
							EventLogPath:                 pointer.To("/tmp/log.txt"),
							AlwaysLogHealthCheckFailures: pointer.To(false),
							NoTrafficInterval:            test.ParseDuration("16s"),
							Http: &api.HttpHealthCheck{
								Disabled: pointer.To(false),
								Path:     pointer.To("/health"),
								RequestHeadersToAdd: &api.HeaderModifier{
									Add: []api.HeaderKeyValue{
										{
											Name:  "x-some-header",
											Value: "value",
										},
									},
									Set: []api.HeaderKeyValue{
										{
											Name:  "x-some-other-header",
											Value: "value",
										},
									},
								},
								ExpectedStatuses: &[]int32{200, 201},
							},
							ReuseConnection: pointer.To(true),
						},
					},
				}},
			expectedClusters: []string{
				"basic_http_health_check_cluster.golden.yaml",
				"basic_http_health_check_split_cluster.golden.yaml",
			},
		}),
		Entry("TCP HealthCheck", testCase{
			resources: tcpCluster,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval:           test.ParseDuration("10s"),
							Timeout:            test.ParseDuration("2s"),
							UnhealthyThreshold: pointer.To[int32](3),
							HealthyThreshold:   pointer.To[int32](1),
							Tcp: &api.TcpHealthCheck{
								Disabled: pointer.To(false),
								Send:     pointer.To("cGluZwo="),
								Receive:  &[]string{"cG9uZwo="},
							},
						},
					},
				}},
			expectedClusters: []string{"basic_tcp_health_check_cluster.golden.yaml"},
		}),

		Entry("gRPC HealthCheck", testCase{
			resources: grpcCluster,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval:           test.ParseDuration("10s"),
							Timeout:            test.ParseDuration("2s"),
							UnhealthyThreshold: pointer.To[int32](3),
							HealthyThreshold:   pointer.To[int32](1),
							Grpc: &api.GrpcHealthCheck{
								ServiceName: pointer.To("grpc-client"),
								Authority:   pointer.To("grpc-client.default.svc.cluster.local"),
							},
						},
					},
				}},
			expectedClusters: []string{"basic_grpc_health_check_cluster.golden.yaml"},
		}),
	)
})
