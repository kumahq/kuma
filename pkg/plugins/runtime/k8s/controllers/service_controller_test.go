package controllers_test

import (
	"context"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"

	"github.com/kumahq/kuma/pkg/core"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ServiceReconciler", func() {

	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewFakeClientWithScheme(
			k8sClientScheme,
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
			})

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
		result, err := reconciler.Reconcile(req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		svc := &kube_core.Service{}
		err = kubeClient.Get(context.Background(), req.NamespacedName, svc)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(svc.GetAnnotations()).ToNot(HaveKey(metadata.IngressServiceUpstream))
	})

	It("should update service in an annotated namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "non-system-ns-with-sidecar-injection", Name: "service"},
		}

		// when
		result, err := reconciler.Reconcile(req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		svc := &kube_core.Service{}
		err = kubeClient.Get(context.Background(), req.NamespacedName, svc)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(svc.GetAnnotations()).To(HaveKey(metadata.IngressServiceUpstream))
		// and
		Expect(svc.GetAnnotations()[metadata.IngressServiceUpstream]).To(Equal(metadata.AnnotationTrue))
	})

})
