package server

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
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
						Mesh: "demo",
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memory.NewStore()),
				DefaultProxyTemplate:    &mesh_proto.ProxyTemplate{},
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
						Mesh: "demo",
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
					Mesh: "demo",
					Name: "expected",
				},
				Spec: mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{"custom-template"},
					},
				},
			}

			other := &mesh_core.ProxyTemplateResource{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "other",
				},
				Spec: mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{"irrelevant-template"},
					},
				},
			}

			// setup
			memStore := memory.NewStore()
			for _, template := range []*mesh_core.ProxyTemplateResource{expected, other} {
				err := memStore.Create(context.Background(), template, store.CreateByKey(template.Meta.GetName(), template.Meta.GetMesh()))
				Expect(err).ToNot(HaveOccurred())
			}

			resolver := &simpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memStore),
				DefaultProxyTemplate:    &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(Equal(&mesh_proto.ProxyTemplate{
				Conf: &mesh_proto.ProxyTemplate_Conf{
					Imports: []string{"custom-template"},
				},
			}))
		})

		It("should fallback to the default ProxyTemplate if there are no custom templates in a given Mesh", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memory.NewStore()),
				DefaultProxyTemplate:    &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})

	})
})
