package tracker

import (
	"context"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	v3 "github.com/kumahq/kuma/pkg/hds/v3"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
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
	case !util_xds_v3.SingleTypeSnapshotEqual(oldSnapshot, newSnapshot):
		snap = newSnapshot
	default:
		snap = oldSnapshot
	}
	return r.cache.SetSnapshot(ctx, id, snap)
}

func (r *reconciler) Clear(ctx context.Context, node *envoy_config_core_v3.Node) error {
	// cache.Clear() operation does not push a new (empty) configuration to Envoy.
	// That is why instead of calling cache.Clear() we set configuration to an empty Snapshot.
	// This fake value will be removed from cache on Envoy disconnect.
	return r.cache.SetSnapshot(ctx, r.hasher.ID(node), util_xds_v3.NewSingleTypeSnapshot("", v3.HealthCheckSpecifierType, nil))
}
