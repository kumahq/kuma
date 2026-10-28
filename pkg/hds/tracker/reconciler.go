package tracker

import (
	"context"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	
	util_xds_v3 "github.com/kumahq/kuma/v3/pkg/util/xds/v3"
)

type reconciler struct {
	hasher    envoy_cache.NodeHash
	cache     envoy_cache.SnapshotCache
	generator *SnapshotGenerator
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_config_core_v3.Node) error {
	newSnapshot, err := r.generator.GenerateSnapshot(ctx, node)
	if err != nil {
		return err
	}
	id := r.hasher.ID(node)
	var snap envoy_cache.ResourceSnapshot
	oldSnapshot, _ := r.cache.GetSnapshot(id)
	switch {
	case oldSnapshot == nil:
		snap = newSnapshot
	case !util_xds_v3.SingleTypeSnapshotEqual(newSnapshot, oldSnapshot):
		snap = newSnapshot
	default:
		snap = oldSnapshot
	}
	return r.cache.SetSnapshot(ctx, id, snap)
}

// Clear removes the snapshot and the status entry (which holds the full Envoy Node) of a Dataplane.
// It must only be called once the Dataplane has no active HDS streams, so there is no Envoy
// to push an empty configuration to. Nothing else removes the status entry: cancelling a watch
// only drops the watch, so without this every Dataplane ever connected stays in the cache.
func (r *reconciler) Clear(node *envoy_config_core_v3.Node) {
       r.cache.ClearSnapshot(r.hasher.ID(node))
}
