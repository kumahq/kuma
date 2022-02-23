package defaults_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
			err = manager.Get(context.Background(), core_mesh.NewMeshResource(), core_store.GetByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not override already created mesh", func() {
			// given
			mesh := &core_mesh.MeshResource{
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
			mesh = core_mesh.NewMeshResource()
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
			err = manager.Get(context.Background(), core_mesh.NewMeshResource(), core_store.GetByKey("default", "default"))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})
	})
})
