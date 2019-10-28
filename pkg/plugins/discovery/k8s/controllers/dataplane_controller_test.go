package controllers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/plugins/discovery/k8s/controllers"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("DataplaneController", func() {
	Describe("ProxyTemplateToDataplanesMapper", func() {
		It("should use Client to list Dataplanes in all namespaces", func() {
			// given
			template := &mesh_k8s.ProxyTemplate{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "example",
					Name:      "custom-proxy-template",
				},
			}

			dataplane1 := &mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "example",
					Name:      "app1",
				},
			}

			dataplane2 := &mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "app2",
				},
			}

			expected := []reconcile.Request{
				{NamespacedName: types.NamespacedName{Namespace: "example", Name: "app1"}},
				{NamespacedName: types.NamespacedName{Namespace: "demo", Name: "app2"}},
			}

			// setup
			scheme := runtime.NewScheme()
			err := mesh_k8s.AddToScheme(scheme)
			Expect(err).ToNot(HaveOccurred())
			mapper := &ProxyTemplateToDataplanesMapper{
				Client: client_fake.NewFakeClientWithScheme(scheme, dataplane1, dataplane2),
			}

			// when
			actual := mapper.Map(handler.MapObject{Meta: template})

			// then
			Expect(actual).To(ConsistOf(expected))
		})

		It("should handle Client errors gracefully", func() {
			// given
			template := &mesh_k8s.ProxyTemplate{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "example",
					Name:      "custom-proxy-template",
				},
			}

			// setup
			mapper := &ProxyTemplateToDataplanesMapper{
				// List operation will fail since Dataplane type hasn't been registered with the scheme
				Client: client_fake.NewFakeClientWithScheme(runtime.NewScheme()),
			}

			// when
			actual := mapper.Map(handler.MapObject{Meta: template})

			// then
			Expect(actual).To(BeNil())
		})
	})
})
