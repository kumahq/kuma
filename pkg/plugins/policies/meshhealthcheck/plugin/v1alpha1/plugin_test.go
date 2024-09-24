package v1alpha1_test

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	meshtcproute_plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
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
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

var _ = Describe("MeshHealthCheck", func() {
	httpServiceTag := "echo-http"
	splitHttpServiceTag := "echo-http-_0_"
	tcpServiceTag := "echo-tcp"
	grpcServiceTag := "echo-grpc"

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

	backendMeshExternalServiceIdentifier := func(mesh string) *core_model.TypedResourceIdentifier {
		return &core_model.TypedResourceIdentifier{
			ResourceIdentifier: core_model.ResourceIdentifier{
				Name:      "external",
				Mesh:      mesh,
				Namespace: "",
				Zone:      "",
			},
			ResourceType: "MeshExternalService",
		}
	}

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
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(xds_context.NewResources()).
				AddServiceProtocol(httpServiceTag, core_mesh.ProtocolHTTP).
				AddServiceProtocol(tcpServiceTag, core_mesh.ProtocolTCP).
				AddServiceProtocol(grpcServiceTag, core_mesh.ProtocolGRPC).
				AddServiceProtocol(splitHttpServiceTag, core_mesh.ProtocolHTTP).
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
		Entry("TCP HealthCheck to real MeshService", testCase{
			resources: tcpCluster,
			toRules: core_rules.ToRules{
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
					backendMeshServiceIdentifier: {
						Conf: []interface{}{
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

	XIt("should generate correct configuration for MeshExternalService with ZoneEgress", func() {
		// given
		rs := core_xds.NewResourceSet()
		rs.Add(&core_xds.Resource{
			Name:           "external-default",
			Origin:         egress.OriginEgress,
			Resource:       clusters.NewClusterBuilder(envoy_common.APIV3, "external").MustBuild(),
			ResourceOrigin: backendMeshExternalServiceIdentifier("default"),
			Protocol:       core_mesh.ProtocolTCP,
		})
		rs.Add(&core_xds.Resource{
			Name:           "external-mesh2",
			Origin:         egress.OriginEgress,
			Resource:       clusters.NewClusterBuilder(envoy_common.APIV3, "external").MustBuild(),
			ResourceOrigin: backendMeshExternalServiceIdentifier("mesh2"),
			Protocol:       core_mesh.ProtocolTCP,
		})

		proxy := &core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
			ZoneEgressProxy: &core_xds.ZoneEgressProxy{
				ZoneEgressResource: &core_mesh.ZoneEgressResource{
					Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "default"},
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
						Mesh: builders.Mesh().WithName("default").WithEnabledMTLSBackend("ca-1").WithBuiltinMTLSBackend("ca-1").Build(),
						Resources: map[core_model.ResourceType]core_model.ResourceList{
							meshexternalservice_api.MeshExternalServiceType: &meshexternalservice_api.MeshExternalServiceResourceList{
								Items: []*meshexternalservice_api.MeshExternalServiceResource{
									samples.MeshExternalServiceExampleBuilder().WithMesh("default").WithName("external").Build(),
								},
							},
						},
						Dynamic: core_xds.ExternalServiceDynamicPolicies{
							"default_external___extsvc_9000": {
								api.MeshHealthCheckType: core_xds.TypedMatchingPolicies{
									ToRules: core_rules.ToRules{
										ResourceRules: core_rules.ResourceRules{
											*backendMeshExternalServiceIdentifier("default"): {
												Conf: []interface{}{
													api.Conf{
														Interval:           test.ParseDuration("11s"),
														Timeout:            test.ParseDuration("99s"),
														UnhealthyThreshold: pointer.To[int32](7),
														HealthyThreshold:   pointer.To[int32](77),
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
								},
							},
						},
					},
					{
						Mesh: builders.Mesh().WithName("mesh2").WithEnabledMTLSBackend("ca-2").WithBuiltinMTLSBackend("ca-2").Build(),
						Resources: map[core_model.ResourceType]core_model.ResourceList{
							meshexternalservice_api.MeshExternalServiceType: &meshexternalservice_api.MeshExternalServiceResourceList{
								Items: []*meshexternalservice_api.MeshExternalServiceResource{
									samples.MeshExternalServiceExampleBuilder().WithName("external").WithMesh("mesh-2").Build(),
								},
							},
						},
						Dynamic: core_xds.ExternalServiceDynamicPolicies{
							"mesh-2_external___extsvc_9000": {
								api.MeshHealthCheckType: core_xds.TypedMatchingPolicies{
									ToRules: core_rules.ToRules{
										ResourceRules: core_rules.ResourceRules{
											*backendMeshExternalServiceIdentifier("mesh2"): {
												Conf: []interface{}{
													api.Conf{
														Interval:           test.ParseDuration("55s"),
														Timeout:            test.ParseDuration("5s"),
														UnhealthyThreshold: pointer.To[int32](122),
														HealthyThreshold:   pointer.To[int32](311),
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
								},
							},
						},
					},
				},
			},
		}

		// when
		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		err := p.Apply(rs, xds_context.Context{}, proxy)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(getResource(rs, envoy_resource.ClusterType)).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "basic_egress_meshexternalservice_cluster.golden.yaml")))
	})

	type gatewayTestCase struct {
		name           string
		gatewayRoutes  []*core_mesh.MeshGatewayRouteResource
		meshhttproutes core_rules.GatewayRules
		meshtcproutes  core_rules.GatewayRules
		meshservices   []*meshservice_api.MeshServiceResource
		rules          core_rules.GatewayRules
	}
	DescribeTable("should generate proper Envoy config for MeshGateways",
		func(given gatewayTestCase) {
			Expect(given.name).ToNot(BeEmpty())
			resources := xds_context.NewResources()

			gateway := &core_mesh.MeshGatewayResource{
				Meta: &test_model.ResourceMeta{Name: "sample-gateway", Mesh: "default"},
				Spec: &mesh_proto.MeshGateway{
					Selectors: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								mesh_proto.ServiceTag: "sample-gateway",
							},
						},
					},
					Conf: &mesh_proto.MeshGateway_Conf{
						Listeners: []*mesh_proto.MeshGateway_Listener{
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8080,
								Tags: map[string]string{
									"protocol": "http",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_TCP,
								Port:     8081,
								Tags: map[string]string{
									"protocol": "tcp",
								},
							},
						},
					},
				},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{gateway},
			}
			if len(given.gatewayRoutes) > 0 {
				resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
					Items: given.gatewayRoutes,
				}
			}
			if len(given.meshservices) > 0 {
				resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
					Items: given.meshservices,
				}
			}

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().
					WithGatewayPolicy(api.MeshHealthCheckType, given.rules).
					WithGatewayPolicy(meshhttproute_api.MeshHTTPRouteType, given.meshhttproutes).
					WithGatewayPolicy(meshtcproute_api.MeshTCPRouteType, given.meshtcproutes)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			httpRoutePlugin := meshhttproute_plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(httpRoutePlugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			tcpRoutePlugin := meshtcproute_plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(tcpRoutePlugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			getResourcesYaml := func(list core_xds.ResourceList) []byte {
				resources, err := list.ToDeltaDiscoveryResponse()
				Expect(err).ToNot(HaveOccurred())
				actual, err := util_proto.ToYAML(resources)
				Expect(err).ToNot(HaveOccurred())

				return actual
			}

			// then
			Expect(getResourcesYaml(generatedResources.ListOf(envoy_resource.ClusterType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_cluster.golden.yaml", given.name))))
		},
		Entry("basic outbound cluster with HTTP health check", gatewayTestCase{
			name:          "basic",
			gatewayRoutes: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
			rules: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
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
			},
		}),
		Entry("HTTP HealthCheck to real MeshService", gatewayTestCase{
			name: "real-mesh-service-meshhttproute",
			meshservices: []*meshservice_api.MeshServiceResource{
				{
					Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
					Spec: &meshservice_api.MeshService{
						Selector: meshservice_api.Selector{},
						Ports: []meshservice_api.Port{{
							Port:        80,
							TargetPort:  intstr.FromInt(8084),
							AppProtocol: core_mesh.ProtocolHTTP,
						}},
						Identities: []meshservice_api.MeshServiceIdentity{
							{
								Type:  meshservice_api.MeshServiceIdentityServiceTagType,
								Value: "backend",
							},
							{
								Type:  meshservice_api.MeshServiceIdentityServiceTagType,
								Value: "other-backend",
							},
						},
					},
					Status: &meshservice_api.MeshServiceStatus{
						VIPs: []meshservice_api.VIP{{
							IP: "10.0.0.1",
						}},
					},
				},
			},
			meshhttproutes: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
							{
								Subset: core_rules.MeshSubset(),
								Conf: meshhttproute_api.PolicyDefault{
									Rules: []meshhttproute_api.Rule{{
										Matches: []meshhttproute_api.Match{{
											Path: &meshhttproute_api.PathMatch{
												Type:  meshhttproute_api.Exact,
												Value: "/",
											},
										}},
										Default: meshhttproute_api.RuleConf{
											BackendRefs: &[]common_api.BackendRef{{
												TargetRef: builders.TargetRefService("backend"),
												Port:      pointer.To(uint32(80)),
												Weight:    pointer.To(uint(100)),
											}},
										},
									}},
								},
								Origin: []core_model.ResourceMeta{
									&test_model.ResourceMeta{Mesh: "default", Name: "http-route"},
								},
								BackendRefOriginIndex: core_rules.BackendRefOriginIndex{
									meshhttproute_api.HashMatches([]meshhttproute_api.Match{{Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.Exact, Value: "/"}}}): 0,
								},
							},
						},
					},
				},
			},
			rules: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
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
			},
		}),
		Entry("basic outbound cluster with TCP/HTTP health check and MeshTCPRoute/MeshHTTPRoute", gatewayTestCase{
			name: "basic-meshhttproute",
			meshhttproutes: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
							{
								Subset: core_rules.MeshSubset(),
								Conf: meshhttproute_api.PolicyDefault{
									Rules: []meshhttproute_api.Rule{{
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
									}},
								},
							},
						},
					},
				},
			},
			meshtcproutes: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						rules.NewInboundListenerHostname("192.168.0.1", 8081, "*"): {
							{
								Subset: core_rules.MeshSubset(),
								Conf: meshtcproute_api.Rule{
									Default: meshtcproute_api.RuleConf{
										BackendRefs: []common_api.BackendRef{{
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
			rules: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.Rules{
						rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
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
						rules.NewInboundListenerHostname("192.168.0.1", 8081, "*"): {
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
									HealthyPanicThreshold:        pointer.To(intstr.FromString("72.9")),
									FailTrafficOnPanic:           pointer.To(true),
									EventLogPath:                 pointer.To("/tmp/log.txt"),
									AlwaysLogHealthCheckFailures: pointer.To(false),
									NoTrafficInterval:            test.ParseDuration("16s"),
									ReuseConnection:              pointer.To(true),
								},
							},
						},
					},
				},
			},
		}),
	)
})

func getResource(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}
