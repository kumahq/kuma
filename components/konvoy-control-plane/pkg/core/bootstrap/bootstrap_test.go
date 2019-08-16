package bootstrap_test

import (
	"context"
	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/bootstrap"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bootstrap", func() {

	It("should create default mesh", func() {
		// given
		cfg := konvoy_cp.DefaultConfig()

		rt, err := bootstrap.Bootstrap(cfg)
		Expect(err).ToNot(HaveOccurred())

		// when
		rs := rt.ResourceStore()
		getOpts := core_store.GetByKey(core_model.DefaultNamespace,
			core_model.DefaultMesh, core_model.DefaultMesh)
		defaultMesh := mesh.MeshResource{}
		err = rs.Get(context.Background(), &defaultMesh, getOpts)
		Expect(err).ToNot(HaveOccurred())

		// then
		meshMeta := defaultMesh.GetMeta()
		Expect(meshMeta.GetName()).To(Equal("default"))
		Expect(meshMeta.GetMesh()).To(Equal("default"))
		Expect(meshMeta.GetNamespace()).To(Equal("default"))
	})

})
