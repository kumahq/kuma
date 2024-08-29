package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
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

var _ = Describe("MeshAccessLog", func() {
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
			Name:      "example",
			Mesh:      "default",
			Namespace: "",
			Zone:      "",
		},
		ResourceType: "MeshExternalService",
	}

	type sidecarTestCase struct {
		resources         []core_xds.Resource
		outbounds         xds_types.Outbounds
		toRules           core_rules.ToRules
		fromRules         core_rules.FromRules
		expectedListeners []string
		expectedClusters  []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			xdsCtx := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithMetadata(&core_xds.DataplaneMetadata{
					WorkDir: "/tmp",
				}).
				WithDataplane(
					builders.Dataplane().
						WithName("backend").
						WithMesh("default").
						AddInbound(builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777),
						),
				).
				WithOutbounds(append(given.outbounds, &xds_types.Outbound{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				})).
				WithPolicies(
					xds_builders.MatchedPolicies().WithPolicy(api.MeshAccessLogType, given.toRules, given.fromRules),
				).
				Build()

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// then
			Expect(plugin.Apply(resourceSet, xdsCtx, proxy)).To(Succeed())
			for i, expectedListener := range given.expectedListeners {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[i].Resource)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", expectedListener)))
			}
			for i, expectedCluster := range given.expectedClusters {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[i].Resource)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", expectedCluster)))
			}
		},
		Entry("basic outbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
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
							),
						),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"basic_outbound.listener.golden.yaml"},
		}),
		Entry("basic outbound route from real MeshService", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
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
							),
						),
					)).MustBuild(),
				ResourceOrigin: &backendMeshServiceIdentifier,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
					backendMeshServiceIdentifier: {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/log",
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"basic_outbound_real_meshservice.listener.golden.yaml"},
		}),
		Entry("basic outbound route from real MeshExternalService", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
								"example",
								envoy_common.Routes{{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("example"),
										envoy_common.WithWeight(100),
									)},
								}},
								map[string]map[string]bool{
									"kuma.io/service": {
										"web": true,
									},
								},
							),
						),
					)).MustBuild(),
				ResourceOrigin: &backendMeshExternalServiceIdentifier,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
					backendMeshExternalServiceIdentifier: {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/log",
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"basic_outbound_real_meshexternalservice.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with file backend and default format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_file_backend_default_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with file backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
									Format: &api.Format{
										Plain: pointer.To("custom format [%START_TIME%] %RESPONSE_FLAGS%"),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_file_backend_plain_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with file backend and json format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
									Format: &api.Format{
										Json: pointer.To([]api.JsonValue{
											{Key: "protocol", Value: "%PROTOCOL%"},
											{Key: "duration", Value: "%DURATION%"},
										}),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_file_backend_json_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with tcp backend and default format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Tcp: &api.TCPBackend{
									Address: "logging.backend",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_tcp_backend_default_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with opentelemetry backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "other-service",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("other-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}, {
				Name:   "foo",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27778, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27778",
							envoy_common.NewCluster(
								envoy_common.WithService("foo-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}, {
				Name:   "bar",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27779, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27779",
							envoy_common.NewCluster(
								envoy_common.WithService("bar-service"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			outbounds: xds_types.Outbounds{
				{LegacyOutbound: builders.Outbound().
					WithService("foo-service").
					WithAddress("127.0.0.1").
					WithPort(27778).Build()},
				{LegacyOutbound: builders.Outbound().
					WithService("bar-service").
					WithAddress("127.0.0.1").
					WithPort(27779).Build()},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OtelBackend{
									Endpoint: "otel-collector",
								},
							}},
						},
					},
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "foo-service",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OtelBackend{
									Endpoint: "otel-collector",
									Body: &apiextensionsv1.JSON{
										Raw: []byte("%KUMA_MESH%"),
									},
								},
							}},
						},
					},
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "bar-service",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OtelBackend{
									Endpoint: "other-otel-collector:5317",
									Body: &apiextensionsv1.JSON{
										Raw: []byte(`{
										  "kvlistValue": {
											"values": [
											  {"key": "mesh", "value": {"stringValue": "%KUMA_MESH%"}}
											]
										  }
									    }`),
									},
								},
							}},
						},
					},
				},
			},
			expectedClusters: []string{
				"outbound_otel_backend_plain_format.cluster.golden.yaml",
				"outbound_otel_backend_plain_format_1.cluster.golden.yaml",
			},
			expectedListeners: []string{
				"outbound_otel_backend_plain_format.listener.golden.yaml",
				"outbound_otel_backend_plain_format_1.listener.golden.yaml",
				"outbound_otel_backend_plain_format_2.listener.golden.yaml",
			},
		}),
		Entry("outbound tcpproxy with tcp backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Tcp: &api.TCPBackend{
									Address: "logging.backend",
									Format: &api.Format{
										Plain: pointer.To("custom format [%START_TIME%] %RESPONSE_FLAGS%"),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_tcp_backend_plain_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with tcp backend and json format", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(TcpProxyDeprecated(
							"127.0.0.1:27777",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Tcp: &api.TCPBackend{
									Address: "logging.backend",
									Format: &api.Format{
										Json: pointer.To([]api.JsonValue{
											{Key: "protocol", Value: "%PROTOCOL%"},
											{Key: "duration", Value: "%DURATION%"},
										}),
									},
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_tcp_backend_json_format.listener.golden.yaml"},
		}),
		Entry("basic outbound route without match", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)).
						Configure(
							HttpOutboundRoute(
								"backend",
								envoy_common.Routes{
									{
										Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
											envoy_common.WithService("backend"),
											envoy_common.WithWeight(100),
										)},
									},
								},
								map[string]map[string]bool{
									"kuma.io/service": {
										"web": true,
									},
								},
							),
						),
					)).MustBuild(),
			}},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "other",
						}},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_route_without_match.listener.golden.yaml"},
		}),
		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
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
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
			},
			expectedListeners: []string{"inbound_route.listener.golden.yaml"},
		}),
	)
	type gatewayTestCase struct {
		routes []*core_mesh.MeshGatewayRouteResource
		rules  core_rules.GatewayRules
	}
	DescribeTable("should generate proper Envoy config for MeshGateway Dataplanes",
		func(given gatewayTestCase) {
			gateways := core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{{
					Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									mesh_proto.ServiceTag: "gateway",
								},
							},
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     8080,
								},
							},
						},
					},
				}},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &gateways
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: given.routes,
			}

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
				AddServiceProtocol("other-service", core_mesh.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("gateway").
						WithMesh("default").
						WithBuiltInGateway("gateway"),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshAccessLogType, given.rules)).
				Build()

			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}

			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway.listener.golden.yaml", name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway.cluster.golden.yaml", name))))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway.route.golden.yaml", name))))
		},
		Entry("basic-gateway", gatewayTestCase{
			routes: []*core_mesh.MeshGatewayRouteResource{
				builders.GatewayRoute().
					WithName("sample-gateway-route").
					WithGateway("gateway").
					WithExactMatchHttpRoute("/", "backend", "other-service").
					Build(),
			},
			rules: core_rules.GatewayRules{
				FromRules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 8080}: {
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/from-log",
									},
								}},
							},
						},
					},
				},
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.Rules{
						{Address: "127.0.0.1", Port: 8080}: {
							{
								Subset: core_rules.Subset{},
								Conf: api.Conf{
									Backends: &[]api.Backend{{
										File: &api.FileBackend{
											Path: "/tmp/to-log",
										},
									}},
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
	actualResource, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualResource
}
