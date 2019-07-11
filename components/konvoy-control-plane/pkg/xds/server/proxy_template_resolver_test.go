package server

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	konvoy_mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
)

var _ = Describe("Reconcile", func() {
	Describe("simpleProxyTemplateResolver", func() {
		It("should fallback to the default ProxyTemplate when a Pod has no `mesh.getkonvoy.io/proxy-template` annotation", func() {
			// given
			proxy := &model.Proxy{
				Workload: model.Workload{
					Meta: model.WorkloadMeta{
						Name:      "app",
						Namespace: "example",
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				ResourceStore:        memory.NewStore(),
				DefaultProxyTemplate: &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})

		// todo(jakubdyszkiewicz) restore when Proxy is changed to dataplane in simpleProxyTemplateResolver
		XIt("should use Client to resolve ProxyTemplate according to the annotation on a Pod", func() {
			// given
			proxy := &model.Proxy{
				Workload: model.Workload{
					Meta: model.WorkloadMeta{
						Name:      "app",
						Namespace: "example",
						Labels: map[string]string{
							konvoy_mesh_k8s.ProxyTemplateAnnotation: "custom-proxy-template",
						},
					},
				},
			}

			expected := &mesh_core.ProxyTemplateResource{
				Spec: mesh_proto.ProxyTemplate{
					Sources: []*mesh_proto.ProxyTemplateSource{},
				},
			}

			// setup
			ms := memory.NewStore()
			err := ms.Create(context.Background(), expected, store.CreateByKey("example", "custom-proxy-template", "example"))
			Expect(err).ToNot(HaveOccurred())

			resolver := &simpleProxyTemplateResolver{
				ResourceStore:        ms,
				DefaultProxyTemplate: &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(Equal(&mesh_proto.ProxyTemplate{
				Sources: []*mesh_proto.ProxyTemplateSource{},
			}))
		})

		It("should fallback to the default ProxyTemplate when a Pod refers to a ProxyTemplate that doesn't exist", func() {
			// given
			proxy := &model.Proxy{
				Workload: model.Workload{
					Meta: model.WorkloadMeta{
						Name:      "app",
						Namespace: "example",
						Labels: map[string]string{
							konvoy_mesh_k8s.ProxyTemplateAnnotation: "non-existing",
						},
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				ResourceStore:        memory.NewStore(),
				DefaultProxyTemplate: &mesh_proto.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})
	})
})
