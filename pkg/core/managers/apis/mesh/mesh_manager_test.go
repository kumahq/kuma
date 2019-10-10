package mesh

import (
	"context"
	"github.com/Kong/kuma/pkg/core/ca/builtin"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	secrets_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	secrets_store "github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_resources "github.com/Kong/kuma/pkg/test/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mesh Manager", func() {

	const namespace = "default"

	var resManager manager.ResourceManager
	var resStore store.ResourceStore
	var caManager builtin.BuiltinCaManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		secretManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None())
		caManager = builtin.NewBuiltinCaManager(secretManager)
		manager := manager.NewResourceManager(resStore, test_resources.Global())

		resManager = NewMeshManager(resStore, caManager, manager, secretManager)
	})

	Describe("Create()", func() {
		It("should also create a built-in CA", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Mesh:      meshName,
				Namespace: namespace,
				Name:      meshName,
			}

			// when
			mesh := core_mesh.MeshResource{}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and built-in CA is created
			certs, err := caManager.GetRootCerts(context.Background(), meshName)
			Expect(err).ToNot(HaveOccurred())
			Expect(certs).To(HaveLen(1))
		})
	})

	Describe("Delete()", func() {
		It("should delete all associated resources", func() {
			// given mesh
			meshName := "mesh-1"

			mesh := core_mesh.MeshResource{}
			resKey := model.ResourceKey{
				Mesh:      meshName,
				Namespace: namespace,
				Name:      meshName,
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))
			Expect(err).ToNot(HaveOccurred())

			// and resource associated with it
			dp := core_mesh.DataplaneResource{}
			err = resStore.Create(context.Background(), &dp, store.CreateByKey(namespace, "dp-1", meshName))
			Expect(err).ToNot(HaveOccurred())

			// when mesh is deleted
			err = resManager.Delete(context.Background(), &mesh, store.DeleteBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource is deleted
			err = resStore.Get(context.Background(), &core_mesh.DataplaneResource{}, store.GetByKey(namespace, "dp-1", meshName))
			Expect(store.IsResourceNotFound(err)).To(BeTrue())

			// and built-in mesh CA is deleted
			_, err = caManager.GetRootCerts(context.Background(), meshName)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to load CA key pair for Mesh \"mesh-1\": Resource not found: type=\"Secret\" namespace=\"default\" name=\"builtinca.mesh-1\" mesh=\"mesh-1\""))
		})

		It("should delete all associated resources even if mesh is already removed", func() {
			// given resource that was not deleted with mesh
			dp := core_mesh.DataplaneResource{}
			dpKey := model.ResourceKey{
				Mesh:      "already-deleted",
				Namespace: namespace,
				Name:      "dp-1",
			}
			err := resStore.Create(context.Background(), &dp, store.CreateBy(dpKey))
			Expect(err).ToNot(HaveOccurred())

			// when
			mesh := core_mesh.MeshResource{}
			err = resManager.Delete(context.Background(), &mesh, store.DeleteByKey(namespace, "already-deleted", "already-deleted"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource is deleted
			err = resStore.Get(context.Background(), &dp, store.GetBy(dpKey))
			Expect(store.IsResourceNotFound(err)).To(BeTrue())
		})
	})
})
