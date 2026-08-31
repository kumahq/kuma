package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_common "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

type policiesTestCase struct {
	dataplane      *core_mesh.DataplaneResource
	resources      xds_context.Resources
	expectedRoutes core_rules.ToRules
}

func backendMeshServiceMeta() *test_model.ResourceMeta {
	return &test_model.ResourceMeta{
		Mesh: core_model.DefaultMesh,
		Name: "backend",
		Labels: map[string]string{
			mesh_proto.DisplayName: "backend",
		},
	}
}

func backendMeshServiceList() *meshservice_api.MeshServiceResourceList {
	return &meshservice_api.MeshServiceResourceList{
		Items: []*meshservice_api.MeshServiceResource{{
			Meta:   backendMeshServiceMeta(),
			Spec:   &meshservice_api.MeshService{},
			Status: &meshservice_api.MeshServiceStatus{},
		}},
	}
}

func backendMeshServiceIdentifier() kri.Identifier {
	return kri.Identifier{
		ResourceType: meshservice_api.MeshServiceType,
		Mesh:         core_model.DefaultMesh,
		Name:         "backend",
	}
}

func routeOrigin(name string) rules_common.Origin {
	return rules_common.Origin{
		Resource: &test_model.ResourceMeta{
			Mesh: core_model.DefaultMesh,
			Name: name,
		},
		RuleIndex: 0,
	}
}

func gatewayRouteOrigin(name, creationTimestamp string) rules_common.Origin {
	return rules_common.Origin{
		Resource: &test_model.ResourceMeta{
			Mesh: core_model.DefaultMesh,
			Name: name,
			Labels: map[string]string{
				metadata.GatewayAPIRouteCreationTimestampLabel: creationTimestamp,
			},
		},
		RuleIndex: 0,
	}
}

func expectedBackendRoute(
	conf api.PolicyDefault,
	origins []rules_common.Origin,
	originByMatches map[common_api.MatchesHash]rules_common.Origin,
) core_rules.ToRules {
	return core_rules.ToRules{
		ResourceRules: map[kri.Identifier]outbound.ResourceRule{
			backendMeshServiceIdentifier(): {
				Resource:        backendMeshServiceMeta(),
				Conf:            []any{conf},
				Origin:          origins,
				OriginByMatches: originByMatches,
			},
		},
	}
}

