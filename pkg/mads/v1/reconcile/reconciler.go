package reconcile

import (
	"context"
	"maps"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds_v3.SnapshotCache, generator *SnapshotGenerator) Reconciler {
	return &reconciler{
		hasher:              hasher,
		cache:               cache,
		generator:           generator,
		knownClientIds:      map[string]bool{},
		knownClientIdsMutex: &sync.Mutex{},
	}
}

type reconciler struct {
	hasher              envoy_cache.NodeHash
	cache               util_xds_v3.SnapshotCache
	generator           *SnapshotGenerator
	knownClientIds      map[string]bool
	knownClientIdsMutex *sync.Mutex
}

func (r *reconciler) KnownClientIds() map[string]bool {
	r.knownClientIdsMutex.Lock()
	defer r.knownClientIdsMutex.Unlock()
	return maps.Clone(r.knownClientIds)
}

func (r *reconciler) Reconcile(ctx context.Context, log logr.Logger) error {
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
			log.V(2).Info("no snapshot found", "clientId", clientId)
			snap = newSnapshot
		case !util_xds_v3.SingleTypeSnapshotEqual(oldSnapshot, newSnapshot):
			log.V(2).Info("detected changes in the snapshots", "oldSnapshot", oldSnapshot, "newSnapshot", newSnapshot)
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

func (r *reconciler) NeedsReconciliation(node *envoy_core.Node) bool {
	id := r.hasher.ID(node)
	return !r.cache.HasSnapshot(id)
}
