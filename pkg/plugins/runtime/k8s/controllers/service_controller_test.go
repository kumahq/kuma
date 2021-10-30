package controllers_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kumahq/kuma/pkg/core"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("ServiceReconciler", func() {

	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sClientScheme).WithObjects(
			&kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name: "non-system-ns-with-sidecar-injection",
					Annotations: map[string]string{
						metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationEnabled,
					},
				},
			},
			&kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name: "non-system-ns-without-sidecar-injection",
					Annotations: map[string]string{
						metadata.KumaIngressAnnotation: metadata.AnnotationEnabled,
					},
				},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "non-system-ns-with-sidecar-injection",
					Name:      "service",
					Annotations: map[string]string{
						"bogus-annotation": "1",
					},
				},
				Spec: kube_core.ServiceSpec{},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "non-system-ns-without-sidecar-injection",
					Name:      "service",
					Annotations: map[string]string{
						"bogus-annotation": "1",
					},
				},
				Spec: kube_core.ServiceSpec{},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:   "non-system-ns-with-sidecar-injection",
					Name:        "non-annotations-service",
					Annotations: nil,
				},
				Spec: kube_core.ServiceSpec{},
			}).Build()

		reconciler = &ServiceReconciler{
			Client: kubeClient,
			Log:    core.Log.WithName("test"),
		}
	})

	It("should ignore service not in an annotated namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "non-system-ns-without-sidecar-injection", Name: "service"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and service is not annotated
		svc := &kube_core.Service{}
		err = kubeClient.Get(context.Background(), req.NamespacedName, svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.GetAnnotations()).ToNot(HaveKey(metadata.IngressServiceUpstream))
	})

	It("should update service in an annotated namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "non-system-ns-with-sidecar-injection", Name: "service"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and service is annotated
		svc := &kube_core.Service{}
		err = kubeClient.Get(context.Background(), req.NamespacedName, svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.GetAnnotations()).To(HaveKey(metadata.IngressServiceUpstream))
		Expect(svc.GetAnnotations()[metadata.IngressServiceUpstream]).To(Equal(metadata.AnnotationTrue))
	})

	It("should update service that has no annotations in an annotated namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "non-system-ns-with-sidecar-injection", Name: "non-annotations-service"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and service is annotated
		svc := &kube_core.Service{}
		err = kubeClient.Get(context.Background(), req.NamespacedName, svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.GetAnnotations()).To(HaveKey(metadata.IngressServiceUpstream))
		Expect(svc.GetAnnotations()[metadata.IngressServiceUpstream]).To(Equal(metadata.AnnotationTrue))
	})

})
