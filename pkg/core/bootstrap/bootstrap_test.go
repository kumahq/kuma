package bootstrap

import (
	"context"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	mesh_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	builtin_issuer "github.com/Kong/kuma/pkg/tokens/builtin/issuer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bootstrap", func() {

	It("should create default mesh", func() {
		// given
		cfg := kuma_cp.DefaultConfig()

		// when control plane is started
		rt, err := Bootstrap(cfg)
		ch := make(chan struct{})
		defer func() {
			close(ch)
		}()
		go func() {
			defer GinkgoRecover()
			err := rt.Start(ch)
			Expect(err).ToNot(HaveOccurred())
		}()
		Expect(err).ToNot(HaveOccurred())

		// then wait until resource is created
		resManager := rt.ResourceManager()
		Eventually(func() error {
			getOpts := core_store.GetByKey(core_model.DefaultNamespace, core_model.DefaultMesh, core_model.DefaultMesh)
			return resManager.Get(context.Background(), &mesh.MeshResource{}, getOpts)
		}, "5s").Should(Succeed())

		// when
		getOpts := core_store.GetByKey(core_model.DefaultNamespace, core_model.DefaultMesh, core_model.DefaultMesh)
		defaultMesh := mesh.MeshResource{}
		err = resManager.Get(context.Background(), &defaultMesh, getOpts)
		Expect(err).ToNot(HaveOccurred())

		// then
		meshMeta := defaultMesh.GetMeta()
		Expect(meshMeta.GetName()).To(Equal("default"))
		Expect(meshMeta.GetMesh()).To(Equal("default"))
		Expect(meshMeta.GetNamespace()).To(Equal("default"))
	})

	It("should skip creating mesh if one already exist", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		runtime, err := buildRuntime(cfg)
		Expect(err).ToNot(HaveOccurred())

		template := runtime.Config().Defaults.MeshProto()

		// when
		Expect(mesh_managers.CreateDefaultMesh(runtime.ResourceManager(), template, core_model.DefaultNamespace)).To(Succeed())

		// then mesh exists
		getOpts := core_store.GetByKey(core_model.DefaultNamespace, core_model.DefaultMesh, core_model.DefaultMesh)
		err = runtime.ResourceManager().Get(context.Background(), &mesh.MeshResource{}, getOpts)
		Expect(err).ToNot(HaveOccurred())

		// when createDefaultMesh is called once mesh already exist
		err = mesh_managers.CreateDefaultMesh(runtime.ResourceManager(), template, core_model.DefaultNamespace)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a default signing key", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		rt, err := Bootstrap(cfg)
		Expect(err).ToNot(HaveOccurred())

		// when
		key, err := builtin_issuer.GetSigningKey(rt.SecretManager())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(key).ToNot(HaveLen(0))

		// when kuma-cp is run again
		err = onStartup(rt)
		Expect(err).ToNot(HaveOccurred())
		key2, err := builtin_issuer.GetSigningKey(rt.SecretManager())

		// then it should skip creating a new signing key
		Expect(err).ToNot(HaveOccurred())
		Expect(key).To(Equal(key2))
	})
})
