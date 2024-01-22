package v1alpha1_test

import (
	"context"
	"path/filepath"
	"strings"
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
	"github.com/kumahq/kuma/pkg/metrics"
	_ "github.com/kumahq/kuma/pkg/plugins/policies"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_protocol "github.com/kumahq/kuma/pkg/util/protocol"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

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

var _ = Describe("MeshTCPRoute", func() {
	type policiesTestCase struct {
		dataplane      *core_mesh.DataplaneResource
		resources      xds_context.Resources
		expectedRoutes core_rules.ToRules
	}

	DescribeTable("MatchedPolicies",
		func(given policiesTestCase) {
			routes, err := plugin.NewPlugin().(core_plugins.PolicyPlugin).
				MatchedPolicies(given.dataplane, given.resources)
			Expect(err).ToNot(HaveOccurred())
			Expect(routes.ToRules).To(Equal(given.expectedRoutes))
		},

		Entry("basic", policiesTestCase{
			dataplane: samples.DataplaneWeb(),
			resources: xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					api.MeshTCPRouteType: &api.MeshTCPRouteResourceList{
						Items: []*api.MeshTCPRouteResource{
							{
								Meta: &test_model.ResourceMeta{
									Mesh: core_model.DefaultMesh,
									Name: "route-1",
								},
								Spec: &api.MeshTCPRoute{
									TargetRef: builders.TargetRefMesh(),
									To: []api.To{
										{
											TargetRef: builders.TargetRefService("backend"),
										},
									},
								},
							},
							{
								Meta: &test_model.ResourceMeta{
									Mesh: core_model.DefaultMesh,
									Name: "route-2",
								},
								Spec: &api.MeshTCPRoute{
									TargetRef: builders.TargetRefService("web"),
									To: []api.To{
										{
											TargetRef: builders.TargetRefService("backend"),
											Rules: []api.Rule{
												{
													Default: api.RuleConf{
														BackendRefs: []common_api.BackendRef{
															{
																TargetRef: builders.TargetRefServiceSubset(
																	"backend",
																	"version", "v1",
																),
																Weight: pointer.To(uint(50)),
															},
															{
																TargetRef: builders.TargetRefServiceSubset(
																	"backend",
																	"version", "v2",
																),
																Weight: pointer.To(uint(50)),
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
			},
			expectedRoutes: core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"version", "v1",
										),
										Weight: pointer.To(uint(50)),
									},
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"version", "v2",
										),
										Weight: pointer.To(uint(50)),
									},
								},
							},
						},
						Origin: []core_model.ResourceMeta{
							&test_model.ResourceMeta{
								Mesh: "default",
								Name: "route-1",
							},
							&test_model.ResourceMeta{
								Mesh: "default",
								Name: "route-2",
							},
						},
					},
				},
			},
		}),
	)

	type outboundsTestCase struct {
		proxy      *core_xds.Proxy
		xdsContext xds_context.Context
	}

	DescribeTable("Apply",
		func(given outboundsTestCase) {
			metrics, err := metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())

			claCache, err := cla.NewCache(1*time.Second, metrics)
			Expect(err).ToNot(HaveOccurred())
			given.xdsContext.ControlPlane.CLACache = claCache

			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), given.xdsContext.Mesh, given.proxy)).To(Succeed(), n)
			}
			resourceSet := core_xds.NewResourceSet()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resourceSet, given.xdsContext, given.proxy)).
				To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			listenersGolden := filepath.Join("testdata",
				name+".listeners.golden.yaml",
			)
			clustersGolden := filepath.Join("testdata",
				name+".clusters.golden.yaml",
			)
			endpointsGolden := filepath.Join("testdata",
				name+".endpoints.golden.yaml",
			)

			Expect(getResource(resourceSet, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(listenersGolden))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(clustersGolden))
			Expect(getResource(resourceSet, envoy_resource.EndpointType)).
				To(matchers.MatchGoldenYAML(endpointsGolden))
		},

		Entry("split-traffic", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8005).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("other-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			externalServiceOutboundTargets := xds_builders.EndpointMap().
				AddEndpoint("externalservice", xds_builders.Endpoint().
					WithTarget("192.168.0.7").
					WithPort(8007).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "externalservice", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP2))
			rules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "eu",
										),
										Weight: pointer.To(uint(40)),
									},
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "us",
										),
										Weight: pointer.To(uint(15)),
									},
									{
										TargetRef: builders.TargetRefService(
											"other-backend",
										),
										Weight: pointer.To(uint(15)),
									},
									{
										TargetRef: builders.TargetRefService(
											"externalservice",
										),
										Weight: pointer.To(uint(15)),
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithExternalServicesEndpointMap(externalServiceOutboundTargets).
					AddServiceProtocol("backend", util_protocol.GetCommonProtocol(core_mesh.ProtocolTCP, core_mesh.ProtocolHTTP)).
					AddServiceProtocol("other-backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("externalservice", core_mesh.ProtocolHTTP2).
					AddExternalService("externalservice").
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						samples.DataplaneWebBuilder().
							AddOutboundToService("other-backend").
							AddOutboundToService("externalservice"),
					).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets).
							WithExternalServiceOutboundTargets(externalServiceOutboundTargets),
					).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),

		Entry("basic-no-policies", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8005).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("other-service", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			externalServiceOutboundTargets := xds_builders.EndpointMap().
				AddEndpoint("externalservice", xds_builders.Endpoint().
					WithTarget("192.168.0.7").
					WithPort(8007).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "externalservice", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP2))

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", util_protocol.GetCommonProtocol(core_mesh.ProtocolTCP, core_mesh.ProtocolHTTP)).
					AddServiceProtocol("other-backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("externalservice", core_mesh.ProtocolHTTP2).
					AddExternalService("externalservice").
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets).
							WithExternalServiceOutboundTargets(externalServiceOutboundTargets),
					).
					Build(),
			}
		}()),

		Entry("redirect-traffic", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP)).
				AddEndpoint("tcp-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP))
			rules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefService(
											"tcp-backend",
										),
										Weight: pointer.To(uint(1)),
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("tcp-backend", core_mesh.ProtocolTCP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						samples.DataplaneWebBuilder().
							AddOutboundToService("tcp-backend"),
					).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),

		Entry("meshhttproute-clash-http-destination", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP)).
				AddEndpoint("tcp-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)).
				AddEndpoint("http-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "http-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			tcpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefService(
											"tcp-backend",
										),
										Weight: pointer.To(uint(1)),
									},
								},
							},
						},
					},
				},
			}

			httpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: meshhttproute_api.PolicyDefault{
							Rules: []meshhttproute_api.Rule{
								{
									Matches: []meshhttproute_api.Match{
										{
											Path: &meshhttproute_api.PathMatch{
												Type:  meshhttproute_api.PathPrefix,
												Value: "/",
											},
										},
									},
									Default: meshhttproute_api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{
											{
												TargetRef: builders.TargetRefService("http-backend"),
												Weight:    pointer.To(uint(1)),
											},
										},
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("tcp-backend", core_mesh.ProtocolTCP).
					AddServiceProtocol("http-backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						samples.DataplaneWebBuilder().
							AddOutboundToService("tcp-backend").
							AddOutboundToService("http-backend"),
					).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, tcpRules).
							WithToPolicy(meshhttproute_api.MeshHTTPRouteType, httpRules),
					).
					Build(),
			}
		}()),

		Entry("meshhttproute-clash-tcp-destination", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8004).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)).
				AddEndpoint("tcp-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8005).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "tcp-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)).
				AddEndpoint("http-backend", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8006).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "http-backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP))
			tcpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefService(
											"tcp-backend",
										),
										Weight: pointer.To(uint(1)),
									},
								},
							},
						},
					},
				},
			}

			httpRules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: meshhttproute_api.PolicyDefault{
							Rules: []meshhttproute_api.Rule{
								{
									Matches: []meshhttproute_api.Match{
										{
											Path: &meshhttproute_api.PathMatch{
												Type:  meshhttproute_api.PathPrefix,
												Value: "/",
											},
										},
									},
									Default: meshhttproute_api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{
											{
												TargetRef: builders.TargetRefService("http-backend"),
												Weight:    pointer.To(uint(1)),
											},
										},
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolTCP).
					AddServiceProtocol("tcp-backend", core_mesh.ProtocolTCP).
					AddServiceProtocol("http-backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						samples.DataplaneWebBuilder().
							AddOutboundToService("tcp-backend").
							AddOutboundToService("http-backend"),
					).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshTCPRouteType, tcpRules).
							WithToPolicy(meshhttproute_api.MeshHTTPRouteType, httpRules),
					).
					Build(),
			}
		}()),
		Entry("kafka", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8004).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolKafka, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8005).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolKafka, "region", "us"))
			rules := core_rules.ToRules{
				Rules: core_rules.Rules{
					{
						Subset: core_rules.MeshService("backend"),
						Conf: api.Rule{
							Default: api.RuleConf{
								BackendRefs: []common_api.BackendRef{
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "eu",
										),
										Weight: pointer.To(uint(60)),
									},
									{
										TargetRef: builders.TargetRefServiceSubset(
											"backend",
											"region", "us",
										),
										Weight: pointer.To(uint(40)),
									},
								},
							},
						},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolKafka).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(
						samples.DataplaneWebBuilder(),
					).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets),
					).
					WithPolicies(xds_builders.MatchedPolicies().WithToPolicy(api.MeshTCPRouteType, rules)).
					Build(),
			}
		}()),
		Entry("gateway", func() outboundsTestCase {
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
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8081,
								Hostname: "go.dev",
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_TCP,
								Port:     9080,
							},
						},
					},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{gateway},
			}
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "us"),
				)
			xdsContext := xds_builders.Context().
				WithMesh(samples.MeshDefaultBuilder()).
				WithResources(resources).
				WithEndpointMap(outboundTargets).Build()
			return outboundsTestCase{
				xdsContext: *xdsContext,
				proxy: xds_builders.Proxy().
					WithDataplane(samples.GatewayDataplaneBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithGatewayPolicy(api.MeshTCPRouteType, core_rules.GatewayRules{
								ToRules: map[core_rules.InboundListener]core_rules.Rules{
									{Address: "192.168.0.1", Port: 9080}: {
										{
											Subset: core_rules.MeshSubset(),
											Conf: api.RuleConf{
												BackendRefs: []common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
	)
})
