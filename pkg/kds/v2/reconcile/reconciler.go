package reconcile

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
	util_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/util/xds"
)

func NewReconciler(hasher envoy_cache.NodeHash, cache envoy_cache.SnapshotCache, generator SnapshotGenerator, mode config_core.CpMode, statsCallbacks xds.StatsCallbacks, tenants multitenant.Tenants) Reconciler {
	return &reconciler{
		hasher:         hasher,
		cache:          cache,
		generator:      generator,
		mode:           mode,
		statsCallbacks: statsCallbacks,
		tenants:        tenants,
	}
}

type reconciler struct {
	hasher         envoy_cache.NodeHash
	cache          envoy_cache.SnapshotCache
	generator      SnapshotGenerator
	mode           config_core.CpMode
	statsCallbacks xds.StatsCallbacks
	tenants        multitenant.Tenants

	lock sync.Mutex
}

func (r *reconciler) Clear(node *envoy_core.Node) error {
	id := r.hasher.ID(node)
	r.lock.Lock()
	defer r.lock.Unlock()
	snapshot, err := r.cache.GetSnapshot(id)
	if err != nil {
		return nil // GetSnapshot returns an error if there is no snapshot. We don't need to error here
	}
	r.cache.ClearSnapshot(id)
	if snapshot == nil {
		return nil
	}
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		r.statsCallbacks.DiscardConfig(node.Id + typ)
	}
	return nil
}

func (r *reconciler) Reconcile(ctx context.Context, node *envoy_core.Node, changedTypes map[core_model.ResourceType]struct{}, logger logr.Logger) (error, bool) {
	id := r.hasher.ID(node)
	old, _ := r.cache.GetSnapshot(id)

	// construct builder with unchanged types from the old snapshot
	builder := cache_v2.NewSnapshotBuilder()
	if old != nil {
		for _, typ := range util_kds_v2.GetSupportedTypes() {
			resType := core_model.ResourceType(typ)
			if _, ok := changedTypes[resType]; ok {
				continue
			}

			oldRes := old.GetResources(typ)
			if len(oldRes) > 0 {
				builder = builder.With(resType, maps.Values(oldRes))
			}
		}
	}

	new, err := r.generator.GenerateSnapshot(ctx, node, builder, changedTypes)
	if err != nil {
		return err, false
	}
	if new == nil {
		return errors.New("nil snapshot"), false
	}
	// call ConstructVersionMap, so we can override versions if needed and compute what changed
	if old != nil {
		// this should already be computed by SetSnapshot, but we call it just to make sure we have versions.
		if err := old.ConstructVersionMap(); err != nil {
			return errors.Wrap(err, "could not construct version map"), false
		}
	}
	if err := new.ConstructVersionMap(); err != nil {
		return errors.Wrap(err, "could not construct version map"), false
	}

	if changed := r.changedTypes(old, new); len(changed) > 0 {
		r.logChanges(logger, changed, node)
		r.meterConfigReadyForDelivery(changed, node.Id)
		return r.cache.SetSnapshot(ctx, id, new), true
	}
	return nil, false
}

func (r *reconciler) changedTypes(old, new envoy_cache.ResourceSnapshot) []core_model.ResourceType {
	var changed []core_model.ResourceType
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if (old == nil && len(new.GetVersionMap(typ)) > 0) ||
			(old != nil && !maps.Equal(old.GetVersionMap(typ), new.GetVersionMap(typ))) {
			changed = append(changed, core_model.ResourceType(typ))
		}
	}
	return changed
}

func (r *reconciler) logChanges(logger logr.Logger, changedTypes []core_model.ResourceType, node *envoy_core.Node) {
	for _, typ := range changedTypes {
		client := node.Id
		if r.mode == config_core.Zone {
			// we need to override client name because Zone is always a client to Global (on gRPC level)
			client = "global"
		}
		logger.Info("detected changes in the resources. Sending changes to the client.", "resourceType", typ, "client", client) // todo is client needed?
	}
}

func (r *reconciler) meterConfigReadyForDelivery(changedTypes []core_model.ResourceType, nodeID string) {
	for _, typ := range changedTypes {
		r.statsCallbacks.ConfigReadyForDelivery(nodeID + string(typ))
	}
}
