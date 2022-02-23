package policy_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
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
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
					},
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
					},
				},
				expected: []policy.DataplanePolicy{
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
					},
					&core_mesh.ProxyTemplateResource{
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
				proxy:    &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: nil,
				expected: nil,
			}),
			Entry("policies have no selectors (latest should be selected)", testCase{
				proxy: &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: []policy.DataplanePolicy{
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "b",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.ProxyTemplate{},
					},
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "a",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.ProxyTemplate{},
					},
				},
				expected: &core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					},
					Spec: &mesh_proto.ProxyTemplate{},
				},
			}),
			Entry("policies have empty selectors (latest should be selected)", testCase{
				proxy: &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: []policy.DataplanePolicy{
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "b",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{},
							},
						},
					},
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "a",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{},
							},
						},
					},
				},
				expected: &core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					},
					Spec: &mesh_proto.ProxyTemplate{
						Selectors: []*mesh_proto.Selector{
							{},
						},
					},
				},
			}),
			Entry("policies have non-empty selectors (the one with the highest number of matching key-value pairs should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
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
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
						Spec: &mesh_proto.ProxyTemplate{
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
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "first",
						},
						Spec: &mesh_proto.ProxyTemplate{
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
				expected: &core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
						Name: "first",
					},
					Spec: &mesh_proto.ProxyTemplate{
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
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "b",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"app": "example",
									},
								},
							},
						},
					},
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "a",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.ProxyTemplate{
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
				expected: &core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "b",
						CreationTime: time.Unix(1, 1),
					},
					Spec: &mesh_proto.ProxyTemplate{
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
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh:         "demo",
							Name:         "first",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.ProxyTemplate{
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
				expected: &core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:         "demo",
						Name:         "first",
						CreationTime: time.Unix(1, 1),
					},
					Spec: &mesh_proto.ProxyTemplate{
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
				proxy: &model.Proxy{Dataplane: core_mesh.NewDataplaneResource()},
				policies: []policy.DataplanePolicy{
					&core_mesh.ProxyTemplateResource{
						Meta: &test_model.ResourceMeta{
							Mesh: "demo",
							Name: "last",
						},
						Spec: &mesh_proto.ProxyTemplate{
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
