package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
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
								konvoy_mesh.ProxyTemplateAnnotation: "custom-proxy-template",
							},
						},
					},
				},
			}

			expected := &konvoy_mesh.ProxyTemplate{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "custom-proxy-template",
					Namespace: "example",
				},
			}

			// setup
			scheme := runtime.NewScheme()
			konvoy_mesh.AddToScheme(scheme)
			resolver := &simpleProxyTemplateResolver{
				Client:               client_fake.NewFakeClientWithScheme(scheme, expected),
				DefaultProxyTemplate: &konvoy_mesh.ProxyTemplate{},
			}

			// when
			actual := resolver.GetTemplate(proxy)

			// then
			Expect(actual).To(Equal(expected))
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
								konvoy_mesh.ProxyTemplateAnnotation: "non-existing",
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
