package defaults_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_component "github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/defaults"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Defaults Component", func() {

	Describe("when skip mesh creation is set to false", func() {

		var component core_component.Component
		var manager core_manager.ResourceManager

		BeforeEach(func() {
			cfg := &kuma_cp.Defaults{
				SkipMeshCreation: false,
			}
			store := resources_memory.NewStore()
			manager = core_manager.NewResourceManager(store)
			component = defaults.NewDefaultsComponent(cfg, core.Standalone, core.UniversalEnvironment, manager, store)
		})

		It("should create default mesh", func() {
			// when
			err := component.Start(nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			err = manager.Get(context.Background(), mesh_core.NewMeshResource(), core_store.GetByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not override already created mesh", func() {
			// given
			mesh := &mesh_core.MeshResource{
				Spec: &v1alpha1.Mesh{
					Mtls: &v1alpha1.Mesh_Mtls{
						EnabledBackend: "builtin",
						Backends: []*v1alpha1.CertificateAuthorityBackend{
							{
								Name: "builtin",
								Type: "builtin",
							},
						},
					},
				},
			}
			err := manager.Create(context.Background(), mesh, core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = component.Start(nil)

			// then
			mesh = mesh_core.NewMeshResource()
			Expect(err).ToNot(HaveOccurred())
			err = manager.Get(context.Background(), mesh, core_store.GetByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(mesh.Spec.Mtls.EnabledBackend).To(Equal("builtin"))
		})
	})

	Describe("when skip mesh creation is set to true", func() {

		var component core_component.Component
		var manager core_manager.ResourceManager

		BeforeEach(func() {
			cfg := &kuma_cp.Defaults{
				SkipMeshCreation: true,
			}
			store := resources_memory.NewStore()
			manager = core_manager.NewResourceManager(store)
			component = defaults.NewDefaultsComponent(cfg, core.Standalone, core.UniversalEnvironment, manager, store)
		})

		It("should not create default mesh", func() {
			// when
			err := component.Start(nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			err = manager.Get(context.Background(), mesh_core.NewMeshResource(), core_store.GetByKey("default", "default"))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})
	})

	Describe("zone ingress signing key creation", func() {

		var component core_component.Component
		var manager core_manager.ResourceManager

		BeforeEach(func() {
			cfg := &kuma_cp.Defaults{}
			store := resources_memory.NewStore()
			defaultManager := core_manager.NewResourceManager(store)
			customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
			customManagers[system.GlobalSecretType] = secret_manager.NewGlobalSecretManager(store, cipher.None())
			manager = core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
			component = defaults.NewDefaultsComponent(cfg, core.Standalone, core.UniversalEnvironment, manager, store)
		})

		It("should create zone ingress signing key and default mesh", func() {
			// when
			err := component.Start(nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			err = manager.Get(context.Background(), system.NewGlobalSecretResource(), core_store.GetByKey("zone-ingress-token-signing-key", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			err = manager.Get(context.Background(), mesh_core.NewMeshResource(), core_store.GetByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not override already created signing key", func() {
			// given
			signingKey := &system.GlobalSecretResource{
				Spec: &system_proto.Secret{
					Data: wrapperspb.Bytes([]byte("hello")),
				},
			}
			err := manager.Create(context.Background(), signingKey, core_store.CreateByKey("zone-ingress-token-signing-key", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = component.Start(nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			actual := system.NewGlobalSecretResource()
			err = manager.Get(context.Background(), actual, core_store.GetByKey("zone-ingress-token-signing-key", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Spec.GetData().GetValue()).To(Equal([]byte("hello")))
		})
	})
})
