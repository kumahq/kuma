package defaults_test

import (
	"context"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_component "github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/defaults"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
			err = manager.Get(context.Background(), &mesh_core.MeshResource{}, core_store.GetByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not override already created mesh", func() {
			// given
			mesh := &mesh_core.MeshResource{
				Spec: v1alpha1.Mesh{
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
			err := manager.Create(context.Background(), mesh, core_store.CreateByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = component.Start(nil)

			// then
			mesh = &mesh_core.MeshResource{}
			Expect(err).ToNot(HaveOccurred())
			err = manager.Get(context.Background(), mesh, core_store.GetByKey("default", "default"))
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
			err = manager.Get(context.Background(), &mesh_core.MeshResource{}, core_store.GetByKey("default", "default"))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})
	})

	type testCase struct {
		cpMode       core.CpMode
		environment  core.EnvironmentType
		shouldCreate bool
	}
	DescribeTable("create signing key",
		func(given testCase) {
			// given
			store := resources_memory.NewStore()
			manager := core_manager.NewResourceManager(store)
			component := defaults.NewDefaultsComponent(&kuma_cp.Defaults{}, given.cpMode, given.environment, manager, store)
			err := manager.Create(context.Background(), &mesh_core.MeshResource{}, core_store.CreateByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = component.Start(nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			_, err = issuer.GetSigningKey(manager)
			if given.shouldCreate {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(Equal(issuer.SigningKeyNotFound))
			}
		},
		Entry("should succeed when mode is global and env is universal", testCase{
			cpMode:       core.Global,
			environment:  core.UniversalEnvironment,
			shouldCreate: true,
		}),
		Entry("should succeed when mode is global and env is kubernetes", testCase{
			cpMode:       core.Global,
			environment:  core.KubernetesEnvironment,
			shouldCreate: true,
		}),
		Entry("should succeed when mode is standalone and env is universal", testCase{
			cpMode:       core.Standalone,
			environment:  core.UniversalEnvironment,
			shouldCreate: true,
		}),
		Entry("should fail when mode is remote and env is universal", testCase{
			cpMode:       core.Remote,
			environment:  core.UniversalEnvironment,
			shouldCreate: false,
		}),
		Entry("should fail when mode is remote and env is kubernetes", testCase{
			cpMode:       core.Remote,
			environment:  core.KubernetesEnvironment,
			shouldCreate: false,
		}),
		Entry("should fail when mode is standalone and env is kubernetes", testCase{
			cpMode:       core.Standalone,
			environment:  core.KubernetesEnvironment,
			shouldCreate: false,
		}),
	)

})
