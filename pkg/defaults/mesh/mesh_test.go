package mesh_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/defaults/mesh"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

var _ = Describe("EnsureDefaultMeshResources", func() {

	var resManager manager.ResourceManager

	BeforeEach(func() {
		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)

		err := resManager.Create(context.Background(), &core_mesh.MeshResource{}, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create default resources", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(resManager, model.DefaultMesh)
		Expect(err).ToNot(HaveOccurred())

		// then default TrafficPermission for the mesh exist
		err = resManager.Get(context.Background(), &core_mesh.TrafficPermissionResource{}, core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// and default TrafficRoute for the mesh exists
		err = resManager.Get(context.Background(), &core_mesh.TrafficRouteResource{}, core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// and Signing Key for the mesh exists
		err = resManager.Get(context.Background(), &system.SecretResource{}, core_store.GetBy(issuer.SigningKeyResourceKey(model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
		// given already ensured default resources
		err := mesh.EnsureDefaultMeshResources(resManager, model.DefaultMesh)
		Expect(err).ToNot(HaveOccurred())

		// when ensuring again
		err = mesh.EnsureDefaultMeshResources(resManager, model.DefaultMesh)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and all resources are in place
		err = resManager.Get(context.Background(), &core_mesh.TrafficPermissionResource{}, core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), &core_mesh.TrafficRouteResource{}, core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), &system.SecretResource{}, core_store.GetBy(issuer.SigningKeyResourceKey(model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})
})
