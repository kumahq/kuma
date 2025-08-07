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
	"github.com/kumahq/kuma/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	meshtcproute_plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshAccessLog", func() {
	otherMeshExternalServiceHTTP := &kri.Identifier{
		ResourceType: "MeshExternalService",
		Mesh:         "default",
		Zone:         "",
		Namespace:    "",
		Name:         "other-meshexternalservice-http",
	}

	otherMeshServiceHTTP := &kri.Identifier{
		ResourceType: meshservice_api.MeshServiceType,
		Mesh:         "default",
		Zone:         "zone-1",
		Namespace:    "other-ns",
		Name:         "other-meshservice-http",
		SectionName:  "",
	}

	type sidecarTestCase struct {
		resources         []core_xds.Resource
		outbounds         xds_types.Outbounds
		toRules           core_rules.ToRules
		fromRules         core_rules.FromRules
		expectedListeners []string
		expectedClusters  []string
		features          xds_types.Features
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			xdsCtx := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(xds_context.NewResources()).
				WithEndpointMap(
					xds_builders.EndpointMap().
						AddEndpoint("backend", xds_builders.Endpoint().WithTags("kuma.io/service", "backend")).
						AddEndpoint("other-service-http", xds_builders.Endpoint().WithTags("kuma.io/service", "other-service")).
						AddEndpoint("other-service-tcp", xds_builders.Endpoint().WithTags("kuma.io/service", "other-service-tcp")),
				).
				AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
				AddServiceProtocol("other-service-http", core_mesh.ProtocolHTTP).
				AddServiceProtocol("other-service-tcp", core_mesh.ProtocolTCP).
				Build()

			proxy := xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithMetadata(&core_xds.DataplaneMetadata{
					WorkDir:  "/tmp",
					Features: given.features,
				}).
				WithDataplane(
					builders.Dataplane().
						WithName("backend").
						WithMesh("default").
						AddInbound(builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777).
							WithTags(map[string]string{
								mesh_proto.ProtocolTag: "http",
							}),
						),
				).
				WithOutbounds(append(given.outbounds, &xds_types.Outbound{
					LegacyOutbound: builders.Outbound().
						WithService("other-service-http").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				}, &xds_types.Outbound{
					LegacyOutbound: builders.Outbound().
						WithService("other-service-tcp").
						WithAddress("127.0.0.1").
						WithPort(37777).Build(),
				},
				)).
				WithPolicies(
					xds_builders.MatchedPolicies().WithPolicy(api.MeshAccessLogType, given.toRules, given.fromRules),
				).
				WithInternalAddresses(core_xds.InternalAddress{AddressPrefix: "172.16.0.0", PrefixLen: 12}, core_xds.InternalAddress{AddressPrefix: "fc00::", PrefixLen: 7}).
				Build()

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// then
			Expect(plugin.Apply(resourceSet, *xdsCtx, proxy)).To(Succeed())
			for i, expectedListener := range given.expectedListeners {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ListenerType)[i].Resource)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", expectedListener)))
			}
			for i, expectedCluster := range given.expectedClusters {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[i].Resource)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", expectedCluster)))
			}
		},
		Entry("basic outbound route", sidecarTestCase{
			resources: []core_xds.Resource{
				otherServiceHTTPListener(),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{{
					Split: []envoy_common.Split{
						xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
					},
				}}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
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
		Entry("basic outbound with MeshHTTPRoute", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{
					{
						Name: routeKRI("route-1").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-1"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-2").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-2"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-3").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-3"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
				}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/meshservice/log",
									},
								}},
							},
						},
					},
					routeKRI("route-2"): {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/route-2/log",
									},
								}},
							},
						},
					},
					routeKRI("route-3"): {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/route-3/log",
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"basic_outbound_meshhttproute.listener.golden.yaml"},
		}),
		Entry("disable MAL for MeshHTTPRoute", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{
					{
						Name: routeKRI("route-1").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-1"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-2").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-2"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-3").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-3"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
				}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/meshservice/log",
									},
								}},
							},
						},
					},
					routeKRI("route-2"): {
						Conf: []interface{}{
							api.Conf{
								Backends: &[]api.Backend{},
							},
						},
					},
				},
			},
			expectedListeners: []string{"disable_mal_for_meshhttproute.listener.golden.yaml"},
		}),
		Entry("basic outbound route from real MeshExternalService", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshExternalServiceHTTP, 47777, []meshhttproute_xds.OutboundRoute{{
					Split: []envoy_common.Split{
						xds.NewSplitBuilder().WithClusterName(serviceName(*otherMeshExternalServiceHTTP, 47777)).Build(),
					},
				}}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshExternalServiceHTTP: {
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
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
		Entry("outbound tcpproxy with opentelemetry backend, plain format, unified naming", sidecarTestCase{
			features: map[string]bool{
				xds_types.FeatureUnifiedResourceNaming: true,
			},
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
				outboundServiceTCPListener("foo-service", 37778),
				outboundServiceTCPListener("bar-service", 37779),
			},
			outbounds: xds_types.Outbounds{
				{LegacyOutbound: builders.Outbound().
					WithService("foo-service").
					WithAddress("127.0.0.1").
					WithPort(37778).Build()},
				{LegacyOutbound: builders.Outbound().
					WithService("bar-service").
					WithAddress("127.0.0.1").
					WithPort(37779).Build()},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service-tcp",
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
						Subset: subsetutils.Subset{{
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
						Subset: subsetutils.Subset{{
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
				"outbound_otel_unified_naming.cluster.golden.yaml",
				"outbound_otel_unified_naming_1.cluster.golden.yaml",
			},
			expectedListeners: []string{
				"outbound_otel_unified_naming.listener.golden.yaml",
				"outbound_otel_unified_naming_1.listener.golden.yaml",
				"outbound_otel_unified_naming_2.listener.golden.yaml",
			},
		}),
		Entry("outbound tcpproxy with opentelemetry backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
				outboundServiceTCPListener("foo-service", 37778),
				outboundServiceTCPListener("bar-service", 37779),
			},
			outbounds: xds_types.Outbounds{
				{LegacyOutbound: builders.Outbound().
					WithService("foo-service").
					WithAddress("127.0.0.1").
					WithPort(37778).Build()},
				{LegacyOutbound: builders.Outbound().
					WithService("bar-service").
					WithAddress("127.0.0.1").
					WithPort(37779).Build()},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service-tcp",
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
						Subset: subsetutils.Subset{{
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
						Subset: subsetutils.Subset{{
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
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
			resources: []core_xds.Resource{
				outboundServiceTCPListener("other-service-tcp", 37777),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{},
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
			resources: []core_xds.Resource{
				otherServiceHTTPListener(),
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{{
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
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil)).
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
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: &api.Rule{Default: api.Conf{
							Backends: &[]api.Backend{{
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						}},
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
							Subset: subsetutils.Subset{},
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
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 8080}: {
						{Conf: &api.Rule{
							Default: api.Conf{
								Backends: &[]api.Backend{{
									File: &api.FileBackend{
										Path: "/tmp/from-log",
									},
								}},
							},
						}},
					},
				},
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.ToRules{
						{Address: "127.0.0.1", Port: 8080}: {
							Rules: core_rules.Rules{{
								Subset: subsetutils.Subset{},
								Conf: api.Conf{
									Backends: &[]api.Backend{{
										File: &api.FileBackend{
											Path: "/tmp/to-log",
										},
									}},
								},
							}},
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

func otherServiceHTTPListener() core_xds.Resource {
	listener, err := meshhttproute_plugin.GenerateOutboundListener(
		&core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
		},
		meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Address: "127.0.0.1",
				Port:    27777,
			},
			Protocol:            core_mesh.ProtocolHTTP,
			KumaServiceTagValue: "other-service-http",
		},
		[]meshhttproute_xds.OutboundRoute{{
			Split: []envoy_common.Split{
				xds.NewSplitBuilder().WithClusterName("other-service-http").Build(),
			},
		}},
		mesh_proto.MultiValueTagSet{"kuma.io/service": {"backend": true}},
	)
	Expect(err).ToNot(HaveOccurred())
	return *listener
}

func outboundServiceTCPListener(service string, port uint32) core_xds.Resource {
	listener, err := meshtcproute_plugin.GenerateOutboundListener(
		&core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
		},
		meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Address: "127.0.0.1",
				Port:    port,
			},
			Protocol:            core_mesh.ProtocolTCP,
			KumaServiceTagValue: service,
		},
		[]envoy_common.Split{
			xds.NewSplitBuilder().WithClusterName(service).Build(),
		},
	)
	Expect(err).ToNot(HaveOccurred())
	return *listener
}

func outboundRealServiceHTTPListener(serviceResourceKRI kri.Identifier, port int32, routes []meshhttproute_xds.OutboundRoute) core_xds.Resource {
	listener, err := meshhttproute_plugin.GenerateOutboundListener(
		&core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
		},
		meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Address:  "127.0.0.1",
				Port:     uint32(port),
				Resource: serviceResourceKRI,
			},
			Protocol:            core_mesh.ProtocolHTTP,
			KumaServiceTagValue: serviceName(serviceResourceKRI, port),
		},
		routes,
		mesh_proto.MultiValueTagSet{"kuma.io/service": {"backend": true}},
	)
	Expect(err).ToNot(HaveOccurred())
	return *listener
}

func serviceName(id kri.Identifier, port int32) string {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	Expect(err).ToNot(HaveOccurred())
	return destinationname.ResolveLegacyFromKRI(id, desc.ShortName, port)
}

func routeKRI(name string) kri.Identifier {
	return kri.Identifier{ResourceType: meshhttproute_api.MeshHTTPRouteType, Name: name, Mesh: "default"}
}
