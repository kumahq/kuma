package policy_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

// fakeDataplanePolicy is a minimal policy.DataplanePolicy implementation used
// to exercise the generic matching/sorting logic without depending on any
// concrete resource type.
type fakeDataplanePolicy struct {
	test_model.Resource

	selectors []*mesh_proto.Selector
}

func (f *fakeDataplanePolicy) Selectors() []*mesh_proto.Selector {
	return f.selectors
}

func newFakePolicy(meta *test_model.ResourceMeta, selectors ...*mesh_proto.Selector) *fakeDataplanePolicy {
	return &fakeDataplanePolicy{
		Resource:  test_model.Resource{Meta: meta},
		selectors: selectors,
	}
}

var _ = Describe("Dataplane matcher", func() {
	Describe("DataplanePolicyByName", func() {
		type testCase struct {
			input    []policy.DataplanePolicy
			expected []policy.DataplanePolicy
		}

		DescribeTable("should sort DataplanePolicy by Name",
			func(given testCase) {
				// when
				sort.Stable(policy.DataplanePolicyByName(given.input))
				// then
				Expect(given.input).To(ConsistOf(given.expected))
			},
			Entry("DataplanePolicy in the same mesh", testCase{
				input: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "last",
					}),
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					}),
				},
				expected: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					}),
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "last",
					}),
				},
			}),
		)
	})

	Describe("SelectDataplanePolicy()", func() {
		type testCase struct {
			proxy    *model.Proxy
			policies []policy.DataplanePolicy
			expected policy.DataplanePolicy
		}

		DescribeTable("should find the best match (the one with the highest number of matching key-value pairs)",
			func(given testCase) {
				// when
				actual := policy.SelectDataplanePolicy(given.proxy.Dataplane, given.policies)
				// then
				if given.expected == nil {
					Expect(actual).To(BeNil())
				} else {
					Expect(actual).To(Equal(given.expected))
				}
			},
			Entry("there are no policies", testCase{
				proxy:    &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: nil,
				expected: nil,
			}),
			Entry("policies have no selectors (latest should be selected)", testCase{
				proxy: &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					}),
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "a",
						CreationTime: time.Unix(0, 0),
					}),
				},
				expected: newFakePolicy(&test_model.ResourceMeta{
					Mesh:         "demo",
					Name:         "b",
					CreationTime: time.Unix(1, 1),
				}),
			}),
			Entry("policies have empty selectors (latest should be selected)", testCase{
				proxy: &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					}, &mesh_proto.Selector{}),
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "a",
						CreationTime: time.Unix(0, 0),
					}, &mesh_proto.Selector{}),
				},
				expected: newFakePolicy(&test_model.ResourceMeta{
					Mesh:         "demo",
					Name:         "b",
					CreationTime: time.Unix(1, 1),
				}, &mesh_proto.Selector{}),
			}),
			Entry("policies have non-empty selectors (the one with the highest number of matching key-value pairs should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: builders.Dataplane().
					WithAddress("192.168.0.1").
					AddInboundOfTags(
						mesh_proto.ServiceTag, "example",
						"app", "example",
					).
					AddInboundOfTags(
						mesh_proto.ServiceTag, "example",
						"app", "example",
						"version", "1.0",
						"env", "prod",
					).
					Build()},
				policies: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "last",
					}, &mesh_proto.Selector{
						Match: map[string]string{
							"app":     "example",
							"version": "1.0",
						},
					}),
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					},
						&mesh_proto.Selector{
							Match: map[string]string{
								"app": "example",
							},
						},
						&mesh_proto.Selector{
							Match: map[string]string{
								"app":     "example",
								"version": "1.0",
								"env":     "prod",
							},
						},
					),
				},
				expected: newFakePolicy(&test_model.ResourceMeta{
					Mesh: "demo",
					Name: "first",
				},
					&mesh_proto.Selector{
						Match: map[string]string{
							"app": "example",
						},
					},
					&mesh_proto.Selector{
						Match: map[string]string{
							"app":     "example",
							"version": "1.0",
							"env":     "prod",
						},
					},
				),
			}),
			Entry("two policies with the same rank (latest should be picked)", testCase{
				proxy: &model.Proxy{Dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"app":     "example",
										"version": "1.0",
										"env":     "prod",
									},
								},
							},
						},
					},
				}},
				policies: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					}, &mesh_proto.Selector{
						Match: map[string]string{
							"app": "example",
						},
					}),
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "a",
						CreationTime: time.Unix(0, 0),
					}, &mesh_proto.Selector{
						Match: map[string]string{
							"app": "example",
						},
					}),
				},
				expected: newFakePolicy(&test_model.ResourceMeta{
					Mesh:         "demo",
					Name:         "b",
					CreationTime: time.Unix(1, 1),
				}, &mesh_proto.Selector{
					Match: map[string]string{
						"app": "example",
					},
				}),
			}),
			Entry("gateway dataplane matches policies", testCase{
				proxy: &model.Proxy{Dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{
								Tags: map[string]string{
									"app":     "example",
									"version": "1.0",
									"env":     "prod",
								},
							},
						},
					},
				}},
				policies: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "first",
						CreationTime: time.Unix(1, 1),
					}, &mesh_proto.Selector{
						Match: map[string]string{
							"app": "example",
						},
					}),
				},
				expected: newFakePolicy(&test_model.ResourceMeta{
					Mesh:         "demo",
					Name:         "first",
					CreationTime: time.Unix(1, 1),
				}, &mesh_proto.Selector{
					Match: map[string]string{
						"app": "example",
					},
				}),
			}),
			Entry("none of policies have matching selectors", testCase{
				proxy: &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: []policy.DataplanePolicy{
					newFakePolicy(&test_model.ResourceMeta{
						Mesh: "demo",
						Name: "last",
					}, &mesh_proto.Selector{
						Match: map[string]string{
							"app": "example",
						},
					}),
				},
				expected: nil,
			}),
		)
	})
})
