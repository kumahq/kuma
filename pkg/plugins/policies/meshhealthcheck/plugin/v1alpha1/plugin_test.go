package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshHealthCheck", func() {
	httpServiceTag := "echo-http"
	tcpServiceTag := "echo-tcp"
	grpcServiceTag := "echo-grpc"
	type testCase struct {
		resources        []core_xds.Resource
		toRules          core_xds.ToRules
		expectedClusters []string
	}
	httpCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-http",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
				Configure(policies_xds.WithName(httpServiceTag)).
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
			policies_xds.ResourceArrayShouldEqual(resources.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("HTTP HealthCheck", testCase{
			resources: httpCluster,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval:                     *test.ParseDuration("10s"),
							Timeout:                      *test.ParseDuration("2s"),
							UnhealthyThreshold:           3,
							HealthyThreshold:             1,
							InitialJitter:                test.ParseDuration("13s"),
							IntervalJitter:               test.ParseDuration("15s"),
							IntervalJitterPercent:        policies_xds.PointerOf[int32](10),
							HealthyPanicThreshold:        policies_xds.PointerOf[int32](11),
							FailTrafficOnPanic:           policies_xds.PointerOf[bool](true),
							EventLogPath:                 policies_xds.PointerOf[string]("/tmp/log.txt"),
							AlwaysLogHealthCheckFailures: policies_xds.PointerOf[bool](false),
							NoTrafficInterval:            test.ParseDuration("16s"),
							Http: &api.HttpHealthCheck{
								Disabled: false,
								Path:     "/health",
								RequestHeadersToAdd: &[]api.HeaderValueOption{
									{
										Header: &api.HeaderValue{
											Key:   "x-some-header",
											Value: "value",
										},
										Append: policies_xds.PointerOf[bool](true),
									},
								},
								ExpectedStatuses: &[]int32{200, 201},
							},
							ReuseConnection: policies_xds.PointerOf[bool](true),
						},
					},
				}},
			expectedClusters: []string{`
name: echo-http
commonLbConfig:
  healthyPanicThreshold:
    value: 11
  zoneAwareLbConfig:
    failTrafficOnPanic: true
healthChecks:
- eventLogPath: /tmp/log.txt
  healthyThreshold: 1
  httpHealthCheck:
    expectedStatuses:
      - end: "201"
        start: "200"
      - end: "202"
        start: "201"
    path: /health
    requestHeadersToAdd:
      - header:
          key: x-kuma-tags
          value: '&kuma.io/service=backend&'
      - append: true
        header:
          key: x-some-header
          value: value
  initialJitter: 13s
  interval: 10s
  intervalJitter: 15s
  intervalJitterPercent: 10
  noTrafficInterval: 16s
  reuseConnection: true
  timeout: 2s
  unhealthyThreshold: 3
`},
		}),
		Entry("TCP HealthCheck", testCase{
			resources: tcpCluster,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval:           *test.ParseDuration("10s"),
							Timeout:            *test.ParseDuration("2s"),
							UnhealthyThreshold: 3,
							HealthyThreshold:   1,
							Tcp: &api.TcpHealthCheck{
								Disabled: false,
								Send:     policies_xds.PointerOf[string]("cGluZwo="),
								Receive:  &[]string{"cG9uZwo="},
							},
						},
					},
				}},
			expectedClusters: []string{`
name: echo-tcp
healthChecks:
- healthyThreshold: 1
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
  tcpHealthCheck:
    send:
        text: "63476c755a776f3d"
    receive:
        - text: "634739755a776f3d"
`},
		}),

		Entry("gRPC HealthCheck", testCase{
			resources: grpcCluster,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval:           *test.ParseDuration("10s"),
							Timeout:            *test.ParseDuration("2s"),
							UnhealthyThreshold: 3,
							HealthyThreshold:   1,
							Grpc: &api.GrpcHealthCheck{
								ServiceName: "grpc-client",
								Authority:   "grpc-client.default.svc.cluster.local",
							},
						},
					},
				}},
			expectedClusters: []string{`
name: echo-grpc
healthChecks:
- healthyThreshold: 1
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
  grpcHealthCheck:
    authority: grpc-client.default.svc.cluster.local
    serviceName: grpc-client
`},
		}),
	)
})
