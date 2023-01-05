package v1alpha1_test

import (
	"path/filepath"
	"strings"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_envoy "github.com/kumahq/kuma/pkg/xds/envoy"
)

func getResource(resourceSet *core_xds.ResourceSet, typ envoy_resource.Type) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshHTTPRoute", func() {
	type policiesTestCase struct {
		dataplane      *core_mesh.DataplaneResource
		resources      xds_context.Resources
		expectedRoutes core_xds.ToRules
	}
	DescribeTable("MatchedPolicies", func(given policiesTestCase) {
		routes, err := plugin.NewPlugin().(core_plugins.PolicyPlugin).MatchedPolicies(given.dataplane, given.resources)
		Expect(err).ToNot(HaveOccurred())
		Expect(routes.ToRules).To(Equal(given.expectedRoutes))
	}, Entry("basic", policiesTestCase{
		dataplane: samples.DataplaneWeb(),
		resources: xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
					Items: []*api.MeshHTTPRouteResource{{
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "route-1",
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: builders.TargetRefMesh(),
							To: []api.To{{
								TargetRef: builders.TargetRefService("backend"),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: api.PathMatch{
											Prefix: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]api.BackendRef{{
											TargetRef: builders.TargetRefService("backend"),
											Weight:    100,
										}},
									},
								}},
							}},
						},
					}, {
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "route-2",
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: builders.TargetRefService("web"),
							To: []api.To{{
								TargetRef: builders.TargetRefService("backend"),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: api.PathMatch{
											Prefix: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]api.BackendRef{{
											TargetRef: builders.TargetRefService("backend"),
											Weight:    100,
										}},
									},
								}},
							}},
						},
					}},
				},
			},
		},
		expectedRoutes: core_xds.ToRules{
			Rules: core_xds.Rules{
				{
					Subset: core_xds.MeshService("backend"),
					Conf: api.PolicyDefault{
						AppendRules: []api.Rule{
							{
								Matches: []api.Match{{
									Path: api.PathMatch{
										Prefix: "/v1",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]api.BackendRef{{
										TargetRef: builders.TargetRefService("backend"),
										Weight:    100,
									}},
								},
							},
							{
								Matches: []api.Match{{
									Path: api.PathMatch{
										Prefix: "/v1",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]api.BackendRef{{
										TargetRef: builders.TargetRefService("backend"),
										Weight:    100,
									}},
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
	type routesTestCase struct {
		rules          []plugin.ToRouteRule
		serviceName    string
		expectedRoutes []plugin.Route
	}
	DescribeTable("FindRoutes", func(given routesTestCase) {
		routes := plugin.FindRoutes(given.rules, given.serviceName)
		Expect(routes).To(Equal(given.expectedRoutes))
	}, Entry("basic", routesTestCase{
		rules: []plugin.ToRouteRule{{
			Subset: core_xds.MeshService("backend"),
			Rules: []api.Rule{{
				Matches: []api.Match{{
					Path: api.PathMatch{
						Prefix: "/v1",
					},
				}},
				Default: api.RuleConf{
					BackendRefs: &[]api.BackendRef{{
						TargetRef: builders.TargetRefService("backend"),
						Weight:    100,
					}},
				},
			}},
		}, {
			Subset: core_xds.MeshService("backend"),
			Rules: []api.Rule{{
				Matches: []api.Match{{
					Path: api.PathMatch{
						Prefix: "/v2",
					},
				}},
				Default: api.RuleConf{
					BackendRefs: &[]api.BackendRef{{
						TargetRef: builders.TargetRefService("backend"),
						Weight:    100,
					}},
				},
			}},
		}},
		serviceName: "backend",
		expectedRoutes: []plugin.Route{{
			Matches: []api.Match{{
				Path: api.PathMatch{
					Prefix: "/v1",
				},
			}},
			BackendRefs: []api.BackendRef{
				{
					TargetRef: builders.TargetRefService("backend"),
					Weight:    100,
				},
			},
		}, {
			Matches: []api.Match{{
				Path: api.PathMatch{
					Prefix: "/v2",
				},
			}},
			BackendRefs: []api.BackendRef{
				{
					TargetRef: builders.TargetRefService("backend"),
					Weight:    100,
				},
			},
		}, {
			Matches: []api.Match{{
				Path: api.PathMatch{
					Prefix: "/",
				},
			}},
			BackendRefs: []api.BackendRef{
				{
					TargetRef: builders.TargetRefService("backend"),
					Weight:    100,
				},
			},
		}},
	}),
	)
	type outboundsTestCase struct {
		proxy      core_xds.Proxy
		xdsContext xds_context.Context
	}
	DescribeTable("Apply", func(given outboundsTestCase) {
		metrics, err := metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())

		claCache, err := cla.NewCache(1*time.Second, metrics)
		Expect(err).ToNot(HaveOccurred())
		given.xdsContext.ControlPlane.CLACache = claCache

		resourceSet := core_xds.NewResourceSet()
		plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(plugin.Apply(resourceSet, given.xdsContext, &given.proxy)).To(Succeed())

		nameSplit := strings.Split(GinkgoT().Name(), " ")
		name := nameSplit[len(nameSplit)-1]

		Expect(getResource(resourceSet, envoy_resource.ListenerType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
		Expect(getResource(resourceSet, envoy_resource.ClusterType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
		Expect(getResource(resourceSet, envoy_resource.EndpointType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".endpoints.golden.yaml")))
	}, Entry("default-route", func() outboundsTestCase {
		outboundTargets := core_xds.EndpointMap{
			"backend": []core_xds.Endpoint{{
				Target: "192.168.0.4",
				Port:   8084,
				Tags:   map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http", "region": "us"},
				Weight: 1,
			}},
		}
		return outboundsTestCase{
			xdsContext: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource:    samples.MeshDefault(),
					EndpointMap: outboundTargets,
				},
			},
			proxy: core_xds.Proxy{
				APIVersion: xds_envoy.APIV3,
				Dataplane:  samples.DataplaneWeb(),
				Routing: core_xds.Routing{
					OutboundTargets: outboundTargets,
				},
			},
		}
	}()), Entry("basic", func() outboundsTestCase {
		outboundTargets := core_xds.EndpointMap{
			"backend": []core_xds.Endpoint{{
				Target: "192.168.0.4",
				Port:   8084,
				Tags:   map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http", "region": "eu"},
				Weight: 1,
			}, {
				Target: "192.168.0.5",
				Port:   8084,
				Tags:   map[string]string{"kuma.io/service": "backend", "kuma.io/protocol": "http", "region": "us"},
				Weight: 1,
			}},
		}
		return outboundsTestCase{
			xdsContext: xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource:    builders.Mesh().WithName("default").Build(),
					EndpointMap: outboundTargets,
				},
			},
			proxy: core_xds.Proxy{
				APIVersion: xds_envoy.APIV3,
				Dataplane:  samples.DataplaneWeb(),
				Routing: core_xds.Routing{
					OutboundTargets: outboundTargets,
				},
				Policies: core_xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						api.MeshHTTPRouteType: {
							ToRules: core_xds.ToRules{
								Rules: core_xds.Rules{{
									Subset: core_xds.MeshService("backend"),
									Conf: []api.Rule{{
										Matches: []api.Match{{
											Path: api.PathMatch{
												Prefix: "/v1",
											},
										}},
										Default: api.RuleConf{
											BackendRefs: &[]api.BackendRef{{
												TargetRef: builders.TargetRefService("backend"),
												Weight:    100,
											}},
										},
									}, {
										Matches: []api.Match{{
											Path: api.PathMatch{
												Prefix: "/v2",
											},
										}},
										Default: api.RuleConf{
											BackendRefs: &[]api.BackendRef{{
												TargetRef: builders.TargetRefServiceSubset("backend", "region", "us"),
												Weight:    100,
											}},
										},
									}},
								}},
							},
						},
					},
				},
			},
		}
	}()),
	)
})
