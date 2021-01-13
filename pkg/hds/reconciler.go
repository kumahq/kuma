package hds

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	"github.com/kumahq/kuma/pkg/hds/cache"
)

func NewReconciler(hasher cache.NodeHash, cache cache.SnapshotCache, generator *generator, versioner cache.SnapshotVersioner) *reconciler {
	return &reconciler{
		hasher:    hasher,
		cache:     cache,
		generator: generator,
		versioner: versioner,
	}
}

type reconciler struct {
	hasher    cache.NodeHash
	cache     cache.SnapshotCache
	generator *generator
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
