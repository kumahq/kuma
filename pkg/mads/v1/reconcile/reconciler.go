package reconcile

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds_v3.SnapshotCache, generator *SnapshotGenerator, versioner util_xds_v3.SnapshotVersioner) Reconciler {
	return &reconciler{
		hasher:              hasher,
		cache:               cache,
		generator:           generator,
		versioner:           versioner,
		knownClientIds:      map[string]bool{},
		knownClientIdsMutex: &sync.Mutex{},
	}
}

type reconciler struct {
	hasher              envoy_cache.NodeHash
	cache               util_xds_v3.SnapshotCache
	generator           *SnapshotGenerator
	versioner           util_xds_v3.SnapshotVersioner
	knownClientIds      map[string]bool
	knownClientIdsMutex *sync.Mutex
}

func (r *reconciler) KnownClientIds() map[string]bool {
	r.knownClientIdsMutex.Lock()
	defer r.knownClientIdsMutex.Unlock()
	return util_maps.Clone(r.knownClientIds)
}

func (r *reconciler) Reconcile(ctx context.Context) error {
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
		old, _ := r.cache.GetSnapshot(clientId)
		newSnapshot = r.versioner.Version(newSnapshot, old)
		err := r.cache.SetSnapshot(clientId, newSnapshot)
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
