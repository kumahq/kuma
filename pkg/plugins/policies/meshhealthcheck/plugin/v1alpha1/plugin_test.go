//nolint:staticcheck // SA1019 Test file: tests backward compatibility with deprecated core_rules.Rule
package v1alpha1_test

import (
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhealthcheck/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test"
	test_matchers "github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v3/pkg/test/xds/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var _ = Describe("MeshHealthCheck", func() {
	httpServiceTag := "echo-http"
	splitHttpServiceTag := "echo-http-_0_"
	tcpServiceTag := "echo-tcp"
	grpcServiceTag := "echo-grpc"

	backendMeshServiceIdentifier := kri.Identifier{
		ResourceType: "MeshService",
		Mesh:         "default",
		Name:         "backend",
		SectionName:  "",
	}

	type testCase struct {
		resources        []core_xds.Resource
		toRules          core_rules.ToRules
		expectedClusters []string
	}
	httpClusters := []core_xds.Resource{
		{
			Name:   "cluster-echo-http",
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, httpServiceTag).
				MustBuild(),
		},
		{
			Name:   "cluster-echo-http-_0_",
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, splitHttpServiceTag).
				MustBuild(),
		},
	}
	tcpCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-tcp",
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, tcpServiceTag).
				MustBuild(),
		},
	}
	grpcCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-grpc",
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, grpcServiceTag).
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

			context := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(xds_context.NewResources()).
				AddServiceProtocol(httpServiceTag, core_meta.ProtocolHTTP).
				AddServiceProtocol(tcpServiceTag, core_meta.ProtocolTCP).
				AddServiceProtocol(grpcServiceTag, core_meta.ProtocolGRPC).
				AddServiceProtocol(splitHttpServiceTag, core_meta.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithOutbounds(xds_types.Outbounds{
					{LegacyOutbound: builders.Outbound().WithAddress("127.0.0.1").WithPort(27777).WithTags(map[string]string{
						mesh_proto.ServiceTag:  httpServiceTag,
						mesh_proto.ProtocolTag: "http",
					}).Build()},
					{LegacyOutbound: builders.Outbound().WithAddress("127.0.0.1").WithPort(27778).WithTags(map[string]string{
						mesh_proto.ServiceTag:  tcpServiceTag,
						mesh_proto.ProtocolTag: "tcp",
					}).Build()},
					{LegacyOutbound: builders.Outbound().WithAddress("127.0.0.1").WithPort(27779).WithTags(map[string]string{
						mesh_proto.ServiceTag:  grpcServiceTag,
						mesh_proto.ProtocolTag: "grpc",
					}).Build()},
					{LegacyOutbound: builders.Outbound().WithAddress("240.0.0.1").WithPort(27779).WithTags(map[string]string{
						mesh_proto.ServiceTag:  grpcServiceTag,
						mesh_proto.ProtocolTag: "grpc",
					}).Build()},
					{LegacyOutbound: builders.Outbound().WithAddress("127.0.0.1").WithPort(27780).WithTags(map[string]string{
						mesh_proto.ServiceTag:  splitHttpServiceTag,
						mesh_proto.ProtocolTag: "http",
					}).Build()},
				}).
				WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshHealthCheckType, given.toRules)).
				WithRouting(
					xds_builders.Routing().
						WithOutboundTargets(
							xds_builders.EndpointMap().
								AddEndpoint(httpServiceTag, xds_samples.HttpEndpointBuilder()).
								AddEndpoint(splitHttpServiceTag, xds_samples.HttpEndpointBuilder()).
								AddEndpoint(grpcServiceTag, xds_samples.GrpcEndpointBuilder()).
								AddEndpoint(tcpServiceTag, xds_samples.TcpEndpointBuilder()),
						),
				).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, proxy)).To(Succeed())

			for idx, expected := range given.expectedClusters {
				Expect(util_proto.ToYAML(resources.ListOf(envoy_resource.ClusterType)[idx].Resource)).
					To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("HTTP HealthCheck", testCase{
			resources: httpClusters,
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Interval:                     test.ParseDuration("10s"),
							UnhealthyInterval:            test.ParseDuration("5s"),
							Timeout:                      test.ParseDuration("2s"),
							UnhealthyThreshold:           pointer.To[int32](3),
							HealthyThreshold:             pointer.To[int32](1),
							InitialJitter:                test.ParseDuration("13s"),
							IntervalJitter:               test.ParseDuration("15s"),
							IntervalJitterPercent:        pointer.To[int32](10),
							HealthyPanicThreshold:        pointer.To(intstr.FromString("62.9")),
							FailTrafficOnPanic:           pointer.To(true),
							EventLogPath:                 pointer.To("/tmp/log.txt"),
							AlwaysLogHealthCheckFailures: pointer.To(false),
							NoTrafficInterval:            test.ParseDuration("16s"),
							Http: &api.HttpHealthCheck{
								Disabled: pointer.To(false),
								Path:     pointer.To("/health"),
								RequestHeadersToAdd: &api.HeaderModifier{
									Add: &[]api.HeaderKeyValue{
										{
											Name:  "x-some-header",
											Value: "value",
										},
									},
									Set: &[]api.HeaderKeyValue{
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
				},
			},
			expectedClusters: []string{
				"basic_http_health_check_cluster.golden.yaml",
				"basic_http_health_check_split_cluster.golden.yaml",
			},
		}),
		Entry("TCP HealthCheck", testCase{
			resources: tcpCluster,
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
				},
			},
			expectedClusters: []string{"basic_tcp_health_check_cluster.golden.yaml"},
		}),

		Entry("gRPC HealthCheck", testCase{
			resources: grpcCluster,
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
				},
			},
			expectedClusters: []string{"basic_grpc_health_check_cluster.golden.yaml"},
		}),
		Entry("TCP HealthCheck to real MeshService", testCase{
			resources: tcpCluster,
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					backendMeshServiceIdentifier: {
						Conf: []any{
							api.Conf{
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
					},
				},
			},
			expectedClusters: []string{"basic_tcp_health_check_cluster_real_ms.golden.yaml"},
		}),
	)

	type zoneProxyTestCase struct {
		resources       []*core_xds.Resource
		toRules         core_rules.ToRules
		expectedCluster string
	}

	mesMHCKRI := kri.Identifier{
		ResourceType: meshexternalservice_api.MeshExternalServiceType,
		Mesh:         "default",
		Name:         "example",
		SectionName:  "9000",
	}

	DescribeTable("should generate proper Envoy config for mesh-scoped zone proxy",
		func(given zoneProxyTestCase) {
			resourceSet := core_xds.NewResourceSet()
			for _, r := range given.resources {
				resourceSet.Add(r)
			}

			mes := builders.MeshExternalService().Build()

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithMeshLocalResources([]core_model.Resource{mes}).
				Build()

			proxy := xds_builders.Proxy().
				With(func(p *core_xds.Proxy) {
					p.Dataplane = &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{Name: "zone-proxy", Mesh: "default"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "10.0.0.1",
								Listeners: []*mesh_proto.Dataplane_Networking_Listener{{
									Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
									Address: "10.0.0.1",
									Port:    10002,
									Name:    "ze-port",
									State:   mesh_proto.Dataplane_Networking_Listener_Ready,
								}},
							},
						},
					}
				}).
				WithPolicies(
					xds_builders.MatchedPolicies().WithToPolicy(api.MeshHealthCheckType, given.toRules),
				).
				Build()

			p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(p.Apply(resourceSet, xdsCtx, proxy)).To(Succeed())

			Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[0].Resource)).
				To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", given.expectedCluster)))
		},
		Entry("MeshExternalService cluster with TCP health check on zone proxy", zoneProxyTestCase{
			resources: []*core_xds.Resource{{
				Name:   mesMHCKRI.String(),
				Origin: metadata.OriginEgress,
				Resource: clusters.NewClusterBuilder(envoy_common.APIV3, mesMHCKRI.String()).
					MustBuild(),
				ResourceOrigin: mesMHCKRI,
				Protocol:       core_meta.ProtocolTCP,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: outbound.ResourceRules{
					mesMHCKRI: {
						Conf: []any{api.Conf{
							Interval:           test.ParseDuration("10s"),
							Timeout:            test.ParseDuration("2s"),
							UnhealthyThreshold: pointer.To[int32](3),
							HealthyThreshold:   pointer.To[int32](1),
							Tcp: &api.TcpHealthCheck{
								Disabled: pointer.To(false),
							},
						}},
					},
				},
			},
			expectedCluster: "basic-meshexternalservice-tcp.zone_proxy_cluster.golden.yaml",
		}),
	)
})
