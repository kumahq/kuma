package reconcile

import (
	"context"
	"errors"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/proto"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
	util_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/util/xds"
)

var log = core.Log.WithName("kds-delta").WithName("reconcile")

func NewReconciler(hasher envoy_cache.NodeHash, cache envoy_cache.SnapshotCache, generator SnapshotGenerator, mode config_core.CpMode, statsCallbacks xds.StatsCallbacks, hashingFn multitenant.HashingFn) Reconciler {
	return &reconciler{
		hasher:         hasher,
		cache:          cache,
		generator:      generator,
		mode:           mode,
		statsCallbacks: statsCallbacks,
		hashingFn:      hashingFn,
	}
}

type reconciler struct {
	hasher         envoy_cache.NodeHash
	cache          envoy_cache.SnapshotCache
	generator      SnapshotGenerator
	mode           config_core.CpMode
	statsCallbacks xds.StatsCallbacks
	hashingFn      multitenant.HashingFn

	lock sync.Mutex
}

func (r *reconciler) Clear(ctx context.Context, node *envoy_core.Node) {
	id := r.hashId(ctx, node)
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
	id := r.hashId(ctx, node)
	old, _ := r.cache.GetSnapshot(id)
	new = r.Version(ctx, new, old)
	r.logChanges(new, old, node)
	r.meterConfigReadyForDelivery(new, old)
	return r.cache.SetSnapshot(ctx, id, new)
}

func (r *reconciler) Version(ctx context.Context, new, old envoy_cache.ResourceSnapshot) envoy_cache.ResourceSnapshot {
	if new == nil {
		return nil
	}
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
		}
		if new == nil {
			continue
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
	}
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

func (r *reconciler) hashId(ctx context.Context, node *envoy_core.Node) string {
	return r.hasher.ID(node) + ":" + r.hashingFn.ResourceHashKey(ctx)
}
