package reconcile

import (
	"context"
	"sync/atomic"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewReconciler(cache envoy_cache.SnapshotCache, generator *SnapshotGenerator) Reconciler {
	r := &reconciler{
		cache:          cache,
		generator:      generator,
		knownClientIds: atomic.Pointer[[]string]{},
	}
	base := []string{}
	r.knownClientIds.Store(&base)
	return r
}

type reconciler struct {
	cache          envoy_cache.SnapshotCache
	generator      *SnapshotGenerator
	knownClientIds atomic.Pointer[[]string]
}

func (r *reconciler) KnownClientIds() []string {
	p := r.knownClientIds.Load()
	return *p
}

func (r *reconciler) Reconcile(ctx context.Context) error {
	newSnapshotPerClient, err := r.generator.GenerateSnapshot(ctx)
	if err != nil {
		return err
	}
	knownClients := []string{}
	for clientId, newSnapshot := range newSnapshotPerClient {
		knownClients = append(knownClients, clientId)
		var snap envoy_cache.ResourceSnapshot
		oldSnapshot, _ := r.cache.GetSnapshot(clientId)
		switch {
		case oldSnapshot == nil:
			snap = newSnapshot
		case !util_xds_v3.SingleTypeSnapshotEqual(oldSnapshot, newSnapshot):
			snap = newSnapshot
		default:
			snap = oldSnapshot
		}
		err := r.cache.SetSnapshot(ctx, clientId, snap)
		if err != nil {
			return err
		}
	}

	r.knownClientIds.Store(&knownClients)
	return nil
}
