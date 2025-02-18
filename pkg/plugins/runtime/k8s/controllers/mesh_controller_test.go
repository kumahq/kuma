package controllers_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	resources_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	secret_cipher "github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/scheme"
	ca_builtin "github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	v1alpha12 "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	secrets_k8s "github.com/kumahq/kuma/pkg/plugins/secrets/k8s"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("MeshReconciler", func() {
	var kubeClient kube_client.Client
	var resourceManager resources_manager.ResourceManager
	var reconciler kube_reconcile.Reconciler
	var builtinCaManager core_ca.Manager

	BeforeAll(func() {
		kubeClient = kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithIndex(&kube_core.Secret{}, "type",
				func(object kube_client.Object) []string {
					secret := object.(*kube_core.Secret)
					return []string{string(secret.Type)}
				}).
			Build()
		store, err := k8s.NewStore(kubeClient, k8sClientScheme, k8s.NewSimpleConverter())
		Expect(err).ToNot(HaveOccurred())

		// we need to bring in the actual scheme we're using so that the Mesh CRD can be hooked up as owner,
		// otherwise we will get "no kind is registered for the type v1alpha1.Mesh in scheme"
		scheme, err := scheme.NewScheme()
		Expect(err).ToNot(HaveOccurred())
		secretStore, err := secrets_k8s.NewStore(kubeClient, kubeClient, scheme, "default")
		Expect(err).ToNot(HaveOccurred())

		resourceManager = resources_manager.NewResourceManager(store)
		customizableManager := resources_manager.NewCustomizableResourceManager(resourceManager, nil)
		secretManager := secret_manager.NewSecretManager(
			secretStore,
			secret_cipher.None(),
			secret_manager.ValidateDelete(func(ctx context.Context, secretName string, secretMesh string) error { return nil }),
			false,
		)

		customizableManager.Customize(
			system.SecretType,
			secretManager,
		)

		builtinCaManager = ca_builtin.NewBuiltinCaManager(secretManager)
		reconciler = &controllers.MeshReconciler{
			ResourceManager: customizableManager,
			Log:             logr.Discard(),
			Extensions:      context.Background(),
			K8sStore:        true,
			SystemNamespace: "kuma-system",
			CaManagers: core_ca.Managers{
				"builtin": builtinCaManager,
			},
		}
	})

	createMesh := func() {
		mesh := mesh.NewMeshResource()
		mesh.Spec = samples.MeshMTLS().Spec
		Expect(
			resourceManager.Create(context.Background(), mesh, core_store.CreateByKey("default", core_model.NoMesh)),
		).To(Succeed())
	}

	reconcile := func() {
		_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Name: "default",
			},
		})
		Expect(err).ToNot(HaveOccurred())
	}

	Context("on reconcile mesh", func() {
		BeforeAll(func() {
			createMesh()
			reconcile()
		})

		It("should create a default policy", func() {
			meshRetries := &v1alpha12.MeshRetryResourceList{}
			Expect(resourceManager.List(context.Background(), meshRetries, core_store.ListByMesh("default"))).To(Succeed())
			Expect(meshRetries.Items).To(HaveLen(1))
		})

		It("should create default CA", func() {
			_, err := builtinCaManager.GetRootCert(context.Background(), "default", samples.MeshMTLS().Spec.Mtls.Backends[0])
			Expect(err).ToNot(HaveOccurred())
		})
	})
}, Ordered)
