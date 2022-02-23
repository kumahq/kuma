package template

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	model "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("Reconcile", func() {
	Describe("SimpleProxyTemplateResolver", func() {
		It("should return nil when there no other candidates", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
					},
				},
			}

			// setup
			resolver := &SimpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memory.NewStore()),
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeNil())
		})

		It("should use Client to list ProxyTemplates in the same Mesh as Dataplane", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
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

			expected := &core_mesh.ProxyTemplateResource{
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "expected",
				},
				Spec: &mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{"custom-template"},
					},
				},
			}

			other := &core_mesh.ProxyTemplateResource{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "other",
				},
				Spec: &mesh_proto.ProxyTemplate{
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{"irrelevant-template"},
					},
				},
			}

			// setup
			memStore := memory.NewStore()
			for _, template := range []*core_mesh.ProxyTemplateResource{expected, other} {
				err := memStore.Create(context.Background(), template, store.CreateByKey(template.Meta.GetName(), template.Meta.GetMesh()))
				Expect(err).ToNot(HaveOccurred())
			}

			resolver := &SimpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memStore),
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(MatchProto(&mesh_proto.ProxyTemplate{
				Conf: &mesh_proto.ProxyTemplate_Conf{
					Imports: []string{"custom-template"},
				},
			}))
		})

		It("should return nil if there are no custom templates in a given Mesh", func() {
			// given
			proxy := &model.Proxy{
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "demo",
					},
				},
			}

			// setup
			resolver := &SimpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memory.NewStore()),
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeNil())
		})

	})
})
