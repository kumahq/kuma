package controllers_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

var _ = Describe("MeshDefaultsReconciler", func() {
	var kubeClient kube_client.Client
	var resourceManager resources_manager.ResourceManager

	BeforeEach(func() {
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

		resourceManager = resources_manager.NewResourceManager(store)
	})

	createMesh := func() {
		Expect(
			resourceManager.Create(context.Background(), mesh.NewMeshResource(), core_store.CreateByKey("default", core_model.NoMesh)),
		).To(Succeed())
	}

	hasTrafficPermissions := func() bool {
		trafficPermissions := &mesh.TrafficPermissionResourceList{}
		Expect(
			resourceManager.List(context.Background(), trafficPermissions, core_store.ListByMesh("default")),
		).To(Succeed())
		return len(trafficPermissions.Items) == 1
	}

	It("should not create a default policy", func() {
		createMesh()
		Expect(hasTrafficPermissions()).To(BeFalse())
	})
})
