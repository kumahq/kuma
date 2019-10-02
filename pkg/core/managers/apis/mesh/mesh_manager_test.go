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
		caManager = builtin.NewBuiltinCaManager(secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None()))
		resManager = NewMeshManager(resStore, caManager)
	})

	It("Create() should also create a built-in CA", func() {
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

	It("Delete() should delete all associated resources", func() {
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
})
