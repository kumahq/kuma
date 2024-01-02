package mesh_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/defaults/mesh"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

var _ = Describe("EnsureDefaultMeshResources", func() {
	var resManager manager.ResourceManager
	var defaultMesh *core_mesh.MeshResource

	BeforeEach(func() {
		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)
		defaultMesh = core_mesh.NewMeshResource()

		err := resManager.Create(context.Background(), defaultMesh, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create default secret resource", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false)
		Expect(err).ToNot(HaveOccurred())

		// and Dataplane Token Signing Key for the mesh exists
		err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should not create default policies", func() {
		// given already ensured default resources
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false)
		Expect(err).ToNot(HaveOccurred())

		// when ensuring again
		err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and all resources are not present
		err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		err = resManager.Get(context.Background(), core_mesh.NewRetryResource(), core_store.GetByKey("retry-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		err = resManager.Get(context.Background(), core_mesh.NewTimeoutResource(), core_store.GetByKey("timeout-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		err = resManager.Get(context.Background(), core_mesh.NewCircuitBreakerResource(), core_store.GetByKey("circuit-breaker-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
	})

	It("should create default resources", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true)
		Expect(err).ToNot(HaveOccurred())

		// then default TrafficPermission for the mesh exist
		err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// and default TrafficRoute for the mesh exists
		err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// and default Retry for the mesh exists
		err = resManager.Get(context.Background(), core_mesh.NewRetryResource(), core_store.GetByKey("retry-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// and Dataplane Token Signing Key for the mesh exists
		err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
		// given already ensured default resources
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true)
		Expect(err).ToNot(HaveOccurred())
		// when ensuring again
		err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true)
		// then
		Expect(err).ToNot(HaveOccurred())

		// and all resources are in place
		err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), core_mesh.NewRetryResource(), core_store.GetByKey("retry-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should skip creating all default policies", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"*"}, context.Background(), true)
		Expect(err).ToNot(HaveOccurred())

		// then default TrafficPermission doesn't exist
		err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

		// then default TrafficRoute doesn't exist
		err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

		// then default Retry doesn't exist
		err = resManager.Get(context.Background(), core_mesh.NewRetryResource(), core_store.GetByKey("retry-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

		// and Dataplane Token Signing Key for the mesh exists
		err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should skip creating selected default policies", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"TrafficPermission", "Retry"}, context.Background(), true)
		Expect(err).ToNot(HaveOccurred())

		// then default TrafficPermission doesn't exist
		err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

		// then default Retry doesn't exist
		err = resManager.Get(context.Background(), core_mesh.NewRetryResource(), core_store.GetByKey("retry-all-default", model.DefaultMesh))
		Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

		// then default TrafficRoute does exist
		err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// and Dataplane Token Signing Key for the mesh exists
		err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
		Expect(err).ToNot(HaveOccurred())
	})
})
