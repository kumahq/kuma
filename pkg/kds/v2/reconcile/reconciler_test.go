package reconcile_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/kds/v2/server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

var _ = Describe("Reconciler", func() {
	var reconciler reconcile.Reconciler
	var store core_store.ResourceStore
	var snapshotCache envoy_cache.SnapshotCache

	node := &envoy_core.Node{
		Id: "a",
	}
	changedTypes := map[core_model.ResourceType]struct{}{
		core_mesh.MeshType: {},
	}

	BeforeEach(func() {
		store = memory.NewStore()
		generator := reconcile.NewSnapshotGenerator(core_manager.NewResourceManager(store), reconcile.Any, reconcile.NoopResourceMapper)
		hasher := &server.Hasher{}
		snapshotCache = envoy_cache.NewSnapshotCache(false, hasher, util_xds.NewLogger(logr.Discard()))
		metrics, err := core_metrics.NewMetrics("zone-1")
		Expect(err).ToNot(HaveOccurred())
		statsCallbacks, err := util_xds.NewStatsCallbacks(metrics, "kds_delta", util_xds.NoopVersionExtractor)
		Expect(err).ToNot(HaveOccurred())
		reconciler = reconcile.NewReconciler(hasher, snapshotCache, generator, config_core.Zone, statsCallbacks, multitenant.SingleTenant, []core_model.ResourceType{
			core_mesh.MeshType,
		})
	})

	It("should reconcile snapshot in snapshot cache", func() {
		// given
		Expect(store.Create(context.Background(), samples.MeshDefault(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))).To(Succeed())

		// when
		err, changed := reconciler.Reconcile(context.Background(), node, changedTypes, logr.Discard())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(changed).To(BeTrue())
		snapshot, err := snapshotCache.GetSnapshot(node.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(snapshot.GetResources(string(core_mesh.MeshType))).To(HaveLen(1))

		// when reconciled again without resource changes
		err, changed = reconciler.Reconcile(context.Background(), node, changedTypes, logr.Discard())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(changed).To(BeFalse())
		newSnapshot, err := snapshotCache.GetSnapshot(node.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(snapshot).To(BeIdenticalTo(newSnapshot))
	})
})
