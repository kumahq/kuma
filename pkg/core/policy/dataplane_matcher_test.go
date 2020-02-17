package policy_test

import (
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

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
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
					},
				},
				expected: []policy.DataplanePolicy{
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
					},
				},
			}),
		)
	})

	Describe("ScoreMatch()", func() {

		type testCase struct {
			selector      map[string]string
			target        map[string]string
			expectedMatch bool
			expectedScore int
		}

		DescribeTable("should match and score properly",
			func(given testCase) {
				// when
				match, score := policy.ScoreMatch(given.selector, given.target)
				// then
				Expect(match).To(Equal(given.expectedMatch))
				Expect(score).To(Equal(given.expectedScore))
			},
			Entry("both selector and target are nil", testCase{
				selector:      nil,
				target:        nil,
				expectedMatch: true,
				expectedScore: 0,
			}),
			Entry("both selector and target are empty", testCase{
				selector:      map[string]string{},
				target:        map[string]string{},
				expectedMatch: true,
				expectedScore: 0,
			}),
			Entry("empty selector and non-empty target", testCase{
				selector: nil,
				target: map[string]string{
					"app": "example",
				},
				expectedMatch: true,
				expectedScore: 0,
			}),
			Entry("non-empty selector and empty target", testCase{
				selector: map[string]string{
					"app": "example",
				},
				target:        nil,
				expectedMatch: false,
				expectedScore: 0,
			}),
			Entry("selector with more tags than target", testCase{
				selector: map[string]string{
					"app":     "example",
					"version": "0.1",
				},
				target: map[string]string{
					"app": "example",
				},
				expectedMatch: false,
				expectedScore: 0,
			}),
			Entry("selector with the same tags as target", testCase{
				selector: map[string]string{
					"app": "example",
				},
				target: map[string]string{
					"app": "example",
				},
				expectedMatch: true,
				expectedScore: 1,
			}),
			Entry("selector with the same 2 tags as target", testCase{
				selector: map[string]string{
					"app":     "example",
					"version": "0.1",
				},
				target: map[string]string{
					"app":     "example",
					"version": "0.1",
				},
				expectedMatch: true,
				expectedScore: 2,
			}),
		)
	})

	Describe("FindBestMatch()", func() {

		type testCase struct {
			proxy    *model.Proxy
			policies []policy.DataplanePolicy
			expected policy.DataplanePolicy
		}

		DescribeTable("should find the best match (the one with the highest number of matching key-value pairs)",
			func(given testCase) {
				// when
				actual := policy.FindBestMatch(given.proxy.Dataplane, given.policies)
				// then
				if given.expected == nil {
					Expect(actual).To(BeNil())
				} else {
					Expect(actual).To(Equal(given.expected))
				}
			},
			Entry("there are no policies", testCase{
				proxy:    &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: nil,
				expected: nil,
			}),
			Entry("policies have no selectors (the first one should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: []policy.DataplanePolicy{
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
					},
				},
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					},
				},
			}),
			Entry("policies have empty selectors (the first one should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: []policy.DataplanePolicy{
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{},
							},
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{},
							},
						},
					},
				},
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					},
					Spec: mesh_proto.ProxyTemplate{
						Selectors: []*mesh_proto.Selector{
							{},
						},
					},
				},
			}),
			Entry("policies have non-empty selectors (the one with the highest number of matching key-value pairs should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{},
								{
									Tags: map[string]string{
										"app": "example",
									},
								},
								{},
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
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"app":     "example",
										"version": "1.0",
									},
								},
							},
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"app": "example",
									},
								},
								{
									Match: map[string]string{
										"app":     "example",
										"version": "1.0",
										"env":     "prod",
									},
								},
							},
						},
					},
				},
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					},
					Spec: mesh_proto.ProxyTemplate{
						Selectors: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"app": "example",
								},
							},
							{
								Match: map[string]string{
									"app":     "example",
									"version": "1.0",
									"env":     "prod",
								},
							},
						},
					},
				},
			}),
			Entry("none of policies have matching selectors", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: []policy.DataplanePolicy{
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"app": "example",
									},
								},
							},
						},
					},
				},
				expected: nil,
			}),
		)
	})
})
