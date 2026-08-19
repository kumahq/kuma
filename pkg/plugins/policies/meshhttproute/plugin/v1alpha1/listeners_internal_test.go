package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	meshroute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("prepareRoutes", func() {
	// A mirror backendRef without a namespace takes it from the policy it comes
	// from, so the same ref in two namespaces must reach two different services.
	DescribeTable("should resolve a namespace-less request mirror backendRef in the namespace of the policy",
		func(policyNamespace string) {
			backend := builders.MeshService().
				WithName("backend").
				WithMesh(core_model.DefaultMesh).
				AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
				Build()
			paymentsTeamA := builders.MeshService().
				WithName("payments-team-a-hash").
				WithMesh(core_model.DefaultMesh).
				WithLabels(map[string]string{
					mesh_proto.DisplayName:      "payments",
					mesh_proto.KubeNamespaceTag: "team-a",
				}).
				AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
				Build()
			paymentsTeamB := builders.MeshService().
				WithName("payments-team-b-hash").
				WithMesh(core_model.DefaultMesh).
				WithLabels(map[string]string{
					mesh_proto.DisplayName:      "payments",
					mesh_proto.KubeNamespaceTag: "team-b",
				}).
				AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
				Build()

			meshCtx := xds_builders.Context().
				WithMeshLocalResources([]core_model.Resource{backend, paymentsTeamA, paymentsTeamB}).
				Build().
				Mesh

			matches := []api.Match{{
				Path: &api.PathMatch{Type: api.PathPrefix, Value: "/mirror"},
			}}
			policyMeta := &test_model.ResourceMeta{
				Name: "web-route",
				Mesh: core_model.DefaultMesh,
				Labels: map[string]string{
					mesh_proto.KubeNamespaceTag: policyNamespace,
				},
			}
			toRules := core_rules.ToRules{
				ResourceRules: outbound.ResourceRules{
					kri.From(backend): {
						Resource: backend.GetMeta(),
						Conf: []any{api.PolicyDefault{
							Rules: []api.Rule{{
								Matches: matches,
								Default: api.RuleConf{
									Filters: &[]api.Filter{{
										Type: api.RequestMirrorType,
										RequestMirror: &api.RequestMirror{
											BackendRef: common_api.BackendRef{
												TargetRef: common_api.TargetRef{
													Kind: common_api.MeshService,
													Labels: pointer.To(map[string]string{
														mesh_proto.DisplayName: "payments",
													}),
												},
												Port: pointer.To(uint32(8080)),
											},
										},
									}},
								},
							}},
						}},
						OriginByMatches: map[common_api.MatchesHash]common.Origin{
							api.HashMatches(matches): {Resource: policyMeta},
						},
					},
				},
			}
			svc := meshroute_xds.DestinationService{
				Outbound: &xds_types.Outbound{
					Resource: kri.WithSectionName(kri.From(backend), "8080"),
					Port:     8080,
				},
				Protocol: core_meta.ProtocolHTTP,
			}

			routes := prepareRoutes(toRules, svc, meshCtx)

			var mirrored []api.Route
			for _, route := range routes {
				if len(route.MirrorBackendRefs) > 0 {
					mirrored = append(mirrored, route)
				}
			}
			Expect(mirrored).To(HaveLen(1))
			Expect(mirrored[0].MirrorBackendRefs).To(HaveKey(0))

			expected := map[string]core_model.Resource{
				"team-a": paymentsTeamA,
				"team-b": paymentsTeamB,
			}[policyNamespace]

			ref := mirrored[0].MirrorBackendRefs[0]
			Expect(ref.ReferencesRealResource()).To(BeTrue())
			Expect(ref.Resource()).To(Equal(kri.WithSectionName(kri.From(expected), "8080")))
		},
		Entry("policy in team-a", "team-a"),
		Entry("policy in team-b", "team-b"),
	)

	DescribeTable("should keep resolved backendRefs while tracking unresolved declared weight", func(refNames []string, refWeights []uint, expectedAllUnresolved bool, expectedResolved []string, expectedUnresolvedWeight uint) {
		backend := builders.MeshService().
			WithName("backend").
			WithMesh(core_model.DefaultMesh).
			AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
			Build()
		payments := builders.MeshService().
			WithName("payments-hash").
			WithMesh(core_model.DefaultMesh).
			WithLabels(map[string]string{
				mesh_proto.DisplayName:      "payments",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}).
			AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
			Build()

		meshCtx := xds_builders.Context().
			WithMeshLocalResources([]core_model.Resource{backend, payments}).
			Build().
			Mesh

		matches := []api.Match{{
			Path: &api.PathMatch{Type: api.PathPrefix, Value: "/split"},
		}}
		policyMeta := &test_model.ResourceMeta{
			Name: "web-route",
			Mesh: core_model.DefaultMesh,
			Labels: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}
		var backendRefs []common_api.BackendRef
		for i, name := range refNames {
			backendRefs = append(backendRefs, common_api.BackendRef{
				TargetRef: builders.TargetRefMeshService(name, "kuma-demo", ""),
				Port:      pointer.To(uint32(8080)),
				Weight:    pointer.To(refWeights[i]),
			})
		}
		toRules := core_rules.ToRules{
			ResourceRules: outbound.ResourceRules{
				kri.From(backend): {
					Resource: backend.GetMeta(),
					Conf: []any{api.PolicyDefault{
						Rules: []api.Rule{{
							Matches: matches,
							Default: api.RuleConf{
								BackendRefs: &backendRefs,
							},
						}},
					}},
					OriginByMatches: map[common_api.MatchesHash]common.Origin{
						api.HashMatches(matches): {Resource: policyMeta},
					},
				},
			},
		}
		svc := meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Resource: kri.WithSectionName(kri.From(backend), "8080"),
				Port:     8080,
			},
			Protocol: core_meta.ProtocolHTTP,
		}

		routes := prepareRoutes(toRules, svc, meshCtx)

		var matched *api.Route
		for i := range routes {
			if routes[i].Match.Path != nil && routes[i].Match.Path.Value == "/split" {
				matched = &routes[i]
				break
			}
		}
		Expect(matched).ToNot(BeNil())
		Expect(matched.AllBackendRefsUnresolved).To(Equal(expectedAllUnresolved))
		Expect(matched.UnresolvedBackendRefsWeight).To(Equal(expectedUnresolvedWeight))
		Expect(matched.BackendRefs).To(HaveLen(len(expectedResolved)))
		for i, name := range expectedResolved {
			Expect(matched.BackendRefs[i].ReferencesRealResource()).To(BeTrue())
			Expect(matched.BackendRefs[i].Resource().Name).To(Equal(name))
		}
	},
		Entry("equal split unresolved share", []string{"payments", "missing-backend"}, []uint{1, 1}, false, []string{"payments"}, uint(1)),
		Entry("90/10 unresolved share", []string{"payments", "missing-backend"}, []uint{90, 10}, false, []string{"payments"}, uint(10)),
		Entry("all unresolved missing resources", []string{"missing-backend", "other-missing"}, []uint{30, 70}, true, nil, uint(100)),
		Entry("all resolved keeps zero unresolved share", []string{"payments", "payments"}, []uint{30, 70}, false, []string{"payments", "payments"}, uint(0)),
	)

	DescribeTable("should keep resolved backendRefs while tracking missing-port declared weight", func(refPorts []uint32, refWeights []uint, expectedAllUnresolved bool, expectedResolved []uint32, expectedUnresolvedWeight uint) {
		backend := builders.MeshService().
			WithName("backend").
			WithMesh(core_model.DefaultMesh).
			AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
			Build()
		payments := builders.MeshService().
			WithName("payments-hash").
			WithMesh(core_model.DefaultMesh).
			WithLabels(map[string]string{
				mesh_proto.DisplayName:      "payments",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}).
			AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
			Build()

		meshCtx := xds_builders.Context().
			WithMeshLocalResources([]core_model.Resource{backend, payments}).
			Build().
			Mesh

		matches := []api.Match{{
			Path: &api.PathMatch{Type: api.PathPrefix, Value: "/split"},
		}}
		policyMeta := &test_model.ResourceMeta{
			Name: "web-route",
			Mesh: core_model.DefaultMesh,
			Labels: map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			},
		}
		var backendRefs []common_api.BackendRef
		for i, port := range refPorts {
			backendRefs = append(backendRefs, common_api.BackendRef{
				TargetRef: builders.TargetRefMeshService("payments", "kuma-demo", ""),
				Port:      pointer.To(port),
				Weight:    pointer.To(refWeights[i]),
			})
		}
		toRules := core_rules.ToRules{
			ResourceRules: outbound.ResourceRules{
				kri.From(backend): {
					Resource: backend.GetMeta(),
					Conf: []any{api.PolicyDefault{
						Rules: []api.Rule{{
							Matches: matches,
							Default: api.RuleConf{
								BackendRefs: &backendRefs,
							},
						}},
					}},
					OriginByMatches: map[common_api.MatchesHash]common.Origin{
						api.HashMatches(matches): {Resource: policyMeta},
					},
				},
			},
		}
		svc := meshroute_xds.DestinationService{
			Outbound: &xds_types.Outbound{
				Resource: kri.WithSectionName(kri.From(backend), "8080"),
				Port:     8080,
			},
			Protocol: core_meta.ProtocolHTTP,
		}

		routes := prepareRoutes(toRules, svc, meshCtx)

		var matched *api.Route
		for i := range routes {
			if routes[i].Match.Path != nil && routes[i].Match.Path.Value == "/split" {
				matched = &routes[i]
				break
			}
		}
		Expect(matched).ToNot(BeNil())
		Expect(matched.AllBackendRefsUnresolved).To(Equal(expectedAllUnresolved))
		Expect(matched.UnresolvedBackendRefsWeight).To(Equal(expectedUnresolvedWeight))
		Expect(matched.BackendRefs).To(HaveLen(len(expectedResolved)))
		for i, port := range expectedResolved {
			Expect(matched.BackendRefs[i].ReferencesRealResource()).To(BeTrue())
			Expect(matched.BackendRefs[i].Resource().SectionName).To(Equal(fmt.Sprintf("%d", port)))
		}
	},
		Entry("all unresolved missing ports", []uint32{9999, 9998}, []uint{40, 60}, true, nil, uint(100)),
		Entry("mixed missing-port share", []uint32{8080, 9999}, []uint{90, 10}, false, []uint32{8080}, uint(10)),
	)
})
