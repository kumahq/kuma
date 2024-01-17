package reconcile

import (
	"context"
	"errors"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
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

func (r *reconciler) Clear(ctx context.Context, node *envoy_core.Node) error {
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

	new, changed := r.Version(new, old)
	if changed {
		r.logChanges(logger, new, old, node)
		r.meterConfigReadyForDelivery(new, old, node.Id)
		return r.cache.SetSnapshot(ctx, id, new), true
	}
	return nil, false
}

func (r *reconciler) Version(new, old envoy_cache.ResourceSnapshot) (envoy_cache.ResourceSnapshot, bool) {
	if new == nil {
		return nil, false
	}
	changed := false
	newResources := map[core_model.ResourceType]envoy_cache.Resources{}
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		version := new.GetVersion(typ)
		if version != "" {
			// favor a version assigned by resource generator
			continue
		}

		if old != nil && r.equal(new.GetResources(typ), old.GetResources(typ)) {
			version = old.GetVersion(typ)
		}
		if version == "" {
			version = core.NewUUID()
			changed = true
		}
		if new.GetVersion(typ) == version {
			continue
		}
		n := map[string]envoy_types.ResourceWithTTL{}
		for k, v := range new.GetResourcesAndTTL(typ) {
			n[k] = v
		}
		newResources[core_model.ResourceType(typ)] = envoy_cache.Resources{Version: version, Items: n}
	}
	return &cache_v2.Snapshot{
		Resources: newResources,
	}, changed
}

func (_ *reconciler) equal(new, old map[string]envoy_types.Resource) bool {
	if len(new) != len(old) {
		return false
	}
	for key, newValue := range new {
		if oldValue, hasOldValue := old[key]; !hasOldValue || !proto.Equal(newValue, oldValue) {
			return false
		}
	}
	return true
}

func (r *reconciler) logChanges(logger logr.Logger, new envoy_cache.ResourceSnapshot, old envoy_cache.ResourceSnapshot, node *envoy_core.Node) {
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if old != nil && old.GetVersion(typ) != new.GetVersion(typ) {
			client := node.Id
			if r.mode == config_core.Zone {
				// we need to override client name because Zone is always a client to Global (on gRPC level)
				client = "global"
			}
			logger.Info("detected changes in the resources. Sending changes to the client.", "resourceType", typ, "client", client) // todo is client needed?
		}
	}
}

func (r *reconciler) meterConfigReadyForDelivery(new envoy_cache.ResourceSnapshot, old envoy_cache.ResourceSnapshot, nodeID string) {
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if old == nil || old.GetVersion(typ) != new.GetVersion(typ) {
			r.statsCallbacks.ConfigReadyForDelivery(nodeID + typ)
		}
	}
}
