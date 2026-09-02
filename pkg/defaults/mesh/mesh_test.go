package mesh_test

import (
	"context"
	"errors"
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
			// given a combined default MeshTimeout as CP 2.14 and earlier stored
			// it: 'to' set, and the inbound half in 'from', a field the current
			// struct no longer has, so it's absent here (dropped on unmarshal).
			legacy := meshtimeout.NewMeshTimeoutResource()
			legacy.Spec = &meshtimeout.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{
					Kind: common_api.TopLevelTargetRefKindMesh,
				},
				To: &[]meshtimeout.To{
					{
						TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
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

			// then the legacy resource is migrated to 'rules' only, restored
			// from the current mesh-wide inbound defaults
			migrated := meshtimeout.NewMeshTimeoutResource()
			err = resManager.Get(context.Background(), migrated, core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(migrated.Spec.TargetRef).ToNot(BeNil())
			Expect(migrated.Spec.TargetRef.Kind).To(Equal(common_api.TopLevelTargetRefKindMesh))
			Expect(migrated.Spec.To).To(BeNil())
			Expect(migrated.Spec.Rules).ToNot(BeNil())
			Expect(*migrated.Spec.Rules).To(Equal(currentDefaultMeshTimeoutRules(resManager)))

			// and a separate 'to' resource now carries the outbound defaults
			toResource := meshtimeout.NewMeshTimeoutResource()
			err = resManager.Get(context.Background(), toResource, core_store.GetByKey("mesh-timeout-to-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(toResource.Spec.To).ToNot(BeNil())
			Expect(toResource.Spec.Rules).To(BeNil())
		})

		It("should migrate a legacy combined MeshTimeout default on the boot reconcile path", func() {
			// given the same legacy combined default as above
			legacy := meshtimeout.NewMeshTimeoutResource()
			legacy.Spec = &meshtimeout.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{
					Kind: common_api.TopLevelTargetRefKindMesh,
				},
				To: &[]meshtimeout.To{
					{
						TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
						Default:   meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Hour}},
					},
				},
			}
			err := rawStore.Create(context.Background(), legacy, core_store.CreateByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs with reconcileExistingOnly=true,
			// the mode used by the boot-time reconciliation loop that crashed in
			// production
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", true)
			Expect(err).ToNot(HaveOccurred())

			// then the legacy resource is migrated the same way
			migrated := meshtimeout.NewMeshTimeoutResource()
			err = resManager.Get(context.Background(), migrated, core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(migrated.Spec.To).To(BeNil())
			Expect(migrated.Spec.Rules).ToNot(BeNil())
			Expect(*migrated.Spec.Rules).To(Equal(currentDefaultMeshTimeoutRules(resManager)))

			toResource := meshtimeout.NewMeshTimeoutResource()
			err = resManager.Get(context.Background(), toResource, core_store.GetByKey("mesh-timeout-to-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(toResource.Spec.To).ToNot(BeNil())
		})

		It("should not fail EnsureDefaultMeshResources when a default-resource migration cannot complete", func() {
			// given a legacy combined default MeshTimeout that needs migrating
			legacy := meshtimeout.NewMeshTimeoutResource()
			legacy.Spec = &meshtimeout.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{
					Kind: common_api.TopLevelTargetRefKindMesh,
				},
				To: &[]meshtimeout.To{
					{
						TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
						Default:   meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Hour}},
					},
				},
			}
			err := rawStore.Create(context.Background(), legacy, core_store.CreateByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// and a resource manager whose Update always fails, simulating a
			// migration that can never complete (e.g. a persistently
			// unreachable store)
			failingManager := &failingUpdateResourceManager{ResourceManager: resManager}

			// when EnsureDefaultMeshResources runs
			// then it does not return an error: the migration is logged and skipped
			err = mesh.EnsureDefaultMeshResources(context.Background(), failingManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete legacy gateway-specific default MeshTimeout resources", func() {
			// given gateway-specific default MeshTimeout resources, as an older
			// CP version stored them before sidecar/gateway defaults were merged
			gatewayRules := meshtimeout.NewMeshTimeoutResource()
			gatewayRules.Spec = &meshtimeout.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
				Rules: &[]meshtimeout.Rule{
					{Default: meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}}},
				},
			}
			err := rawStore.Create(context.Background(), gatewayRules, core_store.CreateByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			gatewayTo := meshtimeout.NewMeshTimeoutResource()
			gatewayTo.Spec = &meshtimeout.MeshTimeout{
				TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
				To: &[]meshtimeout.To{
					{
						TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
						Default:   meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}},
					},
				},
			}
			err = rawStore.Create(context.Background(), gatewayTo, core_store.CreateByKey("mesh-gateways-timeout-to-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())

			// when EnsureDefaultMeshResources runs
			err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
			Expect(err).ToNot(HaveOccurred())

			// then the legacy gateway-specific defaults are gone
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-to-all-default", model.DefaultMesh))
			Expect(core_store.IsNotFound(err)).To(BeTrue())

			// and the single mesh-wide default still exists
			err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
			Expect(err).ToNot(HaveOccurred())
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

// currentDefaultMeshTimeoutRules returns the inbound rules a freshly-created
// mesh-timeout-all default carries today, used to assert a migrated legacy
// resource is restored to the same values.
func currentDefaultMeshTimeoutRules(resManager manager.ResourceManager) []meshtimeout.Rule {
	return currentDefaultMeshTimeoutRulesForStore(resManager, false, "")
}

func currentDefaultMeshTimeoutRulesForStore(resManager manager.ResourceManager, k8sStore bool, systemNamespace string) []meshtimeout.Rule {
	referenceMesh := core_mesh.NewMeshResource()
	err := resManager.Create(context.Background(), referenceMesh, core_store.CreateByKey("reference-defaults-mesh", model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, referenceMesh, []string{}, context.Background(), k8sStore, systemNamespace, config_core.Zone, "zone-1", false)
	Expect(err).ToNot(HaveOccurred())

	resourceName := "mesh-timeout-all-reference-defaults-mesh"
	if k8sStore {
		resourceName = resourceName + "." + systemNamespace
	}
	reference := meshtimeout.NewMeshTimeoutResource()
	err = resManager.Get(context.Background(), reference, core_store.GetByKey(resourceName, "reference-defaults-mesh"))
	Expect(err).ToNot(HaveOccurred())
	Expect(reference.Spec.Rules).ToNot(BeNil())
	return *reference.Spec.Rules
}

// failingUpdateResourceManager wraps a ResourceManager whose Update fails
// whenever it's asked to persist a migrated (rules-only) MeshTimeout,
// simulating a default-resource migration that can never complete without
// affecting unrelated Update calls (e.g. label reconciliation).
type failingUpdateResourceManager struct {
	manager.ResourceManager
}

func (f *failingUpdateResourceManager) Update(ctx context.Context, res model.Resource, opts ...core_store.UpdateOptionsFunc) error {
	if mt, ok := res.(*meshtimeout.MeshTimeoutResource); ok && mt.Spec.To == nil {
		return errors.New("simulated update failure")
	}
	return f.ResourceManager.Update(ctx, res, opts...)
}
