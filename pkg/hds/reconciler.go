package hds

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"

	"github.com/kumahq/kuma/pkg/hds/cache"
)

type reconciler struct {
	hasher    util_xds_v3.NodeHash
	cache     util_xds_v3.SnapshotCache
	generator *SnapshotGenerator
	versioner cache.SnapshotVersioner
}

func (r *reconciler) Reconcile(node *envoy_config_core_v3.Node) error {
	new, err := r.generator.GenerateSnapshot(node)
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
