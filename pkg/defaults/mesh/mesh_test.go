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
	meshcircuitbreaker "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshretry "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
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
	Context("When creating default routing resources is disabled", func() {
		It("should create default resources in targetRef model", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "")
			Expect(err).ToNot(HaveOccurred())

			// then Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())

			// and default TrafficPermission for the mesh doesn't exists
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default TrafficRoute for the mesh doesn't exists
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshRetry for the mesh exists
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default Gateway's MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshCircuitBreaker for the mesh exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
			// given already ensured default resources
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "")
			Expect(err).ToNot(HaveOccurred())
			// when ensuring again
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "")
			// then
			Expect(err).ToNot(HaveOccurred())

			// and all resources are in place
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating all default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"*"}, context.Background(), false, false, "")
			Expect(err).ToNot(HaveOccurred())

			// then default policies don't exist
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default Gateway's MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshCircuitBreaker for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating selected default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"MeshTimeout", "MeshRetry"}, context.Background(), false, false, "")
			Expect(err).ToNot(HaveOccurred())

			// then default MeshRetry doesn't exist
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exist
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default Gateway's MeshTimeout for the mesh doesn't exist
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshCircuitBreaker for the mesh does exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When creating default routing resources is enabled", func() {
		It("should create default resources", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true, false, "")
			Expect(err).ToNot(HaveOccurred())

			// then default TrafficPermission for the mesh exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default TrafficRoute for the mesh exists
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshRetry for the mesh exists
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default Gateway's MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshCircuitBreaker for the mesh exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
			// given already ensured default resources
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true, false, "")
			Expect(err).ToNot(HaveOccurred())
			// when ensuring again
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true, false, "")
			// then
			Expect(err).ToNot(HaveOccurred())

			// and all resources are in place
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating all default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"*"}, context.Background(), true, false, "")
			Expect(err).ToNot(HaveOccurred())

			// then default TrafficPermission doesn't exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// then default TrafficRoute doesn't exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshRetry for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and default MeshCircuitBreaker for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating selected default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"TrafficPermission", "MeshRetry"}, context.Background(), true, false, "")
			Expect(err).ToNot(HaveOccurred())

			// then default TrafficPermission doesn't exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// then default MeshRetry doesn't exist
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// then default TrafficRoute does exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshTimeout for the mesh does exist
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default Gateway's MeshTimeout for the mesh does exist
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-outbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-inbounds-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshCircuitBreaker for the mesh does exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
