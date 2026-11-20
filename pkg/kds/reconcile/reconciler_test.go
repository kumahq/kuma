package reconcile_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_stream "github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds/reconcile"
	"github.com/kumahq/kuma/v3/pkg/kds/server"
	"github.com/kumahq/kuma/v3/pkg/kds/util"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/multitenant"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	util_xds "github.com/kumahq/kuma/v3/pkg/util/xds"
)

var _ = Describe("Reconciler", func() {
	var reconciler reconcile.Reconciler
	var store core_store.ResourceStore
	var snapshotCache envoy_cache.SnapshotCache
	hasher := &server.Hasher{}

	node := &envoy_core.Node{
		Id: "a",
	}
	changedTypes := map[core_model.ResourceType]struct{}{
		core_mesh.MeshType: {},
	}

	BeforeEach(func() {
		store = memory.NewStore()
		generator := reconcile.NewSnapshotGenerator(core_manager.NewResourceManager(store), reconcile.Any, reconcile.NoopResourceMapper)
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

	It("should clear the cache entry of a stream that closed before its first snapshot", func() {
		// given a watch opened before any snapshot exists
		sub := envoy_stream.NewDeltaSubscription([]string{"*"}, nil, nil, true)
		_, err := snapshotCache.CreateDeltaWatch(&envoy_sd.DeltaDiscoveryRequest{Node: node, TypeUrl: string(core_mesh.MeshType)}, sub, make(chan envoy_cache.DeltaResponse, 1))
		Expect(err).ToNot(HaveOccurred())
		Expect(snapshotCache.GetStatusKeys()).To(ConsistOf(node.Id))

		// when
		Expect(reconciler.Clear(node)).To(Succeed())

		// then
		Expect(snapshotCache.GetStatusKeys()).To(BeEmpty())
	})

	It("should clear only the cache entry of the stream that closed", func() {
		// given two streams of the same zone, each with a snapshot and a watch
		Expect(store.Create(context.Background(), samples.MeshDefault(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))).To(Succeed())
		streamNode := func(streamID int64) *envoy_core.Node {
			n := &envoy_core.Node{Id: node.Id}
			util.FillStreamMetadata(streamID, n)
			return n
		}
		oldStream, newStream := streamNode(1), streamNode(2)
		for _, n := range []*envoy_core.Node{oldStream, newStream} {
			err, changed := reconciler.Reconcile(context.Background(), n, changedTypes, logr.Discard())
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			sub := envoy_stream.NewDeltaSubscription([]string{"*"}, nil, nil, true)
			_, err = snapshotCache.CreateDeltaWatch(&envoy_sd.DeltaDiscoveryRequest{Node: n, TypeUrl: string(core_mesh.MeshType)}, sub, make(chan envoy_cache.DeltaResponse, 1))
			Expect(err).ToNot(HaveOccurred())
		}

		// when
		Expect(reconciler.Clear(oldStream)).To(Succeed())

		// then
		Expect(snapshotCache.GetStatusKeys()).To(ConsistOf(hasher.ID(newStream)))
		_, err := snapshotCache.GetSnapshot(hasher.ID(oldStream))
		Expect(err).To(HaveOccurred())
		_, err = snapshotCache.GetSnapshot(hasher.ID(newStream))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should build all types when there is no previous snapshot", func() {
		// given
		Expect(store.Create(context.Background(), samples.MeshDefault(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))).To(Succeed())

		// when only an unrelated type is reported as changed
		err, changed := reconciler.Reconcile(context.Background(), node, map[core_model.ResourceType]struct{}{core_mesh.DataplaneType: {}}, logr.Discard())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(changed).To(BeTrue())
		snapshot, err := snapshotCache.GetSnapshot(node.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(snapshot.GetResources(string(core_mesh.MeshType))).To(HaveLen(1))
	})
})