var _ = DescribeTable(
	"MatchedPolicies", func(given policiesTestCase) {
		routes, err := plugin.NewPlugin().(core_plugins.PolicyPlugin).MatchedPolicies(given.dataplane, given.resources)
		Expect(err).ToNot(HaveOccurred())
		Expect(routes.ToRules).To(Equal(given.expectedRoutes))
	}, Entry("basic-kind-specificity", policiesTestCase{
		dataplane: samples.DataplaneWeb(),
		resources: xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				meshservice_api.MeshServiceType: backendMeshServiceList(),
				api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
					Items: []*api.MeshHTTPRouteResource{{
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "route-1",
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
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
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefDataplaneName("web-01"))),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: &api.PathMatch{
											Type:  api.PathPrefix,
											Value: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{{
											TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "backend", "version": "v1"}, ""),
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
											TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "backend", "version": "v2"}, ""),
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
		expectedRoutes: expectedBackendRoute(
			api.PolicyDefault{
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
							TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "backend", "version": "v1"}, ""),
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
							TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "backend", "version": "v2"}, ""),
							Weight:    pointer.To(uint(100)),
						}},
					},
				}},
			},
			[]rules_common.Origin{routeOrigin("route-1"), routeOrigin("route-2")},
			map[common_api.MatchesHash]rules_common.Origin{
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}}): routeOrigin("route-2"),
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v2"}}}): routeOrigin("route-2"),
			},
		),
	}), Entry("tie-breaking", policiesTestCase{
		dataplane: samples.DataplaneWeb(),
		resources: xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				meshservice_api.MeshServiceType: backendMeshServiceList(),
				api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
					Items: []*api.MeshHTTPRouteResource{{
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "a-route",
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: &api.PathMatch{
											Type:  api.PathPrefix,
											Value: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{{
											TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "a-backend", "version": "v1"}, ""),
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
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: &api.PathMatch{
											Type:  api.PathPrefix,
											Value: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{{
											TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "b-backend", "version": "v1"}, ""),
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
		expectedRoutes: expectedBackendRoute(
			api.PolicyDefault{
				Rules: []api.Rule{{
					Matches: []api.Match{{
						Path: &api.PathMatch{
							Type:  api.PathPrefix,
							Value: "/v1",
						},
					}},
					Default: api.RuleConf{
						BackendRefs: &[]common_api.BackendRef{{
							TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "a-backend", "version": "v1"}, ""),
							Weight:    pointer.To(uint(100)),
						}},
					},
				}},
			},
			[]rules_common.Origin{routeOrigin("b-route"), routeOrigin("a-route")},
			map[common_api.MatchesHash]rules_common.Origin{
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}}): routeOrigin("a-route"),
			},
		),
	}), Entry("creation-timestamp-tie-breaking", policiesTestCase{
		dataplane: samples.DataplaneWeb(),
		resources: xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				meshservice_api.MeshServiceType: backendMeshServiceList(),
				api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
					Items: []*api.MeshHTTPRouteResource{{
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "a-route",
							Labels: map[string]string{
								metadata.GatewayAPIRouteCreationTimestampLabel: "200",
							},
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: &api.PathMatch{
											Type:  api.PathPrefix,
											Value: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{{
											TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "a-backend", "version": "v1"}, ""),
											Weight:    pointer.To(uint(100)),
										}},
									},
								}},
							}},
						},
					}, {
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "z-route",
							Labels: map[string]string{
								metadata.GatewayAPIRouteCreationTimestampLabel: "100",
							},
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
								Rules: []api.Rule{{
									Matches: []api.Match{{
										Path: &api.PathMatch{
											Type:  api.PathPrefix,
											Value: "/v1",
										},
									}},
									Default: api.RuleConf{
										BackendRefs: &[]common_api.BackendRef{{
											TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "z-backend", "version": "v1"}, ""),
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
		expectedRoutes: expectedBackendRoute(
			api.PolicyDefault{
				Rules: []api.Rule{{
					Matches: []api.Match{{
						Path: &api.PathMatch{
							Type:  api.PathPrefix,
							Value: "/v1",
						},
					}},
					Default: api.RuleConf{
						BackendRefs: &[]common_api.BackendRef{{
							TargetRef: builders.TargetRefMeshServiceLabels(map[string]string{mesh_proto.DisplayName: "z-backend", "version": "v1"}, ""),
							Weight:    pointer.To(uint(100)),
						}},
					},
				}},
			},
			[]rules_common.Origin{gatewayRouteOrigin("a-route", "200"), gatewayRouteOrigin("z-route", "100")},
			map[common_api.MatchesHash]rules_common.Origin{
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/v1"}}}): gatewayRouteOrigin("z-route", "100"),
			},
		),
	}), Entry("ordering", policiesTestCase{
		dataplane: samples.DataplaneWeb(),
		resources: xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				meshservice_api.MeshServiceType: backendMeshServiceList(),
				api.MeshHTTPRouteType: &api.MeshHTTPRouteResourceList{
					Items: []*api.MeshHTTPRouteResource{{
						Meta: &test_model.ResourceMeta{
							Mesh: core_model.DefaultMesh,
							Name: "a-route",
						},
						Spec: &api.MeshHTTPRoute{
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
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
							TargetRef: pointer.To(builders.ToTopLevelTargetRef(builders.TargetRefMesh())),
							To: &[]api.To{{
								TargetRef: builders.ToOutboundTargetRef(builders.TargetRefService("backend")),
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
		expectedRoutes: expectedBackendRoute(
			api.PolicyDefault{
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
			[]rules_common.Origin{routeOrigin("b-route"), routeOrigin("a-route")},
			map[common_api.MatchesHash]rules_common.Origin{
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/a-first-prefix"}}}):                 routeOrigin("a-route"),
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/a-second-prefix"}}}):                routeOrigin("a-route"),
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/b-first-prefix"}}}):                 routeOrigin("b-route"),
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/b-second-prefix"}}}):                routeOrigin("b-route"),
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/should-be-first-shared-prefix"}}}):  routeOrigin("a-route"),
				api.HashMatches([]api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/should-be-second-shared-prefix"}}}): routeOrigin("a-route"),
			},
		),
	}),
)
