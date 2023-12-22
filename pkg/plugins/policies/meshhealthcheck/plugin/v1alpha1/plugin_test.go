package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
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
		toRules          core_rules.ToRules
		expectedClusters []string
	}
	httpClusters := []core_xds.Resource{
		{
			Name:   "cluster-echo-http",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, httpServiceTag).
				MustBuild(),
		},
		{
			Name:   "cluster-echo-http-_0_",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, splitHttpServiceTag).
				MustBuild(),
		},
	}
	tcpCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-tcp",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, tcpServiceTag).
				MustBuild(),
		},
	}
	grpcCluster := []core_xds.Resource{
		{
			Name:   "cluster-echo-grpc",
			Origin: generator.OriginOutbound,
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
				WithMesh(samples.MeshDefaultBuilder()).
				WithResources(xds_context.NewResources()).
				AddServiceProtocol(httpServiceTag, core_mesh.ProtocolHTTP).
				AddServiceProtocol(tcpServiceTag, core_mesh.ProtocolTCP).
				AddServiceProtocol(grpcServiceTag, core_mesh.ProtocolGRPC).
				AddServiceProtocol(splitHttpServiceTag, core_mesh.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(
					samples.DataplaneBackendBuilder().
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
						),
				).
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
						Subset: core_rules.Subset{},
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
						Subset: core_rules.Subset{},
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
						Subset: core_rules.Subset{},
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
	)

	type gatewayTestCase struct {
		name  string
		rules core_rules.GatewayRules
	}
	DescribeTable("should generate proper Envoy config for MeshGateways",
		func(given gatewayTestCase) {
			Expect(given.name).ToNot(BeEmpty())
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
			}

			xdsCtx := *xds_builders.Context().
				WithMesh(samples.MeshDefaultBuilder()).
				WithResources(resources).
				AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshHealthCheckType, given.rules)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			getResourceYaml := func(list core_xds.ResourceList) []byte {
				actualResource, err := util_proto.ToYAML(list[0].Resource)
				Expect(err).ToNot(HaveOccurred())
				return actualResource
			}

			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_cluster.golden.yaml", given.name))))
		},
		Entry("basic outbound cluster with HTTP health check", gatewayTestCase{
			name: "basic",
			rules: core_rules.GatewayRules{
				ToRules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "192.168.0.1", Port: 8080}: {
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								Interval:                     test.ParseDuration("10s"),
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
					},
				},
			},
		}),
	)
})
