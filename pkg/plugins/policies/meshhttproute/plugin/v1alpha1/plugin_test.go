package v1alpha1_test

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func getResource(resourceSet *core_xds.ResourceSet, typ envoy_resource.Type) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshHTTPRoute", func() {
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
			gatewayGenerator := plugin_gateway.NewGenerator("test-zone")
			_, err = gatewayGenerator.Generate(context.Background(), nil, given.xdsContext, given.proxy)

			Expect(err).NotTo(HaveOccurred())
			resourceSet := core_xds.NewResourceSet()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resourceSet, given.xdsContext, given.proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]

			Expect(getResource(resourceSet, envoy_resource.ListenerType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.EndpointType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
			Expect(getResource(resourceSet, envoy_resource.RouteType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".routes.golden.yaml")))
		},
		Entry("default-route", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			externalServices := xds_builders.EndpointMap().
				AddEndpoint("external-service", xds_builders.Endpoint().
					WithTarget("192.168.0.5").
					WithPort(8085).
					WithWeight(1).
					WithExternalService(&core_xds.ExternalService{}).
					WithTags(mesh_proto.ServiceTag, "external-service", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithExternalServicesEndpointMap(externalServices).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("external-service", core_mesh.ProtocolHTTP).
					AddExternalService("external-service").
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder().
						AddOutboundToService("external-service")).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					Build(),
			}
		}()),
		Entry("basic", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{{
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}, {
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v2",
												},
											}, {
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v3",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefServiceSubset("backend", "region", "us"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}, {
											Matches: []api.Match{{
												QueryParams: []api.QueryParamsMatch{{
													Type:  api.ExactQueryMatch,
													Name:  "v1",
													Value: "true",
												}},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									},
								}},
							}),
					).
					Build(),
			}
		}()),
		Entry("mixed-tcp-and-http-outbounds", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoints("backend",
					xds_builders.Endpoint().
						WithTarget("192.168.0.4").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "eu"),
					xds_builders.Endpoint().
						WithTarget("192.168.0.5").
						WithPort(8084).
						WithWeight(1).
						WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("other-tcp", xds_builders.Endpoint().
					WithTarget("192.168.0.10").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "other-tcp", mesh_proto.ProtocolTag, core_mesh.ProtocolTCP, "region", "eu"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("other-tcp", core_mesh.ProtocolTCP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{{
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/",
												},
											}},
											Default: api.RuleConf{
												BackendRefs: &[]common_api.BackendRef{{
													TargetRef: builders.TargetRefService("backend"),
													Weight:    pointer.To(uint(100)),
												}},
											},
										}},
									},
								}},
							}),
					).
					Build(),
			}
		}()),
		Entry("header-modifiers", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{{
														Type: api.RequestHeaderModifierType,
														RequestHeaderModifier: &api.HeaderModifier{
															Add: []api.HeaderKeyValue{{
																Name:  "request-add-header",
																Value: "add-value",
															}},
															Set: []api.HeaderKeyValue{{
																Name:  "request-set-header",
																Value: "set-value",
															}, {
																Name:  "request-set-header-multiple",
																Value: "one-value,second-value",
															}},
															Remove: []string{
																"request-header-to-remove",
															},
														},
													}, {
														Type: api.ResponseHeaderModifierType,
														ResponseHeaderModifier: &api.HeaderModifier{
															Add: []api.HeaderKeyValue{{
																Name:  "response-add-header",
																Value: "add-value",
															}},
															Set: []api.HeaderKeyValue{{
																Name:  "response-set-header",
																Value: "set-value",
															}},
															Remove: []string{
																"response-header-to-remove",
															},
														},
													}, {
														Type: api.RequestRedirectType,
														RequestRedirect: &api.RequestRedirect{
															Scheme: pointer.To("other"),
														},
													}},
												},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("url-rewrite", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
							Rules: core_rules.Rules{
								{
									Subset: core_rules.MeshService("backend"),
									Conf: api.PolicyDefault{
										Rules: []api.Rule{{
											Matches: []api.Match{{
												Path: &api.PathMatch{
													Type:  api.PathPrefix,
													Value: "/v1",
												},
											}},
											Default: api.RuleConf{
												Filters: &[]api.Filter{{
													Type: api.URLRewriteType,
													URLRewrite: &api.URLRewrite{
														Path: &api.PathRewrite{
															Type:               api.ReplacePrefixMatchType,
															ReplacePrefixMatch: pointer.To("/v2"),
														},
													},
												}},
											},
										}},
									},
								},
							},
						}),
					).
					Build(),
			}
		}()),
		Entry("headers-match", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Headers: []common_api.HeaderMatch{{
														Type:  pointer.To(common_api.HeaderMatchExact),
														Name:  "foo-exact",
														Value: "bar",
													}, {
														Type: pointer.To(common_api.HeaderMatchPresent),
														Name: "foo-present",
													}, {
														Type:  pointer.To(common_api.HeaderMatchRegularExpression),
														Name:  "foo-regex",
														Value: "x.*y",
													}, {
														Type: pointer.To(common_api.HeaderMatchAbsent),
														Name: "foo-absent",
													}, {
														Type:  pointer.To(common_api.HeaderMatchPrefix),
														Name:  "foo-prefix",
														Value: "x",
													}},
												}},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("grpc-service", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolGRPC, "region", "us"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolGRPC).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: builders.TargetRefService("backend"),
														Weight:    pointer.To(uint(100)),
													}},
												},
											}},
										},
									},
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("request-mirror", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us")).
				AddEndpoint("payments", xds_builders.Endpoint().
					WithTarget("192.168.0.6").
					WithPort(8086).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "payments", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us", "version", "v1", "env", "dev"))
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).
					AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
					AddServiceProtocol("payments", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithToPolicy(api.MeshHTTPRouteType, core_rules.ToRules{
								Rules: core_rules.Rules{
									{
										Subset: core_rules.MeshService("backend"),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: api.RuleConf{
													Filters: &[]api.Filter{
														{
															Type: api.RequestMirrorType,
															RequestMirror: &api.RequestMirror{
																Percentage: pointer.To(intstr.FromString("99.9")),
																BackendRef: common_api.TargetRef{
																	Kind: common_api.MeshServiceSubset,
																	Name: "payments",
																	Tags: map[string]string{
																		"version": "v1",
																		"region":  "us",
																		"env":     "dev",
																	},
																},
															},
														},
														{
															Type: api.RequestMirrorType,
															RequestMirror: &api.RequestMirror{
																BackendRef: common_api.TargetRef{
																	Kind: common_api.MeshService,
																	Name: "backend",
																},
															},
														},
													},
												},
											}},
										},
									},
								},
							}),
					).
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
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8082,
								Hostname: "*.dev",
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"),
				)
			xdsContext := xds_builders.Context().
				WithMesh(samples.MeshDefaultBuilder()).
				WithResources(resources).
				WithEndpointMap(outboundTargets).Build()

			commonRules := core_rules.Rules{
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/wild",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Hostnames: []string{"go.dev"},
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/go-dev",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Hostnames: []string{"*.dev"},
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/wild-dev",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
				{
					Subset: core_rules.MeshSubset(),
					Conf: api.PolicyDefault{
						Hostnames: []string{"other.dev"},
						Rules: []api.Rule{{
							Matches: []api.Match{{
								Path: &api.PathMatch{
									Type:  api.PathPrefix,
									Value: "/other-dev",
								},
							}},
							Default: api.RuleConf{
								BackendRefs: &[]common_api.BackendRef{{
									TargetRef: builders.TargetRefService("backend"),
									Weight:    pointer.To(uint(100)),
								}},
							},
						}},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xdsContext,
				proxy: xds_builders.Proxy().
					WithDataplane(samples.GatewayDataplaneBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(
						xds_builders.MatchedPolicies().
							WithGatewayPolicy(api.MeshHTTPRouteType, core_rules.GatewayRules{
								ToRules: map[rules.InboundListener]rules.Rules{
									{Address: "192.168.0.1", Port: 8080}: commonRules,
									{Address: "192.168.0.1", Port: 8081}: commonRules,
									{Address: "192.168.0.1", Port: 8082}: commonRules,
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("gateway-listener-specific", func() outboundsTestCase {
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
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"),
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
							WithGatewayPolicy(api.MeshHTTPRouteType, core_rules.GatewayRules{
								ToRules: map[rules.InboundListener]rules.Rules{
									{Address: "192.168.0.1", Port: 8080}: {{
										Subset: core_rules.MeshSubset(),
										Conf: api.PolicyDefault{
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/wild",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: builders.TargetRefService("backend"),
														Weight:    pointer.To(uint(100)),
													}},
												},
											}},
										},
									}},
									{Address: "192.168.0.1", Port: 8081}: {{
										Subset: core_rules.MeshSubset(),
										Conf: api.PolicyDefault{
											Hostnames: []string{"go.dev"},
											Rules: []api.Rule{{
												Matches: []api.Match{{
													Path: &api.PathMatch{
														Type:  api.PathPrefix,
														Value: "/go-dev",
													},
												}},
												Default: api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: builders.TargetRefService("backend"),
														Weight:    pointer.To(uint(100)),
													}},
												},
											}},
										},
									}},
								},
							}),
					).
					Build(),
			}
		}()),
	)
})
