package mesh_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
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
	var rawStore core_store.ResourceStore
	var resManager manager.ResourceManager
	var defaultMesh *core_mesh.MeshResource

	BeforeEach(func() {
		rawStore = memory.NewStore()
		resManager = manager.NewResourceManager(rawStore)
		defaultMesh = core_mesh.NewMeshResource()

		err := resManager.Create(context.Background(), defaultMesh, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})
	Context("Default policy creation", func() {
		It("should create default resources in targetRef model", func() {
			// when
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())

			// then Dataplane Token Signing Key for the mesh exists
			err = resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
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
		})

		It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
			// given already ensured default resources
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "default", false)
			Expect(err).ToNot(HaveOccurred())
			// when ensuring again
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "default", false)
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
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"*"}, context.Background(), false, "", config_core.Zone, "default", false)
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
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{"MeshTimeout", "MeshRetry"}, context.Background(), false, "", config_core.Zone, "default", false)
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
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
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
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// and a default policy stripped of labels, as an older CP version stored it
			stale := meshcircuitbreaker.NewMeshCircuitBreakerResource()
			err = resManager.Get(context.Background(), stale, core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Update(context.Background(), stale, core_store.UpdateWithLabels(map[string]string{}))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs in reconcile-only mode
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", true)
			Expect(err).ToNot(HaveOccurred())

			// then the stale default policy is reconciled in place
			healed := meshcircuitbreaker.NewMeshCircuitBreakerResource()
			err = resManager.Get(context.Background(), healed, core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(healed.GetMeta().GetLabels()).To(HaveKeyWithValue(mesh_proto.ZoneTag, "zone-1"))
		})

		It("should migrate a legacy combined MeshTimeout default that predates the rules/to split", func() {
			// given a combined default MeshTimeout, as an older CP version stored it,
			// with both 'rules' and 'to' set (now mutually exclusive)
			legacy := meshtimeout.NewMeshTimeoutResource()
			legacy.Spec = &meshtimeout.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{
					Kind: common_api.Mesh,
					ProxyTypes: &[]common_api.TargetRefProxyType{
						common_api.Sidecar,
					},
				},
				Rules: &[]meshtimeout.Rule{
					{Default: meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Hour}}},
				},
				To: &[]meshtimeout.To{
					{
						TargetRef: common_api.OutboundTargetRef{Kind: common_api.Mesh},
						Default:   meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Hour}},
					},
				},
			}
			// bypass resManager.Create, which would validate against the current
			// (already-fixed) rules: writing directly to the store simulates data
			// an older CP already persisted before this constraint was restored.
			err := rawStore.Create(context.Background(), legacy, core_store.CreateByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// then the legacy resource is migrated to 'rules' only
			migrated := meshtimeout.NewMeshTimeoutResource()
			err = resManager.Get(context.Background(), migrated, core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(migrated.Spec.To).To(BeNil())
			Expect(migrated.Spec.Rules).ToNot(BeNil())

			// and a separate 'to' resource now carries the outbound defaults
			toResource := meshtimeout.NewMeshTimeoutResource()
			err = resManager.Get(context.Background(), toResource, core_store.GetByKey("mesh-timeout-to-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(toResource.Spec.To).ToNot(BeNil())
			Expect(toResource.Spec.Rules).To(BeNil())
		})

		It("should not recreate a default policy deleted by an operator", func() {
			// given default resources exist
			err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// and an operator deleted one of them
			err = resManager.Delete(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.DeleteByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs in reconcile-only mode
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", true)
			Expect(err).ToNot(HaveOccurred())

			// then the deleted policy stays absent
			err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and the remaining policies are untouched
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
