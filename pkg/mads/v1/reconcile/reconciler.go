package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache envoy_cache.SnapshotCache,
	generator util_xds_v3.SnapshotGenerator, versioner util_xds_v3.SnapshotVersioner,
) Reconciler {
	return &reconciler{
		hasher:    hasher,
		cache:     cache,
		generator: generator,
		versioner: versioner,
	}
}

type reconciler struct {
	hasher    envoy_cache.NodeHash
	cache     envoy_cache.SnapshotCache
	generator util_xds_v3.SnapshotGenerator
	versioner util_xds_v3.SnapshotVersioner
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_core.Node) error {
	newSnapshot, err := r.generator.GenerateSnapshot(ctx, node)
	if err != nil {
		return err
	}
	if err := newSnapshot.Consistent(); err != nil {
		return err
	}
	id := r.hasher.ID(node)
	old, _ := r.cache.GetSnapshot(id)
	newSnapshot = r.versioner.Version(newSnapshot, old)
	return r.cache.SetSnapshot(ctx, id, newSnapshot)
}

func (r *reconciler) NeedsReconciliation(node *envoy_core.Node) bool {
	id := r.hasher.ID(node)
	// when error returned there is no snapshot
	if _, err := r.cache.GetSnapshot(id); err != nil {
		return true
	}
	return false
}
