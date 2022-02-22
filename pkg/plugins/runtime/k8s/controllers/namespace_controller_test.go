package controllers_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kumahq/kuma/pkg/core"
	v1 "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/apis/k8s.cni.cncf.io/v1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("NamespaceReconciler", func() {

	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sClientScheme).WithObjects(
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
		).Build()

		reconciler = &controllers.NamespaceReconciler{
			Client:     kubeClient,
			CNIEnabled: true,
			Log:        core.Log.WithName("test"),
		}
	})

	It("should create NetworkAttachmentDefinition", func() {
		// setup CustomResourceDefinition
		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: kube_meta.ObjectMeta{
				Name: "network-attachment-definitions.k8s.cni.cncf.io",
			},
		}
		err := kubeClient.Create(context.Background(), crd)
		Expect(err).ToNot(HaveOccurred())

		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "non-system-ns-with-sidecar-injection",
			},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and NetworkAttachmentDefinition is created
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		Expect(err).ToNot(HaveOccurred())
		Expect(nads.Items).To(HaveLen(1))
		Expect(nads.Items[0].Namespace).To(Equal("non-system-ns-with-sidecar-injection"))
		Expect(nads.Items[0].Name).To(Equal("kuma-cni"))
	})

	It("should delete NetworkAttachmentDefinition when injection annotation is no longer on the namespace", func() {
		// setup CustomResourceDefinition
		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: kube_meta.ObjectMeta{
				Name: "network-attachment-definitions.k8s.cni.cncf.io",
			},
		}
		err := kubeClient.Create(context.Background(), crd)
		Expect(err).ToNot(HaveOccurred())

		// setup NetworkAttachmentDefinition in the namespace
		nad := &v1.NetworkAttachmentDefinition{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "non-system-ns-without-sidecar-injection",
				Name:      metadata.KumaCNI,
			},
		}
		err = kubeClient.Create(context.Background(), nad)
		Expect(err).ToNot(HaveOccurred())

		// given namespace without kuma.io/sidecar-injection annotation
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-without-sidecar-injection",
				Name:      "non-system-ns-without-sidecar-injection",
			},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and NetworkAttachmentDefinition is deleted
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		Expect(err).ToNot(HaveOccurred())
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
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and NetworkAttachmentDefinition is not created
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		Expect(err).ToNot(HaveOccurred())
		Expect(nads.Items).To(HaveLen(0))
	})

	It("should skip creating NetworkAttachmentDefinition when CRD is absent in the cluster", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "non-system-ns-with-sidecar-injection",
			},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and NetworkAttachmentDefinition is created
		nads := &v1.NetworkAttachmentDefinitionList{}
		err = kubeClient.List(context.Background(), nads)
		Expect(err).ToNot(HaveOccurred())
		Expect(nads.Items).To(HaveLen(0))
	})

})
