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

	"github.com/Kong/kuma/pkg/core"
	v1 "github.com/Kong/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io/v1"
	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/controllers"
)

var _ = Describe("NamespaceReconciler", func() {

	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {

		kubeClient = kube_client_fake.NewFakeClientWithScheme(
			k8sClientScheme,
			&kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name:      "non-system-ns-with-sidecar-injection",
					Namespace: "non-system-ns-with-sidecar-injection",
					Labels: map[string]string{
						"kuma.io/sidecar-injection": "enabled",
					},
				},
			},
			&kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name:      "non-system-ns-without-sidecar-injection",
					Namespace: "non-system-ns-without-sidecar-injection",
				},
			},
			&kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name:      "kuma-system",
					Namespace: "kuma-system",
				},
			},
		)

		reconciler = &controllers.NamespaceReconciler{
			Client:          kubeClient,
			SystemNamespace: "kuma-system",
			CNIEnabled:      true,
			Log:             core.Log.WithName("test"),
		}
	})

	It("should create NetworkAttachmentDefinition", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "non-system-ns-with-sidecar-injection",
			},
		}

		// when
		result, err := reconciler.Reconcile(req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(nads.Items).To(HaveLen(1))
		Expect(nads.Items[0].Namespace).To(Equal("non-system-ns-with-sidecar-injection"))
		Expect(nads.Items[0].Name).To(Equal("kuma-cni"))
	})

	It("should ignore system namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "kube-system",
				Name:      "kuma-system",
			},
		}

		// when
		result, err := reconciler.Reconcile(req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(nads.Items).To(HaveLen(0))
	})

	It("should ignore namespace namespaces without label", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-without-sidecar-injection",
				Name:      "non-system-ns-without-sidecar-injection",
			},
		}

		// when
		result, err := reconciler.Reconcile(req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(nads.Items).To(HaveLen(0))
	})
})
