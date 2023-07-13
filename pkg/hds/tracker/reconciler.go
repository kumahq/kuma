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
	versioner util_xds_v3.SnapshotVersioner
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_config_core_v3.Node) error {
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

func (r *reconciler) Clear(node *envoy_config_core_v3.Node) error {
	// cache.Clear() operation does not push a new (empty) configuration to Envoy.
	// That is why instead of calling cache.Clear() we set configuration to an empty Snapshot.
	// This fake value will be removed from cache on Envoy disconnect.
	return r.cache.SetSnapshot(r.hasher.ID(node), &cache.Snapshot{})
}
