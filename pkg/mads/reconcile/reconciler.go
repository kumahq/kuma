package reconcile

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds.SnapshotCache,
	snapshotter Snapshotter, versioner util_xds.SnapshotVersioner) Reconciler {
	return &reconciler{
		hasher:      hasher,
		cache:       cache,
		snapshotter: snapshotter,
		versioner:   versioner,
	}
}

type reconciler struct {
	hasher      envoy_cache.NodeHash
	cache       util_xds.SnapshotCache
	snapshotter Snapshotter
	versioner   util_xds.SnapshotVersioner
}

func (r *reconciler) Reconcile(node *envoy_core.Node) error {
	new, err := r.snapshotter.Snapshot(node)
	if err != nil {
		return err
	}
	id := r.hasher.ID(node)
	old, _ := r.cache.GetSnapshot(id)
	r.versioner.Version(new, old)
	return r.cache.SetSnapshot(id, new)
}
