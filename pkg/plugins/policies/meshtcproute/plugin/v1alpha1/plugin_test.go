package v1alpha1_test

import (
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
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute"
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
			rules := core_rules.Rules{
				{
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
			}

			policies := core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					api.MeshTCPRouteType: {
						ToRules: core_rules.ToRules{Rules: rules},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(
						xds_builders.Routing().
							WithOutboundTargets(outboundTargets).
							WithExternalServiceOutboundTargets(externalServiceOutboundTargets),
					).
					WithPolicies(policies).
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
			rules := core_rules.Rules{
				{
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
			}

			policies := core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					api.MeshTCPRouteType: {
						ToRules: core_rules.ToRules{Rules: rules},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(policies).
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
			tcpRules := core_rules.Rules{
				{
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
			}

			httpRules := core_rules.Rules{
				{
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
			}

			policies := core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					api.MeshTCPRouteType: {
						ToRules: core_rules.ToRules{Rules: tcpRules},
					},
					meshhttproute_api.MeshHTTPRouteType: {
						ToRules: core_rules.ToRules{Rules: httpRules},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(policies).
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
			tcpRules := core_rules.Rules{
				{
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
			}

			httpRules := core_rules.Rules{
				{
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
			}

			policies := core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					api.MeshTCPRouteType: {
						ToRules: core_rules.ToRules{Rules: tcpRules},
					},
					meshhttproute_api.MeshHTTPRouteType: {
						ToRules: core_rules.ToRules{Rules: httpRules},
					},
				},
			}

			return outboundsTestCase{
				xdsContext: *xds_builders.Context().WithEndpointMap(outboundTargets).Build(),
				proxy: xds_builders.Proxy().
					WithDataplane(samples.DataplaneWebBuilder()).
					WithRouting(xds_builders.Routing().WithOutboundTargets(outboundTargets)).
					WithPolicies(policies).
					Build(),
			}
		}()),
	)
})
