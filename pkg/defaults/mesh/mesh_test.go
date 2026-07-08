package mesh_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	"github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	meshcircuitbreaker "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshretry "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
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
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())

			// then Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())

			// and default TrafficPermission for the mesh doesn't exists
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default TrafficRoute for the mesh doesn't exists
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshRetry for the mesh exists
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default Gateway's MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshCircuitBreaker for the mesh exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
			// given already ensured default resources
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())
			// when ensuring again
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "default", false)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and all resources are in place
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating all default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"*"}, context.Background(), false, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())

			// then default policies don't exist
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default Gateway's MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshCircuitBreaker for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating selected default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"MeshTimeout", "MeshRetry"}, context.Background(), false, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())

			// then default MeshRetry doesn't exist
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default Gateway's MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshCircuitBreaker for the mesh does exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Computed labels on default policies", func() {
		It("should set kuma.io/zone and kuma.io/origin on created default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// then a plugin-originated default policy carries zone/origin labels
			mcb := meshcircuitbreaker.NewMeshCircuitBreakerResource()
			err = resManager.Get(context.Background(), mcb, core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(mcb.GetMeta().GetLabels()).To(HaveKeyWithValue(mesh_proto.ZoneTag, "zone-1"))
			Expect(mcb.GetMeta().GetLabels()).To(HaveKeyWithValue(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin)))
		})

		It("should heal a default policy stored without labels by an older CP version", func() {
			// given default resources exist
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// and a default policy stripped of labels, as an older CP version stored it
			stale := meshcircuitbreaker.NewMeshCircuitBreakerResource()
			err = resManager.Get(context.Background(), stale, core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Update(context.Background(), stale, core_store.UpdateWithLabels(map[string]string{}))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs in reconcile-only mode
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "zone-1", true)
			Expect(err).ToNot(HaveOccurred())

			// then the stale default policy is reconciled in place
			healed := meshcircuitbreaker.NewMeshCircuitBreakerResource()
			err = resManager.Get(context.Background(), healed, core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(healed.GetMeta().GetLabels()).To(HaveKeyWithValue(mesh_proto.ZoneTag, "zone-1"))
		})

		It("should not recreate a default policy deleted by an operator", func() {
			// given default resources exist
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// and an operator deleted one of them
			err = resManager.Delete(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.DeleteByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs in reconcile-only mode
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, false, "", config_core.Zone, "zone-1", true)
			Expect(err).ToNot(HaveOccurred())

			// then the deleted policy stays absent
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and the remaining policies are untouched
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When creating default routing resources is enabled", func() {
		It("should create default resources", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true, false, "", config_core.Zone, "default", false)
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
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default Gateway's MeshTimeout for the mesh exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
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
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())
			// when ensuring again
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), true, false, "", config_core.Zone, "default", false)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and all resources are in place
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating all default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"*"}, context.Background(), true, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())

			// then default TrafficPermission doesn't exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// then default TrafficRoute doesn't exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshRetry for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshTimeout for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and default MeshCircuitBreaker for the mesh doesn't exists
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should skip creating selected default policies", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"TrafficPermission", "MeshRetry"}, context.Background(), true, false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())

			// then default TrafficPermission doesn't exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), core_store.GetByKey("allow-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// then default MeshRetry doesn't exist
			err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// then default TrafficRoute does exist
			err = resManager.Get(context.Background(), core_mesh.NewTrafficRouteResource(), core_store.GetByKey("route-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default MeshTimeout for the mesh does exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and default Gateway's MeshTimeout for the mesh does exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
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
