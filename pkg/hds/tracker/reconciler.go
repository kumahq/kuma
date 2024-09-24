package tracker

import (
	"context"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	"github.com/kumahq/kuma/pkg/hds/cache"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

type reconciler struct {
	hasher    util_xds_v3.NodeHash
	cache     util_xds_v3.SnapshotCache
	generator *SnapshotGenerator
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_config_core_v3.Node) error {
	newSnapshot, err := r.generator.GenerateSnapshot(ctx, node)
	if err != nil {
		return err
	}
	if err := newSnapshot.Consistent(); err != nil {
		return err
	}
	id := r.hasher.ID(node)
	var snap util_xds_v3.Snapshot
	oldSnapshot, _ := r.cache.GetSnapshot(id)
	switch {
	case oldSnapshot == nil:
		snap = newSnapshot
	case !util_xds_v3.SingleTypeSnapshotEqual(oldSnapshot, newSnapshot):
		snap = newSnapshot
	default:
		snap = oldSnapshot
	}
	return r.cache.SetSnapshot(id, snap)
}

func (r *reconciler) Clear(node *envoy_config_core_v3.Node) error {
	// cache.Clear() operation does not push a new (empty) configuration to Envoy.
	// That is why instead of calling cache.Clear() we set configuration to an empty Snapshot.
	// This fake value will be removed from cache on Envoy disconnect.
	return r.cache.SetSnapshot(r.hasher.ID(node), &cache.Snapshot{})
}
