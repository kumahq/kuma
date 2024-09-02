package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type policiesTestCase struct {
	dataplane      *core_mesh.DataplaneResource
	resources      xds_context.Resources
	expectedRoutes core_rules.ToRules
}

var _ = DescribeTable("MatchedPolicies", func(given policiesTestCase) {
	routes, err := plugin.NewPlugin().(core_plugins.PolicyPlugin).MatchedPolicies(given.dataplane, given.resources)
	Expect(err).ToNot(HaveOccurred())
	Expect(routes.ToRules).To(Equal(given.expectedRoutes))
}, Entry("basic-kind-specificity", policiesTestCase{
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
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/v1",
									},
								}},
								Default: api.RuleConf{
									Filters: &[]api.Filter{{}},
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
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/v1",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefServiceSubset("backend", "version", "v1"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}, {
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/v2",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefServiceSubset("backend", "version", "v2"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}},
						}},
					},
				}},
			},
		},
	},
	expectedRoutes: core_rules.ToRules{
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
							Filters: &[]api.Filter{{}},
							BackendRefs: &[]common_api.BackendRef{{
								TargetRef: builders.TargetRefServiceSubset("backend", "version", "v1"),
								Weight:    pointer.To(uint(100)),
							}},
						},
					}, {
						Matches: []api.Match{{
							Path: &api.PathMatch{
								Type:  api.PathPrefix,
								Value: "/v2",
							},
						}},
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{{
								TargetRef: builders.TargetRefServiceSubset("backend", "version", "v2"),
								Weight:    pointer.To(uint(100)),
							}},
						},
					}},
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
				BackendRefOriginIndex: map[core_rules.MatchesHash]int{
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}})): 1,
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v2"}}})): 1,
				},
			},
		},
		ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{},
	},
}), Entry("tie-breaking", policiesTestCase{
	dataplane: samples.DataplaneWeb(),
	resources: xds_context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
				Items: []*api.MeshHTTPRouteResource{{
					Meta: &test_model.ResourceMeta{
						Mesh: core_model.DefaultMesh,
						Name: "a-route",
					},
					Spec: &api.MeshHTTPRoute{
						TargetRef: builders.TargetRefMesh(),
						To: []api.To{{
							TargetRef: builders.TargetRefService("backend"),
							Rules: []api.Rule{{
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/v1",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefServiceSubset("a-backend", "version", "v1"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}},
						}},
					},
				}, {
					Meta: &test_model.ResourceMeta{
						Mesh: core_model.DefaultMesh,
						Name: "b-route",
					},
					Spec: &api.MeshHTTPRoute{
						TargetRef: builders.TargetRefMesh(),
						To: []api.To{{
							TargetRef: builders.TargetRefService("backend"),
							Rules: []api.Rule{{
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/v1",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefServiceSubset("b-backend", "version", "v1"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}},
						}},
					},
				}},
			},
		},
	},
	expectedRoutes: core_rules.ToRules{
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
								TargetRef: builders.TargetRefServiceSubset("a-backend", "version", "v1"),
								Weight:    pointer.To(uint(100)),
							}},
						},
					}},
				},
				Origin: []core_model.ResourceMeta{
					&test_model.ResourceMeta{
						Mesh: "default",
						Name: "b-route",
					},
					&test_model.ResourceMeta{
						Mesh: "default",
						Name: "a-route",
					},
				},
				BackendRefOriginIndex: map[core_rules.MatchesHash]int{
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}})): 1,
				},
			},
		},
		ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{},
	},
}), Entry("ordering", policiesTestCase{
	dataplane: samples.DataplaneWeb(),
	resources: xds_context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
				Items: []*api.MeshHTTPRouteResource{{
					Meta: &test_model.ResourceMeta{
						Mesh: core_model.DefaultMesh,
						Name: "a-route",
					},
					Spec: &api.MeshHTTPRoute{
						TargetRef: builders.TargetRefMesh(),
						To: []api.To{{
							TargetRef: builders.TargetRefService("backend"),
							Rules: []api.Rule{{
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/a-first-prefix",
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
										Value: "/a-second-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("first-time-in-list-backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}, {
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/should-be-first-shared-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("a-backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}, {
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/should-be-second-shared-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("a-backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}, {
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/a-second-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("second-time-in-list-backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}},
						}},
					},
				}, {
					Meta: &test_model.ResourceMeta{
						Mesh: core_model.DefaultMesh,
						Name: "b-route",
					},
					Spec: &api.MeshHTTPRoute{
						TargetRef: builders.TargetRefMesh(),
						To: []api.To{{
							TargetRef: builders.TargetRefService("backend"),
							Rules: []api.Rule{{
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/b-first-prefix",
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
										Value: "/should-be-second-shared-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("b-backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}, {
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/should-be-first-shared-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("b-backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}, {
								Matches: []api.Match{{
									Path: &api.PathMatch{
										Type:  api.PathPrefix,
										Value: "/b-second-prefix",
									},
								}},
								Default: api.RuleConf{
									BackendRefs: &[]common_api.BackendRef{{
										TargetRef: builders.TargetRefService("backend"),
										Weight:    pointer.To(uint(100)),
									}},
								},
							}},
						}},
					},
				}},
			},
		},
	},
	expectedRoutes: core_rules.ToRules{
		Rules: core_rules.Rules{
			{
				Subset: core_rules.MeshService("backend"),
				Conf: api.PolicyDefault{
					Rules: []api.Rule{{
						Matches: []api.Match{{
							Path: &api.PathMatch{
								Type:  api.PathPrefix,
								Value: "/a-first-prefix",
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
								Value: "/a-second-prefix",
							},
						}},
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{{
								TargetRef: builders.TargetRefService("first-time-in-list-backend"),
								Weight:    pointer.To(uint(100)),
							}},
						},
					}, {
						Matches: []api.Match{{
							Path: &api.PathMatch{
								Type:  api.PathPrefix,
								Value: "/should-be-first-shared-prefix",
							},
						}},
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{{
								TargetRef: builders.TargetRefService("a-backend"),
								Weight:    pointer.To(uint(100)),
							}},
						},
					}, {
						Matches: []api.Match{{
							Path: &api.PathMatch{
								Type:  api.PathPrefix,
								Value: "/should-be-second-shared-prefix",
							},
						}},
						Default: api.RuleConf{
							BackendRefs: &[]common_api.BackendRef{{
								TargetRef: builders.TargetRefService("a-backend"),
								Weight:    pointer.To(uint(100)),
							}},
						},
					}, {
						Matches: []api.Match{{
							Path: &api.PathMatch{
								Type:  api.PathPrefix,
								Value: "/b-first-prefix",
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
								Value: "/b-second-prefix",
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
				Origin: []core_model.ResourceMeta{
					&test_model.ResourceMeta{
						Mesh: "default",
						Name: "b-route",
					},
					&test_model.ResourceMeta{
						Mesh: "default",
						Name: "a-route",
					},
				},
				BackendRefOriginIndex: map[core_rules.MatchesHash]int{
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/a-first-prefix"}}})):                 1,
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/a-second-prefix"}}})):                1,
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/b-first-prefix"}}})):                 0,
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/b-second-prefix"}}})):                0,
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/should-be-first-shared-prefix"}}})):  1,
					core_rules.MatchesHash(api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/should-be-second-shared-prefix"}}})): 1,
				},
			},
		},
		ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{},
	},
}),
)
