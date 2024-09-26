package reconcile

import (
	"context"
	"maps"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds_v3.SnapshotCache, generator *SnapshotGenerator) Reconciler {
	return &reconciler{
		hasher:              hasher,
		cache:               cache,
		generator:           generator,
		knownClientIds:      map[string]bool{},
		knownClientIdsMutex: &sync.Mutex{},
		cacheMutex:          &sync.Mutex{},
	}
}

type reconciler struct {
	hasher              envoy_cache.NodeHash
	cache               util_xds_v3.SnapshotCache
	cacheMutex          *sync.Mutex
	generator           *SnapshotGenerator
	knownClientIds      map[string]bool
	knownClientIdsMutex *sync.Mutex
}

func (r *reconciler) KnownClientIds() map[string]bool {
	r.knownClientIdsMutex.Lock()
	defer r.knownClientIdsMutex.Unlock()
	return maps.Clone(r.knownClientIds)
}

func (r *reconciler) Reconcile(ctx context.Context) error {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	return r.reconcile(ctx)
}

func (r *reconciler) reconcile(ctx context.Context) error {
	newSnapshotPerClient, err := r.generator.GenerateSnapshot(ctx)
	if err != nil {
		return err
	}
	knownClients := map[string]bool{}
	for clientId, newSnapshot := range newSnapshotPerClient {
		knownClients[clientId] = true
		if err := newSnapshot.Consistent(); err != nil {
			return err
		}
		var snap util_xds_v3.Snapshot
		oldSnapshot, _ := r.cache.GetSnapshot(clientId)
		switch {
		case oldSnapshot == nil:
			snap = newSnapshot
		case !util_xds_v3.SingleTypeSnapshotEqual(oldSnapshot, newSnapshot):
			snap = newSnapshot
		default:
			snap = oldSnapshot
		}
		err := r.cache.SetSnapshot(clientId, snap)
		if err != nil {
			return err
		}
	}

	r.knownClientIdsMutex.Lock()
	r.knownClientIds = knownClients
	r.knownClientIdsMutex.Unlock()
	return nil
}

func (r *reconciler) ReconcileIfNeeded(ctx context.Context, node *envoy_core.Node) error {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	if r.NeedsReconciliation(node) {
		return r.reconcile(ctx)
	}
	return nil
}

func (r *reconciler) NeedsReconciliation(node *envoy_core.Node) bool {
	id := r.hasher.ID(node)
	return !r.cache.HasSnapshot(id)
}
