package server

import (
	"context"
	test_resources "github.com/Kong/kuma/pkg/test/resources"
	"sort"

	"github.com/Kong/kuma/pkg/core/resources/manager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/store"
	model "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("Reconcile", func() {
	Describe("simpleProxyTemplateResolver", func() {
		It("should fallback to the default ProxyTemplate when there are no other candidates", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "pilot",
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				ResourceManager:      manager.NewResourceManager(memory.NewStore(), test_resources.Global()),
				DefaultProxyTemplate: &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})

		It("should use Client to list ProxyTemplates in the same Mesh as Dataplane", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "pilot",
					},
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Tags: map[string]string{
										"app": "example",
									},
								},
							},
						},
					},
				},
			}

			expected := &mesh_core.ProxyTemplateResource{
				Meta: &test_model.ResourceMeta{
					Mesh:      "pilot",
					Namespace: "default",
					Name:      "expected",
				},
				Spec: mesh_proto.ProxyTemplate{
					Imports: []string{"custom-template"},
				},
			}

			other := &mesh_core.ProxyTemplateResource{
				Meta: &test_model.ResourceMeta{
					Mesh:      "default",
					Namespace: "default",
					Name:      "other",
				},
				Spec: mesh_proto.ProxyTemplate{
					Imports: []string{"irrelevant-template"},
				},
			}

			// setup
			memStore := memory.NewStore()
			for _, template := range []*mesh_core.ProxyTemplateResource{expected, other} {
				err := memStore.Create(context.Background(), template, store.CreateByKey(template.Meta.GetNamespace(), template.Meta.GetName(), template.Meta.GetMesh()))
				Expect(err).ToNot(HaveOccurred())
			}

			resolver := &simpleProxyTemplateResolver{
				ResourceManager:      manager.NewResourceManager(memStore, test_resources.Global()),
				DefaultProxyTemplate: &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(Equal(&mesh_proto.ProxyTemplate{
				Imports: []string{"custom-template"},
			}))
		})

		It("should fallback to the default ProxyTemplate if there are no custom templates in a given Mesh", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "pilot",
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				ResourceManager:      manager.NewResourceManager(memory.NewStore(), test_resources.Global()),
				DefaultProxyTemplate: &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})

	})

	Describe("ProxyTemplatesByNamespacedName", func() {

		type testCase struct {
			input    []*mesh_core.ProxyTemplateResource
			expected []*mesh_core.ProxyTemplateResource
		}

		DescribeTable("should sort ProxyTemplates by Namespace and Name",
			func(given testCase) {
				// when
				sort.Stable(ProxyTemplatesByNamespacedName(given.input))
				// then
				Expect(given.input).To(ConsistOf(given.expected))
			},
			Entry("ProxyTemplates in the same Namespace", testCase{
				input: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "last",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "first",
						},
					},
				},
				expected: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "first",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "last",
						},
					},
				},
			}),
			Entry("ProxyTemplates in different Namespaces", testCase{
				input: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "pilot",
							Name:      "a",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "b",
						},
					},
				},
				expected: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "b",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "pilot",
							Name:      "a",
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
				match, score := ScoreMatch(given.selector, given.target)
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
			proxy     *model.Proxy
			templates []*mesh_core.ProxyTemplateResource
			expected  *mesh_core.ProxyTemplateResource
		}

		DescribeTable("should find the best match (the one with the highest number of matching key-value pairs)",
			func(given testCase) {
				// when
				actual := FindBestMatch(given.proxy, given.templates)
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("there are no templates", testCase{
				proxy:     &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				templates: nil,
				expected:  nil,
			}),
			Entry("templates have no selectors (the first one should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				templates: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "last",
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "first",
						},
					},
				},
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:      "pilot",
						Namespace: "default",
						Name:      "first",
					},
				},
			}),
			Entry("templates have empty selectors (the first one should become the best match)", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				templates: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "last",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.ProxyTemplate_Selector{
								{},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "first",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.ProxyTemplate_Selector{
								{},
							},
						},
					},
				},
				expected: &mesh_core.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{
						Mesh:      "pilot",
						Namespace: "default",
						Name:      "first",
					},
					Spec: mesh_proto.ProxyTemplate{
						Selectors: []*mesh_proto.ProxyTemplate_Selector{
							{},
						},
					},
				},
			}),
			Entry("templates have non-empty selectors (the one with the highest number of matching key-value pairs should become the best match)", testCase{
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
				templates: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "last",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.ProxyTemplate_Selector{
								{
									Match: map[string]string{
										"app":     "example",
										"version": "1.0",
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "first",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.ProxyTemplate_Selector{
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
						Mesh:      "pilot",
						Namespace: "default",
						Name:      "first",
					},
					Spec: mesh_proto.ProxyTemplate{
						Selectors: []*mesh_proto.ProxyTemplate_Selector{
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
			Entry("none of templates have matching selectors", testCase{
				proxy: &model.Proxy{Dataplane: &mesh_core.DataplaneResource{}},
				templates: []*mesh_core.ProxyTemplateResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:      "pilot",
							Namespace: "default",
							Name:      "last",
						},
						Spec: mesh_proto.ProxyTemplate{
							Selectors: []*mesh_proto.ProxyTemplate_Selector{
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
