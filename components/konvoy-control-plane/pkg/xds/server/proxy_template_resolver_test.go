package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	konvoy_mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	k8s_core "k8s.io/api/core/v1"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconcile", func() {
	Describe("simpleProxyTemplateResolver", func() {
		It("should fallback to the default ProxyTemplate when a Pod has no `mesh.getkonvoy.io/proxy-template` annotation", func() {
			// given
			proxy := &model.Proxy{
				Workload: model.Workload{
					Meta: &k8s_core.Pod{
						ObjectMeta: k8s_meta.ObjectMeta{
							Name:      "app",
							Namespace: "example",
						},
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				Client:               client_fake.NewFakeClient(),
				DefaultProxyTemplate: &konvoy_mesh.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})

		It("should use Client to resolve ProxyTemplate according to the annotation on a Pod", func() {
			// given
			proxy := &model.Proxy{
				Workload: model.Workload{
					Meta: &k8s_core.Pod{
						ObjectMeta: k8s_meta.ObjectMeta{
							Name:      "app",
							Namespace: "example",
							Annotations: map[string]string{
								konvoy_mesh_k8s.ProxyTemplateAnnotation: "custom-proxy-template",
							},
						},
					},
				},
			}

			expected := &konvoy_mesh_k8s.ProxyTemplate{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "custom-proxy-template",
					Namespace: "example",
				},
				Spec: map[string]interface{}{
					"sources": []interface{}{},
				},
			}

			// setup
			scheme := runtime.NewScheme()
			konvoy_mesh_k8s.AddToScheme(scheme)
			resolver := &simpleProxyTemplateResolver{
				Client:               client_fake.NewFakeClientWithScheme(scheme, expected),
				DefaultProxyTemplate: &konvoy_mesh.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(Equal(&konvoy_mesh.ProxyTemplate{
				Sources: []*konvoy_mesh.ProxyTemplateSource{},
			}))
		})

		It("should fallback to the default ProxyTemplate when a Pod refers to a ProxyTemplate that doesn't exist", func() {
			// given
			proxy := &model.Proxy{
				Workload: model.Workload{
					Meta: &k8s_core.Pod{
						ObjectMeta: k8s_meta.ObjectMeta{
							Name:      "app",
							Namespace: "example",
							Annotations: map[string]string{
								konvoy_mesh_k8s.ProxyTemplateAnnotation: "non-existing",
							},
						},
					},
				},
			}

			// setup
			resolver := &simpleProxyTemplateResolver{
				Client:               client_fake.NewFakeClient(),
				DefaultProxyTemplate: &konvoy_mesh.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(BeIdenticalTo(resolver.DefaultProxyTemplate))
		})
	})
})
