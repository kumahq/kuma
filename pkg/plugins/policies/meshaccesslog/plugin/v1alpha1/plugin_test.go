package v1alpha1_test

import (
	"path/filepath"
	"strconv"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	motb_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	meshroute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/plugin/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	meshhttproute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/xds"
	meshtcproute_plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
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

	otherMeshServiceTCP := &kri.Identifier{
		ResourceType: meshservice_api.MeshServiceType,
		Mesh:         "default",
		Zone:         "zone-1",
		Namespace:    "other-ns",
		Name:         "other-meshservice-tcp",
	}

	fooMeshServiceTCP := &kri.Identifier{
		ResourceType: meshservice_api.MeshServiceType,
		Mesh:         "default",
		Zone:         "zone-1",
		Namespace:    "other-ns",
		Name:         "foo-meshservice-tcp",
	}

	barMeshServiceTCP := &kri.Identifier{
		ResourceType: meshservice_api.MeshServiceType,
		Mesh:         "default",
		Zone:         "zone-1",
		Namespace:    "other-ns",
		Name:         "bar-meshservice-tcp",
	}

	otelCollectorMotb := otelBackendMotb("otel-collector", "otel-collector", 4317)
	otherOtelCollectorMotb := otelBackendMotb("other-otel-collector", "other-otel-collector", 5317)
	otelCollectorBackendRef := &common_api.BackendResourceRef{
		Kind:   common_api.BackendResourceMeshOpenTelemetryBackend,
		Labels: map[string]string{mesh_proto.DisplayName: "otel-collector"},
	}
	otherOtelCollectorBackendRef := &common_api.BackendResourceRef{
		Kind:   common_api.BackendResourceMeshOpenTelemetryBackend,
		Labels: map[string]string{mesh_proto.DisplayName: "other-otel-collector"},
	}

	type sidecarTestCase struct {
		resources         []core_xds.Resource
		toRules           core_rules.ToRules
		fromRules         core_rules.FromRules
		expectedListeners []string
		expectedClusters  []string
		dataplaneLabels   map[string]string
		inboundName       string
		extraInbounds     []*builders.InboundBuilder
		motbBackends      []*motb_api.MeshOpenTelemetryBackendResource
	}
	DescribeTable(
		"should generate proper Envoy config",
		func(given sidecarTestCase) {
			// given
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			meshResources := xds_context.NewResources()
			if len(given.motbBackends) > 0 {
				meshResources.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType] = &motb_api.MeshOpenTelemetryBackendResourceList{
					Items: given.motbBackends,
				}
			}

			xdsCtx := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(meshResources).
				WithEndpointMap(
					xds_builders.EndpointMap().
						AddEndpoint("backend", xds_builders.Endpoint().WithTags("kuma.io/service", "backend")).
						AddEndpoint("other-service-http", xds_builders.Endpoint().WithTags("kuma.io/service", "other-service")).
						AddEndpoint("other-service-tcp", xds_builders.Endpoint().WithTags("kuma.io/service", "other-service-tcp")),
				).
				AddServiceProtocol("backend", core_meta.ProtocolHTTP).
				AddServiceProtocol("other-service-http", core_meta.ProtocolHTTP).
				AddServiceProtocol("other-service-tcp", core_meta.ProtocolTCP).
				Build()

			inboundBuilder := builders.Inbound().
				WithService("backend").
				WithAddress("127.0.0.1").
				WithPort(17777).
				WithProtocol("http")
			if given.inboundName != "" {
				inboundBuilder = inboundBuilder.WithName(given.inboundName)
			}
			dpBuilder := builders.Dataplane().
				WithName("backend").
				WithMesh("default").
				AddInbound(inboundBuilder)
			for _, extra := range given.extraInbounds {
				dpBuilder = dpBuilder.AddInbound(extra)
			}
			if given.dataplaneLabels != nil {
				dpBuilder = dpBuilder.WithLabels(given.dataplaneLabels)
			}

			proxy := xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithMetadata(&core_xds.DataplaneMetadata{
					WorkDir: "/tmp",
					// Outbounds are always built from real resources, so every
					// proxy here supports unified resource naming.
					Features: xds_types.Features{xds_types.FeatureUnifiedResourceNaming: true},
				}).
				WithDataplane(dpBuilder).
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
		Entry("basic outbound route from real MeshService", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{{
					Split: []envoy_common.Split{
						xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
					},
				}}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
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
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-2").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-2"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-3").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-3"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
				}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/meshservice/log",
									},
								}},
							},
						},
					},
					routeKRI("route-2"): {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/route-2/log",
									},
								}},
							},
						},
					},
					routeKRI("route-3"): {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
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
		Entry("matches MeshHTTPRoute policy when route name carries a unified-naming rule section", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{
					{
						Name: routeKRI("route-1").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-1"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						// Under unified naming, a real-resource route carries a
						// "rule_N" section name identifying the specific
						// MeshHTTPRoute rule that produced it.
						Name: kri.WithSectionName(routeKRI("route-2"), "rule_0").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-2"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
				}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/meshservice/log",
									},
								}},
							},
						},
					},
					// This MeshAccessLog targets the whole MeshHTTPRoute
					// resource (no rule section), which must still match
					// route-2's rule-scoped name.
					routeKRI("route-2"): {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/route-2/log",
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"basic_outbound_meshhttproute_unified_naming.listener.golden.yaml"},
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
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-2").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-2"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
					{
						Name: routeKRI("route-3").String(),
						Match: meshhttproute_api.Match{
							Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.PathPrefix, Value: "/route-3"},
						},
						Split: []envoy_common.Split{
							xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
						},
					},
				}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/meshservice/log",
									},
								}},
							},
						},
					},
					routeKRI("route-2"): {
						Conf: []any{
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
						xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshExternalServiceHTTP, 47777)).Build(),
					},
				}}),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshExternalServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
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
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/log",
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_file_backend_default_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with file backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
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
			},
			expectedListeners: []string{"outbound_file_backend_plain_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with file backend and json format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
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
			},
			expectedListeners: []string{"outbound_file_backend_json_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with tcp backend and default format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.TCPBackendType,
									Tcp: &api.TCPBackend{
										Address: "logging.backend",
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_tcp_backend_default_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with opentelemetry backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
				outboundRealServiceTCPListener(*fooMeshServiceTCP, 37778),
				outboundRealServiceTCPListener(*barMeshServiceTCP, 37779),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.OtelTelemetryBackendType,
									OpenTelemetry: &api.OtelBackend{
										BackendRef: otelCollectorBackendRef,
									},
								}},
							},
						},
					},
					*fooMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.OtelTelemetryBackendType,
									OpenTelemetry: &api.OtelBackend{
										BackendRef: otelCollectorBackendRef,
										Body: &apiextensionsv1.JSON{
											Raw: []byte("%KUMA_MESH%"),
										},
									},
								}},
							},
						},
					},
					*barMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.OtelTelemetryBackendType,
									OpenTelemetry: &api.OtelBackend{
										BackendRef: otherOtelCollectorBackendRef,
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
			motbBackends: []*motb_api.MeshOpenTelemetryBackendResource{otelCollectorMotb, otherOtelCollectorMotb},
		}),
		Entry("outbound tcpproxy with tcp backend and plain format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.TCPBackendType,
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
			},
			expectedListeners: []string{"outbound_tcp_backend_plain_format.listener.golden.yaml"},
		}),
		Entry("outbound tcpproxy with tcp backend and json format", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777),
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.TCPBackendType,
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
			},
			expectedListeners: []string{"outbound_tcp_backend_json_format.listener.golden.yaml"},
		}),
		Entry("outbound route with no matching MeshAccessLog policy", sidecarTestCase{
			resources: []core_xds.Resource{
				otherServiceHTTPListener(),
			},
			toRules:           core_rules.ToRules{},
			expectedListeners: []string{"outbound_route_without_match.listener.golden.yaml"},
		}),
		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(
						NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								HttpInboundRoutes(
									envoy_names.GetInboundRouteName("backend"),
									"backend",
									envoy_common.Routes{
										{
											Clusters: []envoy_common.Cluster{xds.NewClusterBuilder().WithService("backend").Build()},
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
								Type: api.FileBackendType,
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
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
		Entry("inbound with two services on the same port does not duplicate access log", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(
						NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								HttpInboundRoutes(
									envoy_names.GetInboundRouteName("backend"),
									"backend",
									envoy_common.Routes{
										{
											Clusters: []envoy_common.Cluster{xds.NewClusterBuilder().WithService("backend").Build()},
										},
									},
								),
							),
					)).MustBuild(),
			}},
			extraInbounds: []*builders.InboundBuilder{
				builders.Inbound().
					WithService("backend-canary").
					WithAddress("127.0.0.1").
					WithPort(17777).
					WithProtocol("http"),
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
			},
			expectedListeners: []string{"inbound_route_duplicate_port.listener.golden.yaml"},
		}),
		Entry("inbound route in tag-free mode", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(
						NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								HttpInboundRoutes(
									envoy_names.GetInboundRouteName("backend"),
									"backend",
									envoy_common.Routes{
										{
											Clusters: []envoy_common.Cluster{xds.NewClusterBuilder().WithService("backend").Build()},
										},
									},
								),
							),
					)).MustBuild(),
			}},
			inboundName: "http",
			dataplaneLabels: map[string]string{
				mesh_proto.ZoneTag:          "zone-1",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: 17777}: {{
						Subset: subsetutils.Subset{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {{
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
								File: &api.FileBackend{
									Path: "/tmp/log",
								},
							}},
						},
					}},
				},
			},
			expectedListeners: []string{"inbound_route_tagless.listener.golden.yaml"},
		}),
		Entry("outbound otel backend with workload identity and legacy placeholder key", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{{
					Split: []envoy_common.Split{
						xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
					},
				}}),
			},
			dataplaneLabels: map[string]string{
				mesh_proto.ZoneTag:          "zone-1",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				k8s_metadata.KumaWorkload:   "backend",
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.OtelTelemetryBackendType,
									OpenTelemetry: &api.OtelBackend{
										BackendRef: otelCollectorBackendRef,
										Body: &apiextensionsv1.JSON{
											Raw: []byte("%KUMA_MESH% %KUMA_ZONE% %KUMA_WORKLOAD%"),
										},
										Attributes: &[]api.OtelAttribute{
											{Key: "mesh", Value: "%KUMA_MESH%"},
											{Key: "zone", Value: "%KUMA_ZONE%"},
											{Key: "workload", Value: "%KUMA_WORKLOAD%"},
											{Key: "static.zone", Value: "static-zone-value"},
											{Key: "%KUMA_ZONE%", Value: "legacy-zone-key"},
										},
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_otel_workload_identity.listener.golden.yaml"},
			expectedClusters:  []string{"outbound_otel_workload_identity.cluster.golden.yaml"},
			motbBackends:      []*motb_api.MeshOpenTelemetryBackendResource{otelCollectorMotb},
		}),
		Entry("outbound file backend with workload variables", sidecarTestCase{
			resources: []core_xds.Resource{
				outboundRealServiceHTTPListener(*otherMeshServiceHTTP, 27777, []meshhttproute_xds.OutboundRoute{{
					Split: []envoy_common.Split{
						xds.NewSplitBuilder().WithClusterName(destinationName(*otherMeshServiceHTTP, 27777)).Build(),
					},
				}}),
			},
			dataplaneLabels: map[string]string{
				mesh_proto.ZoneTag:          "zone-1",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				k8s_metadata.KumaWorkload:   "backend",
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceHTTP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{
										Path: "/tmp/log",
										Format: &api.Format{
											Plain: pointer.To("%KUMA_MESH% %KUMA_ZONE% %KUMA_WORKLOAD%"),
										},
									},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"outbound_file_workload_identity.listener.golden.yaml"},
		}),
		Entry("inbound with rules[].matches[].spiffeID and catch-all (parallel logging)", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(
						NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
							Configure(
								HttpInboundRoutes(
									envoy_names.GetInboundRouteName("backend"),
									"backend",
									envoy_common.Routes{
										{
											Clusters: []envoy_common.Cluster{xds.NewClusterBuilder().WithService("backend").Build()},
										},
									},
								),
							),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: 17777}: {
						{
							Match: &common_api.Match{
								SpiffeID: &common_api.SpiffeIDMatch{
									Type:  common_api.ExactMatchType,
									Value: "spiffe://default/ns/clients/sa/specific-client",
								},
							},
							Conf: api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{Path: "/tmp/specific.log"},
								}},
							},
						},
						{
							Conf: api.Conf{
								Backends: &[]api.Backend{{
									Type: api.FileBackendType,
									File: &api.FileBackend{Path: "/tmp/all.log"},
								}},
							},
						},
					},
				},
			},
			expectedListeners: []string{"inbound_matches_spiffeid_and_catchall.listener.golden.yaml"},
		}),
		Entry("zone egress listener with rules[].matches[].sni", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound:zoneegress",
				Origin: metadata.OriginEgress,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "10.20.30.40", 10002, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(
						NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(MatchTransportProtocol("tls")).
							Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
							Configure(TcpProxyDeprecated("aws-aurora", xds.NewClusterBuilder().WithService("aws-aurora").Build())),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "10.20.30.40", Port: 10002}: {{
						Match: &common_api.Match{
							SNI: &common_api.SNIMatch{
								Type:  common_api.SNIExactMatchType,
								Value: "sni.extsvc.default.zone-1.aws-aurora.8443",
							},
						},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
								File: &api.FileBackend{Path: "/tmp/aurora.log"},
							}},
						},
					}},
				},
			},
			expectedListeners: []string{"zoneegress_matches_sni.listener.golden.yaml"},
		}),
		Entry("zone ingress listener with rules[].matches[].spiffeID", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound:zoneingress",
				Origin: metadata.OriginIngress,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "10.20.30.40", 10001, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(
						NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
							Configure(MatchTransportProtocol("tls")).
							Configure(MatchServerNames("inbound-backend{mesh=default}")).
							Configure(TcpProxyDeprecated("backend", xds.NewClusterBuilder().WithService("backend").Build())),
					)).MustBuild(),
			}},
			fromRules: core_rules.FromRules{
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "10.20.30.40", Port: 10001}: {{
						Match: &common_api.Match{
							SpiffeID: &common_api.SpiffeIDMatch{
								Type:  common_api.PrefixMatchType,
								Value: "spiffe://default/ns/clients/",
							},
						},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Type: api.FileBackendType,
								File: &api.FileBackend{Path: "/tmp/ingress.log"},
							}},
						},
					}},
				},
			},
			expectedListeners: []string{"zoneingress_matches_spiffeid.listener.golden.yaml"},
		}),
	)

	It("should route opentelemetry backendRef via kuma-dp when feature is enabled", func() {
		const (
			workDir     = "/tmp"
			backendName = "otel-backend"
		)

		resourceSet := core_xds.NewResourceSet()
		outboundListener := outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777)
		resourceSet.Add(&outboundListener)

		motb := motb_api.NewMeshOpenTelemetryBackendResource()
		motb.SetMeta(&test_model.ResourceMeta{
			Mesh:   "default",
			Name:   backendName,
			Labels: map[string]string{mesh_proto.DisplayName: backendName},
		})
		motb.Spec.Endpoint = &motb_api.Endpoint{
			Address: pointer.To("collector.mesh"),
			Port:    pointer.To(int32(4317)),
		}
		motb.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)

		meshResources := xds_context.NewResources()
		meshResources.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType] = &motb_api.MeshOpenTelemetryBackendResourceList{
			Items: []*motb_api.MeshOpenTelemetryBackendResource{motb},
		}

		xdsCtx := *xds_builders.Context().
			WithMeshBuilder(samples.MeshDefaultBuilder()).
			WithResources(meshResources).
			WithEndpointMap(
				xds_builders.EndpointMap().
					AddEndpoint("backend", xds_builders.Endpoint().WithTags("kuma.io/service", "backend")).
					AddEndpoint("other-service-tcp", xds_builders.Endpoint().WithTags("kuma.io/service", "other-service-tcp")),
			).
			AddServiceProtocol("backend", core_meta.ProtocolHTTP).
			AddServiceProtocol("other-service-tcp", core_meta.ProtocolTCP).
			Build()

		proxy := xds_builders.Proxy().
			WithID(*core_xds.BuildProxyId("default", "backend")).
			WithMetadata(&core_xds.DataplaneMetadata{
				WorkDir: workDir,
				Features: xds_types.Features{
					xds_types.FeatureOtelViaKumaDp: true,
				},
			}).
			WithDataplane(
				builders.Dataplane().
					WithName("backend").
					WithMesh("default").
					AddInbound(
						builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777).
							WithProtocol("http"),
					),
			).
			WithPolicies(xds_builders.MatchedPolicies().WithPolicy(api.MeshAccessLogType, core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.OtelTelemetryBackendType,
									OpenTelemetry: &api.OtelBackend{
										BackendRef: &common_api.BackendResourceRef{
											Kind:   common_api.BackendResourceMeshOpenTelemetryBackend,
											Labels: map[string]string{mesh_proto.DisplayName: backendName},
										},
									},
								}},
							},
						},
					},
				},
			}, core_rules.FromRules{})).
			WithInternalAddresses(core_xds.InternalAddress{AddressPrefix: "172.16.0.0", PrefixLen: 12}, core_xds.InternalAddress{AddressPrefix: "fc00::", PrefixLen: 7}).
			Build()

		proxy.OtelPipeBackends = &core_xds.OtelPipeBackends{}

		meshAccessLogPlugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(meshAccessLogPlugin.Apply(resourceSet, xdsCtx, proxy)).To(Succeed())

		expectedSocket := core_xds.OpenTelemetrySocketName(workDir, backendName)

		clusterResources, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ClusterType)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(clusterResources)).To(ContainSubstring(expectedSocket))
		Expect(string(clusterResources)).ToNot(ContainSubstring("collector.mesh"))

		// Plugin adds to OtelPipeBackends accumulator instead of writing dynconf directly.
		// The generator writes the /otel route after all plugins run.
		Expect(proxy.OtelPipeBackends.Empty()).To(BeFalse())
		backends := proxy.OtelPipeBackends.All()
		Expect(backends).To(HaveLen(1))
		Expect(backends[0].SocketPath).To(Equal(expectedSocket))
		Expect(backends[0].Endpoint).To(Equal("collector.mesh:4317"))
		Expect(backends[0].UseHTTP).To(BeFalse())
		Expect(backends[0].Logs).ToNot(BeNil())
		Expect(backends[0].Logs.Enabled).To(BeTrue())
	})

	It("should skip access log for dangling opentelemetry backendRef", func() {
		resourceSet := core_xds.NewResourceSet()
		outboundListener := outboundRealServiceTCPListener(*otherMeshServiceTCP, 37777)
		resourceSet.Add(&outboundListener)

		// No MOTB resources - the backendRef will be dangling
		meshResources := xds_context.NewResources()
		meshResources.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType] = &motb_api.MeshOpenTelemetryBackendResourceList{
			Items: []*motb_api.MeshOpenTelemetryBackendResource{},
		}

		xdsCtx := *xds_builders.Context().
			WithMeshBuilder(samples.MeshDefaultBuilder()).
			WithResources(meshResources).
			WithEndpointMap(
				xds_builders.EndpointMap().
					AddEndpoint("backend", xds_builders.Endpoint().WithTags("kuma.io/service", "backend")).
					AddEndpoint("other-service-tcp", xds_builders.Endpoint().WithTags("kuma.io/service", "other-service-tcp")),
			).
			AddServiceProtocol("backend", core_meta.ProtocolHTTP).
			AddServiceProtocol("other-service-tcp", core_meta.ProtocolTCP).
			Build()

		proxy := xds_builders.Proxy().
			WithID(*core_xds.BuildProxyId("default", "backend")).
			WithMetadata(&core_xds.DataplaneMetadata{
				WorkDir: "/tmp",
				Features: xds_types.Features{
					xds_types.FeatureOtelViaKumaDp: true,
				},
			}).
			WithDataplane(
				builders.Dataplane().
					WithName("backend").
					WithMesh("default").
					AddInbound(
						builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777).
							WithProtocol("http"),
					),
			).
			WithPolicies(xds_builders.MatchedPolicies().WithPolicy(api.MeshAccessLogType, core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					*otherMeshServiceTCP: {
						Conf: []any{
							api.Conf{
								Backends: &[]api.Backend{{
									Type: api.OtelTelemetryBackendType,
									OpenTelemetry: &api.OtelBackend{
										BackendRef: &common_api.BackendResourceRef{
											Kind:   common_api.BackendResourceMeshOpenTelemetryBackend,
											Labels: map[string]string{"kuma.io/display-name": "non-existent-backend"},
										},
									},
								}},
							},
						},
					},
				},
			}, core_rules.FromRules{})).
			WithInternalAddresses(core_xds.InternalAddress{AddressPrefix: "172.16.0.0", PrefixLen: 12}, core_xds.InternalAddress{AddressPrefix: "fc00::", PrefixLen: 7}).
			Build()

		proxy.OtelPipeBackends = &core_xds.OtelPipeBackends{}

		meshAccessLogPlugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(meshAccessLogPlugin.Apply(resourceSet, xdsCtx, proxy)).To(Succeed())

		// No clusters should be created for the dangling backendRef
		clusterResources, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ClusterType)
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.TrimSpace(string(clusterResources))).To(Equal("{}"))

		// No pipe backends should be accumulated
		Expect(proxy.OtelPipeBackends.Empty()).To(BeTrue())

		// The listener must still be valid - no partial access_log entries
		listenerResources, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ListenerType)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(listenerResources)).ToNot(ContainSubstring("open_telemetry"))
	})
})

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
			Protocol:            core_meta.ProtocolHTTP,
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

