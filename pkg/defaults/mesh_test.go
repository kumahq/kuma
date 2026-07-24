package defaults_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/defaults"
	defaults_mesh "github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	meshcircuitbreaker "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("EnsureDefaultMeshResourcesUpToDate", func() {
	var resManager core_manager.ResourceManager

	zoneConfig := func() kuma_cp.Config {
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = config_core.Zone
		cfg.Multizone.Zone.Name = "zone-1"
		return cfg
	}

	mcbKey := func(mesh string) core_store.GetOptionsFunc {
		return core_store.GetByKey("mesh-circuit-breaker-all-"+mesh, mesh)
	}

	// createMeshWithStaleDefaults persists a Mesh and its default policies as
	// an older CP version stored them: the MeshCircuitBreaker carries no labels.
	createMeshWithStaleDefaults := func(name string) {
		m := core_mesh.NewMeshResource()
		err := resManager.Create(context.Background(), m, core_store.CreateByKey(name, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = defaults_mesh.EnsureDefaultMeshResources(context.Background(), resManager, m, []string{}, context.Background(), false, "", config_core.Zone, "zone-1", false)
		Expect(err).ToNot(HaveOccurred())
		mcb := meshcircuitbreaker.NewMeshCircuitBreakerResource()
		Expect(resManager.Get(context.Background(), mcb, mcbKey(name))).ToNot(HaveOccurred())
		Expect(resManager.Update(context.Background(), mcb, core_store.UpdateWithLabels(map[string]string{}))).ToNot(HaveOccurred())
	}

	mcbZoneLabel := func(mesh string) string {
		mcb := meshcircuitbreaker.NewMeshCircuitBreakerResource()
		Expect(resManager.Get(context.Background(), mcb, mcbKey(mesh))).ToNot(HaveOccurred())
		return mcb.GetMeta().GetLabels()[mesh_proto.ZoneTag]
	}

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())
	})

	It("reconciles labels of default policies of every Mesh", func() {
		// given two Meshes whose default policies predate computed labels
		createMeshWithStaleDefaults(core_model.DefaultMesh)
		createMeshWithStaleDefaults("other")

		// when
		err := defaults.EnsureDefaultMeshResourcesUpToDate(context.Background(), resManager, logr.Discard(), zoneConfig(), context.Background())

		// then every Mesh's default policy carries 'kuma.io/zone'
		Expect(err).ToNot(HaveOccurred())
		Expect(mcbZoneLabel(core_model.DefaultMesh)).To(Equal("zone-1"))
		Expect(mcbZoneLabel("other")).To(Equal("zone-1"))
	})

	It("does not recreate a default policy deleted by an operator", func() {
		// given default policies exist and an operator deleted one
		createMeshWithStaleDefaults(core_model.DefaultMesh)
		err := resManager.Delete(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.DeleteByKey("mesh-circuit-breaker-all-"+core_model.DefaultMesh, core_model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = defaults.EnsureDefaultMeshResourcesUpToDate(context.Background(), resManager, logr.Discard(), zoneConfig(), context.Background())
		Expect(err).ToNot(HaveOccurred())

		// then the deleted policy stays absent
		err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), mcbKey(core_model.DefaultMesh))
		Expect(core_store.IsNotFound(err)).To(BeTrue())
	})

	It("skips a federated zone CP whose defaults come from Global", func() {
		// given a stale default policy and a federated zone CP config
		createMeshWithStaleDefaults(core_model.DefaultMesh)
		cfg := zoneConfig()
		cfg.Multizone.Zone.GlobalAddress = "grpcs://localhost:5685"

		// when
		err := defaults.EnsureDefaultMeshResourcesUpToDate(context.Background(), resManager, logr.Discard(), cfg, context.Background())
		Expect(err).ToNot(HaveOccurred())

		// then the stale policy is left untouched — the Global CP owns it
		Expect(mcbZoneLabel(core_model.DefaultMesh)).To(BeEmpty())
	})
})
