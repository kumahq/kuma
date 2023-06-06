package v1alpha1_test

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"path/filepath"
	"strings"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_envoy "github.com/kumahq/kuma/pkg/xds/envoy"
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
		expectedRoutes core_xds.ToRules
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
			expectedRoutes: core_xds.ToRules{
				Rules: rules.Rules{
					{
						Subset: rules.MeshService("backend"),
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
		proxy      core_xds.Proxy
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
			Expect(plugin.Apply(resourceSet, given.xdsContext, &given.proxy)).
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
			outboundTargets := core_xds.EndpointMap{
				"backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   8004,
						Tags: map[string]string{
							"kuma.io/service":  "backend",
							"kuma.io/protocol": "tcp",
							"region":           "eu",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.5",
						Port:   8005,
						Tags: map[string]string{
							"kuma.io/service":  "backend",
							"kuma.io/protocol": "http",
							"region":           "us",
						},
						Weight: 1,
					},
				},
				"other-service": []core_xds.Endpoint{
					{
						Target: "192.168.0.6",
						Port:   8006,
						Tags: map[string]string{
							"kuma.io/service":  "other-backend",
							"kuma.io/protocol": "http",
						},
						Weight: 1,
					},
				},
			}

			externalServiceOutboundTargets := core_xds.EndpointMap{
				"externalservice": []core_xds.Endpoint{
					{
						Target: "192.168.0.7",
						Port:   8007,
						Tags: map[string]string{
							"kuma.io/service":  "externalservice",
							"kuma.io/protocol": "http2",
						},
						ExternalService: &core_xds.ExternalService{},
						Weight:          1,
					},
				},
			}

			rules := rules.Rules{
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
						ToRules: core_xds.ToRules{Rules: rules},
					},
				},
			}

			routing := core_xds.Routing{
				OutboundTargets:                outboundTargets,
				ExternalServiceOutboundTargets: externalServiceOutboundTargets,
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
					Routing:    routing,
					Policies:   policies,
				},
			}
		}()),

		Entry("redirect-traffic", func() outboundsTestCase {
			outboundTargets := core_xds.EndpointMap{
				"backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   8004,
						Tags: map[string]string{
							"kuma.io/service":  "backend",
							"kuma.io/protocol": "http",
						},
						Weight: 1,
					},
				},
				"tcp-backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.5",
						Port:   8005,
						Tags: map[string]string{
							"kuma.io/service":  "tcp-backend",
							"kuma.io/protocol": "tcp",
						},
						Weight: 1,
					},
				},
			}

			rules := rules.Rules{
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
						ToRules: core_xds.ToRules{Rules: rules},
					},
				},
			}

			routing := core_xds.Routing{OutboundTargets: outboundTargets}

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
					Routing:    routing,
					Policies:   policies,
				},
			}
		}()),

		Entry("meshhttproute-clash-http-destination", func() outboundsTestCase {
			outboundTargets := core_xds.EndpointMap{
				"backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   8004,
						Tags: map[string]string{
							"kuma.io/service":  "backend",
							"kuma.io/protocol": "http",
						},
						Weight: 1,
					},
				},
				"tcp-backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.5",
						Port:   8005,
						Tags: map[string]string{
							"kuma.io/service":  "tcp-backend",
							"kuma.io/protocol": "tcp",
						},
						Weight: 1,
					},
				},
				"http-backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.6",
						Port:   8006,
						Tags: map[string]string{
							"kuma.io/service":  "http-backend",
							"kuma.io/protocol": "http",
						},
						Weight: 1,
					},
				},
			}

			tcpRules := rules.Rules{
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

			httpRules := rules.Rules{
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
						ToRules: core_xds.ToRules{Rules: tcpRules},
					},
					meshhttproute_api.MeshHTTPRouteType: {
						ToRules: core_xds.ToRules{Rules: httpRules},
					},
				},
			}

			routing := core_xds.Routing{OutboundTargets: outboundTargets}

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
					Routing:    routing,
					Policies:   policies,
				},
			}
		}()),

		Entry("meshhttproute-clash-tcp-destination", func() outboundsTestCase {
			outboundTargets := core_xds.EndpointMap{
				"backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   8004,
						Tags: map[string]string{
							"kuma.io/service":  "backend",
							"kuma.io/protocol": "tcp",
						},
						Weight: 1,
					},
				},
				"tcp-backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.5",
						Port:   8005,
						Tags: map[string]string{
							"kuma.io/service":  "tcp-backend",
							"kuma.io/protocol": "tcp",
						},
						Weight: 1,
					},
				},
				"http-backend": []core_xds.Endpoint{
					{
						Target: "192.168.0.6",
						Port:   8006,
						Tags: map[string]string{
							"kuma.io/service":  "http-backend",
							"kuma.io/protocol": "http",
						},
						Weight: 1,
					},
				},
			}

			tcpRules := rules.Rules{
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

			httpRules := rules.Rules{
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
						ToRules: core_xds.ToRules{Rules: tcpRules},
					},
					meshhttproute_api.MeshHTTPRouteType: {
						ToRules: core_xds.ToRules{Rules: httpRules},
					},
				},
			}

			routing := core_xds.Routing{OutboundTargets: outboundTargets}

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
					Routing:    routing,
					Policies:   policies,
				},
			}
		}()),
	)
})