func outboundRealServiceTCPListener(serviceResourceKRI kri.Identifier, port int32) core_xds.Resource {
	listener, err := meshtcproute_plugin.GenerateOutboundListener(
		&core_xds.Proxy{
			APIVersion: envoy_common.APIV3,
		},
		meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Address:  "127.0.0.1",
				Port:     uint32(port),
				Resource: destinationKRI(serviceResourceKRI, port),
			},
			Protocol: core_meta.ProtocolTCP,
		},
		[]envoy_common.Split{
			xds.NewSplitBuilder().WithClusterName(destinationName(serviceResourceKRI, port)).Build(),
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
				Resource: destinationKRI(serviceResourceKRI, port),
			},
			Protocol: core_meta.ProtocolHTTP,
		},
		routes,
		mesh_proto.MultiValueTagSet{"kuma.io/service": {"backend": true}},
	)
	Expect(err).ToNot(HaveOccurred())
	return *listener
}

// destinationKRI returns the identifier of a destination as it is carried by an
// Outbound: the service resource identifier scoped to the port it is reached on.
func destinationKRI(id kri.Identifier, port int32) kri.Identifier {
	return kri.WithSectionName(id, strconv.Itoa(int(port)))
}

// destinationName is the unified (KRI) name Envoy resources for this destination
// are given, matching what meshroute generates when the proxy supports unified
// resource naming.
func destinationName(id kri.Identifier, port int32) string {
	return destinationKRI(id, port).String()
}

func routeKRI(name string) kri.Identifier {
	return kri.Identifier{ResourceType: meshhttproute_api.MeshHTTPRouteType, Name: name, Mesh: "default"}
}

func otelBackendMotb(name, address string, port int32) *motb_api.MeshOpenTelemetryBackendResource {
	motb := motb_api.NewMeshOpenTelemetryBackendResource()
	motb.SetMeta(&test_model.ResourceMeta{
		Mesh:   "default",
		Name:   name,
		Labels: map[string]string{mesh_proto.DisplayName: name},
	})
	motb.Spec.Endpoint = &motb_api.Endpoint{
		Address: pointer.To(address),
		Port:    pointer.To(port),
	}
	motb.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)
	return motb
}
