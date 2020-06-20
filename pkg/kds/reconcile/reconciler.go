package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds.SnapshotCache,
	generator SnapshotGenerator, versioner util_xds.SnapshotVersioner) Reconciler {
	return &reconciler{
		hasher:    hasher,
		cache:     cache,
		generator: generator,
		versioner: versioner,
	}
}

type reconciler struct {
	hasher    envoy_cache.NodeHash
	cache     util_xds.SnapshotCache
	generator SnapshotGenerator
	versioner util_xds.SnapshotVersioner
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_core.Node) error {
	new, err := r.generator.GenerateSnapshot(ctx, node)
	if err != nil {
		return err
	}
	if err := new.Consistent(); err != nil {
		return err
	}
	id := r.hasher.ID(node)
	old, _ := r.cache.GetSnapshot(id)
	new = r.versioner.Version(new, old)
	return r.cache.SetSnapshot(id, new)
}
