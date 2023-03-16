package reconcile

import (
	"context"
	"errors"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	cache_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
	util_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/util/xds"
)

var log = core.Log.WithName("kds-delta").WithName("reconcile")

func NewReconciler(
	hasher envoy_cache.NodeHash,
	cache envoy_cache.SnapshotCache,
	generator SnapshotGenerator,
	versioner cache_kds_v2.SnapshotVersioner,
	mode config_core.CpMode,
	statsCallbacks xds.StatsCallbacks,
) Reconciler {
	return &reconciler{
		hasher:         hasher,
		cache:          cache,
		generator:      generator,
		versioner:      versioner,
		mode:           mode,
		statsCallbacks: statsCallbacks,
	}
}

type reconciler struct {
	hasher         envoy_cache.NodeHash
	cache          envoy_cache.SnapshotCache
	generator      SnapshotGenerator
	versioner      cache_kds_v2.SnapshotVersioner
	mode           config_core.CpMode
	statsCallbacks xds.StatsCallbacks

	lock sync.Mutex
}

func (r *reconciler) Clear(node *envoy_core.Node) {
	id := r.hasher.ID(node)
	r.lock.Lock()
	defer r.lock.Unlock()
	snapshot, err := r.cache.GetSnapshot(id)
	if err != nil {
		return
	}
	r.cache.ClearSnapshot(id)
	if snapshot == nil {
		return
	}
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		r.statsCallbacks.DiscardConfig(snapshot.GetVersion(typ))
	}
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_core.Node) error {
	new, err := r.generator.GenerateSnapshot(ctx, node)
	if err != nil {
		return err
	}
	if new == nil {
		return errors.New("nil snapshot")
	}
	id := r.hasher.ID(node)
	old, _ := r.cache.GetSnapshot(id)
	new = r.versioner.Version(new, old)
	r.logChanges(new, old, node)
	r.meterConfigReadyForDelivery(new, old)
	return r.cache.SetSnapshot(ctx, id, new)
}

func (r *reconciler) logChanges(new envoy_cache.ResourceSnapshot, old envoy_cache.ResourceSnapshot, node *envoy_core.Node) {
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if old != nil && old.GetVersion(typ) != new.GetVersion(typ) {
			client := node.Id
			if r.mode == config_core.Zone {
				// we need to override client name because Zone is always a client to Global (on gRPC level)
				client = "global"
			}
			log.Info("detected changes in the resources. Sending changes to the client.", "resourceType", typ, "client", client)
		}
	}
}

func (r *reconciler) meterConfigReadyForDelivery(new envoy_cache.ResourceSnapshot, old envoy_cache.ResourceSnapshot) {
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if old == nil || old.GetVersion(typ) != new.GetVersion(typ) {
			r.statsCallbacks.ConfigReadyForDelivery(new.GetVersion(typ))
		}
	}
}
