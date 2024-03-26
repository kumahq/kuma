package v1alpha1_test

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
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

			secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil, false)
			dataSourceLoader := datasource.NewDataSourceLoader(secretManager)
			given.xdsContext.Mesh.DataSourceLoader = dataSourceLoader

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
		Entry("default-meshservice", func() outboundsTestCase {
			outboundTargets := xds_builders.EndpointMap().
				AddEndpoint("backend_svc_8084", xds_builders.Endpoint().
					WithTarget("192.168.0.4").
					WithPort(8084).
					WithWeight(1).
					WithTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP, "region", "us"))
			meshSvc := meshservice_api.MeshServiceResource{
				Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
				Spec: &meshservice_api.MeshService{
					Selector: meshservice_api.Selector{},
					Ports: []meshservice_api.Port{{
						Port:       80,
						TargetPort: 8084,
						Protocol:   core_mesh.ProtocolHTTP,
					}},
					Status: meshservice_api.MeshServiceStatus{
						VIPs: []meshservice_api.VIP{{
							IP: "10.0.0.1",
						}},
					},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
				Items: []*meshservice_api.MeshServiceResource{&meshSvc},
			}
			return outboundsTestCase{
				xdsContext: *xds_builders.Context().
					WithEndpointMap(outboundTargets).
					WithResources(resources).
					AddServiceProtocol("backend_svc_80", core_mesh.ProtocolHTTP).
					Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder().
						AddOutbound(builders.Outbound().
							WithAddress("10.0.0.1").
							WithPort(80).
							WithMeshService("backend", 80),
						),
					).
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
		Entry("match-priority", func() outboundsTestCase {
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
													Type:  api.Exact,
													Value: "/v1/specific",
												},
											}, {
												Method: pointer.To(api.Method("GET")),
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
		Entry("request-header-modifiers", func() outboundsTestCase {
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
		Entry("response-header-modifiers", func() outboundsTestCase {
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
		Entry("request-redirect", func() outboundsTestCase {
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
																BackendRef: common_api.BackendRef{
																	TargetRef: common_api.TargetRef{
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
														},
														{
															Type: api.RequestMirrorType,
															RequestMirror: &api.RequestMirror{
																BackendRef: common_api.BackendRef{
																	TargetRef: common_api.TargetRef{
																		Kind: common_api.MeshService,
																		Name: "backend",
																	},
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
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[rules.InboundListenerHostname]rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"):      commonRules,
										core_rules.NewInboundListenerHostname("192.168.0.1", 8081, "go.dev"): commonRules,
										core_rules.NewInboundListenerHostname("192.168.0.1", 8082, "*.dev"):  commonRules,
									},
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
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[rules.InboundListenerHostname]rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {{
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
										core_rules.NewInboundListenerHostname("192.168.0.1", 8081, "go.dev"): {{
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
								},
							}),
					).
					Build(),
			}
		}()),
		Entry("gateway-listener-different-hostnames-specific", func() outboundsTestCase {
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
								Hostname: "other.dev",
								Tags: map[string]string{
									"hostname": "other",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8080,
								Hostname: "go.dev",
								Tags: map[string]string{
									"hostname": "go",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Port:     8081,
								Tags: map[string]string{
									"hostname": "wild",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTP,
								Hostname: "*",
								Port:     8082,
								Tags: map[string]string{
									"hostname": "wild",
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTPS,
								Hostname: "*.secure.dev",
								Port:     8083,
								Tags: map[string]string{
									"hostname": "secure",
								},
								Tls: &mesh_proto.MeshGateway_TLS_Conf{
									Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
									Certificates: []*system_proto.DataSource{{
										Type: &system_proto.DataSource_Inline{
											Inline: wrapperspb.Bytes([]byte(secret)),
										},
									}},
								},
							},
							{
								Protocol: mesh_proto.MeshGateway_Listener_HTTPS,
								Hostname: "*.super-secure.dev",
								Port:     8083,
								Tags: map[string]string{
									"hostname": "super-secure",
								},
								Tls: &mesh_proto.MeshGateway_TLS_Conf{
									Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
									Certificates: []*system_proto.DataSource{{
										Type: &system_proto.DataSource_Inline{
											Inline: wrapperspb.Bytes([]byte(secret)),
										},
									}},
								},
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
								ToRules: core_rules.GatewayToRules{
									ByListenerAndHostname: map[rules.InboundListenerHostname]rules.Rules{
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "other.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/to-other-dev",
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
										core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "go.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/to-go-dev",
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
										core_rules.NewInboundListenerHostname("192.168.0.1", 8081, ""): {{
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
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8082, ""): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/same-path",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend-wild"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}, {
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"*.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/same-path",
														},
													}},
													Default: api.RuleConf{
														BackendRefs: &[]common_api.BackendRef{{
															TargetRef: builders.TargetRefService("backend-wild-dev"),
															Weight:    pointer.To(uint(100)),
														}},
													},
												}},
											},
										}},
										core_rules.NewInboundListenerHostname("192.168.0.1", 8083, "*.secure.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"first-specific.secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/first-specific-dev",
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
										}, {
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"second-specific.secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/second-specific-dev",
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
										core_rules.NewInboundListenerHostname("192.168.0.1", 8083, "*.super-secure.dev"): {{
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"first-specific.super-secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/first-specific-super-dev",
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
										}, {
											Subset: core_rules.MeshSubset(),
											Conf: api.PolicyDefault{
												Hostnames: []string{"second-specific.super-secure.dev"},
												Rules: []api.Rule{{
													Matches: []api.Match{{
														Path: &api.PathMatch{
															Type:  api.PathPrefix,
															Value: "/second-specific-super-dev",
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
								},
							}),
					).
					Build(),
			}
		}()),
	)
})

// nolint:gosec
const secret = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAqEhj+XS8qgm3raPrP554uXDiPv0np2lCx1wJF4KiwFGJMAV8
qHul/0pcUCX742irsAV39f6sMytXlRpMfBAbJNyZDuqx36s0yMolMxsqMjUHmI+X
W7zrj1xAkPLjB+kireohXkyXESBy3QbcaAW+ftZdFDcNHC7a9W8eSZzCB5R0Sb1S
YMrMq8gXrDbB99fLb2G5wsGXq9xW1g6u7TqWy5TAvYkErMfXfsx3BcvJPPAaI4vb
hPp034KVlucB3h5QSEDF1AIV7A8r1m3I/yHRZqyvhg6Dp4ZTgZw1Sh7QwYsJr6/h
kIVaBjq+gT+I6oPBOnVrc5W3N/fO8yalaAdvswIDAQABAoIBAQCS8ywCMRNy9Ktl
wQdz9aF8Zfvbf1t6UGvVBSSXWCdhA5Jl0dTKl7ccGEZGYvTz33pVamEX+j1LLaT8
eguiJrpdVRl/MikDpVChqgwT9bvCPhaU/YbxwCZ/eNKVANSKGuaCsjpTS1R7yzci
lZQwbhusTOrY9T3Ih44C1va+11mEHY7rAy96r2MgTdpDdWAqhGKxQ88IyNCTvp6u
1I/oWXYDm7QW7HCEWcw2PyFfcfLy4LCPYG7BMX6n1DMSSu6U2PeV1fm6wleawCCN
KxuKQSBHARM9B0pcPpAhGuXO9fHBllz3Tmw0yJYCUopIxPK/r+yMufpsto6KRJOz
had7o4XJAoGBAMSdr1eRG2TBwfQtGS9WxMUrYiCdNCDMFnXsFrKp5kF7ebRyX0lY
41O/KS3SPRmqn6F8t77+VjAvIcCtVWPgTLGo4QyOV09UAcPOrv4qBHRkT8tNyM1n
q15DGd7ICE0LFuK1zjWu1HBz/64hNqJJxC8tcJ1HgQ7sO9Vl0FMHeXcNAoGBANsb
/QqyRixj0UMhST4MoZzxwV+3Y+//mpEL4R1kcFa0K1BrIq80xCzJzK7jrU7XtaeG
0WZpksYqexzN6kXvuJy3w5rC4LC2/+MHspYKvdkUMjctB1XIAPF2FtdrSfMDjweS
ItJ1QqALcc83XzAMkrrCUUeL45SGWxRp3yLljtG/AoGAcPAWwRkEADtf+q9RESUp
QAysgAls4Q36NOBZJWV8cs7HWQR9gXdClV9v+vcRy8V7jlpCfb5AqcrY+4FVVFqK
E17rbrfwpQufO+dkE3D1QBpCz4gtuPc8s5edq5+BTSf6jF1cRu/W7YVkL5S6ejwf
Ke5TCrUBCB5gPDMQmDDp750CgYAHMdwVRdVYD88HTUiCaRfFd4rKAdOeRd5ldOZn
eKzXrALgGSSCbFEkx1uZQpCmTh8A6URnAIB5UVvJjllrAnwlaUNbCZsnMlsksVQD
6UZiom8jsK7U+kRNqXsGh9ddy3ge34WVM5SEfNu32jGd+ku3JjpVBxrp/Z9wBCn3
k2IlMQKBgQCWsVuAoLcEvtKYSBb4KZZY3+pHkLLxe+K7Cpq5fK7RnueaH9+o1g+8
AdY6vX/j9yVHqfF6DI2tyq0qMcuNkjDirlY3yosZEQOXjW8SIGk3YaHwd4JMqVL6
vBGM7k3/smF7hEG97wUeaMe3IDkP7G4SNZOWbLUy1IjLw8BBK+2FVQ==
-----END RSA PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIDNTCCAh2gAwIBAgIRAK2DKOd4qR4eTfFpTHCY0KAwDQYJKoZIhvcNAQELBQAw
GzEZMBcGA1UEAxMQZWNoby5leGFtcGxlLmNvbTAeFw0yMTExMDEwNDMzNDhaFw0z
MTEwMzAwNDMzNDhaMBsxGTAXBgNVBAMTEGVjaG8uZXhhbXBsZS5jb20wggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCoSGP5dLyqCbeto+s/nni5cOI+/Sen
aULHXAkXgqLAUYkwBXyoe6X/SlxQJfvjaKuwBXf1/qwzK1eVGkx8EBsk3JkO6rHf
qzTIyiUzGyoyNQeYj5dbvOuPXECQ8uMH6SKt6iFeTJcRIHLdBtxoBb5+1l0UNw0c
Ltr1bx5JnMIHlHRJvVJgysyryBesNsH318tvYbnCwZer3FbWDq7tOpbLlMC9iQSs
x9d+zHcFy8k88Boji9uE+nTfgpWW5wHeHlBIQMXUAhXsDyvWbcj/IdFmrK+GDoOn
hlOBnDVKHtDBiwmvr+GQhVoGOr6BP4jqg8E6dWtzlbc3987zJqVoB2+zAgMBAAGj
dDByMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMB
Af8EBTADAQH/MB0GA1UdDgQWBBS+iZdWqEBq5IT4b9Dcdx09MTUuCzAbBgNVHREE
FDASghBlY2hvLmV4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQBRUD8uWq0s
IM3sW+MCAtBQq5ppNstlAeH24w3yO+4v64FqjDUwRLq7uMJza9iNdbYDQZW/NRrv
30Om9PSn02WzlANa2Knm/EoCwgPyA4ED1UD77uWnxOUxfEWeqdOYDElJpIRb+7RO
tW9zD7ZJ89ipvEjL2zGuvKCQKkdYaIm7W2aljDz1olsMgQolHpbTEPjN+RMWiyNs
tDaan+pwBI0OoXzuWPpB8o9jfL7I8YeOQXOmNy/qpvELV8ji3vdPH1xu1NSt1EGV
rZigv0SZ20Y+BHgf0y3Tv0X+Rx96lYiUtfU+54vjokEjSsfF+iauxfL75QuVvAf9
7G3tiTJPwFKA
-----END CERTIFICATE-----
`
