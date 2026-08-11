package mesh_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_store "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/config/multizone"
	"github.com/kumahq/kuma/v3/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secrets_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secrets_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_resources "github.com/kumahq/kuma/v3/pkg/test/resources"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
)

var _ = Describe("Mesh Manager", func() {
	var resManager manager.ResourceManager
	var unsafeDeleteResManager manager.ResourceManager
	var secretManager manager.ResourceManager
	var resStore store.ResourceStore

	BeforeEach(func() {
		resStore = memory.NewStore()
		secretManager = secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None())

		manager := manager.NewResourceManager(resStore)
		validator := mesh.NewMeshValidator(resStore)
		resManager = mesh.NewMeshManager(
			resStore, manager, test_resources.Global(),
			validator, context.Background(),
			kuma_cp.Config{
				Store: &config_store.StoreConfig{
					Type: config_store.MemoryStore, UnsafeDelete: false,
				},
				Defaults:  &kuma_cp.Defaults{},
				Multizone: multizone.DefaultMultizoneConfig(),
			})
		unsafeDeleteResManager = mesh.NewMeshManager(
			resStore, manager, test_resources.Global(),
			validator, context.Background(),
			kuma_cp.Config{
				Store: &config_store.StoreConfig{
					Type: config_store.MemoryStore, UnsafeDelete: true,
				},
				Defaults:  &kuma_cp.Defaults{},
				Multizone: multizone.DefaultMultizoneConfig(),
			})
	})

	Describe("Create()", func() {
		It("should create default resources", func() {
			// given
			meshName := "mesh-1"

			// when
			err := samples.MeshDefaultBuilder().WithName(meshName).Create(resManager)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and Dataplane Token Signing Key for the mesh exists
			key := tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(meshName), tokens.DefaultKeyID, meshName)
			err = secretManager.Get(context.Background(), system.NewSecretResource(), store.GetBy(key))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not create mesh with name longer than 64 chars", func() {
			var name strings.Builder
			for range 64 {
				name.WriteString("x")
			}
			err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(name.String(), model.NoMesh))

			// then
			Expect(err).To(MatchError("name: cannot be longer than 63 characters"))
		})
	})

	Describe("Delete()", func() {
		It("should delete secrets within one mesh", func() {
			// given two meshes
			err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("demo-1", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("demo-2", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// when demo-1 is deleted
			err = resManager.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("demo-1", model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and all secrets are deleted
			secrets := &system.SecretResourceList{}
			err = secretManager.List(context.Background(), secrets, store.ListByMesh("demo-1"))
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets.Items).To(BeEmpty())

			// and all secrets from other mesh are preserved
			secrets = &system.SecretResourceList{}
			err = secretManager.List(context.Background(), secrets, store.ListByMesh("demo-2"))
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets.Items).To(HaveLen(1)) // default signing key
		})

		It("should not delete Mesh if there are Dataplanes attached", func() {
			// given mesh and dataplane
			Expect(samples.MeshDefaultBuilder().WithName("mesh-1").Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().WithMesh("mesh-1").Create(resStore)).To(Succeed())

			// when mesh-1 is delete
			err := resManager.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("mesh-1", model.NoMesh))
			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("mesh: unable to delete mesh, there are still some dataplanes attached"))
		})

		It("should delete Mesh if there are Dataplanes attached when unsafe delete is enabled", func() {
			// given mesh and dataplane
			Expect(samples.MeshDefaultBuilder().WithName("mesh-1").Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().WithMesh("mesh-1").Create(resStore)).To(Succeed())

			// when mesh-1 is deleted
			err := unsafeDeleteResManager.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("mesh-1", model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
