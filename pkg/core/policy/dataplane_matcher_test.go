package policy_test

import (
	"sort"
	"time"

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
				proxy:    &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: nil,
				expected: nil,
			}),
			Entry("policies have no selectors (latest should be selected)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: []policy.DataplanePolicy{
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "b",
							CreationTime: time.Unix(1, 1),
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "a",
							CreationTime: time.Unix(0, 0),
						},
					},
				},
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					},
				},
			}),
			Entry("policies have empty selectors (latest should be selected)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				policies: []policy.DataplanePolicy{
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "b",
							CreationTime: time.Unix(1, 1),
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{},
							},
						},
					},
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "a",
							CreationTime: time.Unix(0, 0),
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
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
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
								{
									Tags: map[string]string{
										"app": "example",
									},
								},
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
			Entry("two policies with the same rank (latest should be picked)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
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
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "b",
							CreationTime: time.Unix(1, 1),
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
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "a",
							CreationTime: time.Unix(0, 0),
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
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
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
			}),
			Entry("gateway dataplane matches policies", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
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
					&mesh_core.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "first",
							CreationTime: time.Unix(1, 1),
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
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "first",
						CreationTime: time.Unix(1, 1),
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
