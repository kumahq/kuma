package reconcile

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
<<<<<<< HEAD
	"google.golang.org/protobuf/proto"
=======
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
>>>>>>> 4752f7b82 (fix(kds): fix retry on NACK and add backoff (#9736))

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
	util_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/util/xds"
)

var log = core.Log.WithName("kds-delta").WithName("reconcile")

func NewReconciler(hasher envoy_cache.NodeHash, cache envoy_cache.SnapshotCache, generator SnapshotGenerator, mode config_core.CpMode, statsCallbacks xds.StatsCallbacks, tenants multitenant.Tenants) Reconciler {
	return &reconciler{
		hasher:         hasher,
		cache:          cache,
		generator:      generator,
		mode:           mode,
		statsCallbacks: statsCallbacks,
		tenants:        tenants,
		forceVersions:  map[string][]core_model.ResourceType{},
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

	forceVersions     map[string][]core_model.ResourceType
	forceVersionsLock sync.RWMutex
}

func (r *reconciler) ForceVersion(node *envoy_core.Node, resourceType core_model.ResourceType) {
	nodeID := r.hasher.ID(node)
	r.forceVersionsLock.Lock()
	r.forceVersions[nodeID] = append(r.forceVersions[nodeID], resourceType)
	r.forceVersionsLock.Unlock()
}

func (r *reconciler) Clear(ctx context.Context, node *envoy_core.Node) error {
	id := r.hasher.ID(node)
	r.lock.Lock()
	defer r.lock.Unlock()
	snapshot, err := r.cache.GetSnapshot(id)
	if err != nil {
		return err
	}
	r.cache.ClearSnapshot(id)
	if snapshot == nil {
		return nil
	}
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		r.statsCallbacks.DiscardConfig(snapshot.GetVersion(typ))
	}
	return nil
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
<<<<<<< HEAD
	new = r.Version(new, old)
	r.logChanges(new, old, node)
	r.meterConfigReadyForDelivery(new, old)
	return r.cache.SetSnapshot(ctx, id, new)
}

func (r *reconciler) Version(new, old envoy_cache.ResourceSnapshot) envoy_cache.ResourceSnapshot {
	if new == nil {
		return nil
	}
	newResources := map[core_model.ResourceType]envoy_cache.Resources{}
=======

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
	r.forceNewVersion(new, id)

	if changed := r.changedTypes(old, new); len(changed) > 0 {
		r.logChanges(logger, changed, node)
		r.meterConfigReadyForDelivery(changed, node.Id)
		return r.cache.SetSnapshot(ctx, id, new), true
	}
	return nil, false
}

func (r *reconciler) changedTypes(old, new envoy_cache.ResourceSnapshot) []core_model.ResourceType {
	var changed []core_model.ResourceType
>>>>>>> 4752f7b82 (fix(kds): fix retry on NACK and add backoff (#9736))
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if (old == nil && len(new.GetVersionMap(typ)) > 0) ||
			(old != nil && !maps.Equal(old.GetVersionMap(typ), new.GetVersionMap(typ))) {
			changed = append(changed, core_model.ResourceType(typ))
		}
<<<<<<< HEAD

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
=======
	}
	return changed
>>>>>>> 4752f7b82 (fix(kds): fix retry on NACK and add backoff (#9736))
}

// see kdsRetryForcer for more information
func (r *reconciler) forceNewVersion(snapshot envoy_cache.ResourceSnapshot, id string) {
	r.forceVersionsLock.Lock()
	forceVersionsForTypes := r.forceVersions[id]
	delete(r.forceVersions, id)
	r.forceVersionsLock.Unlock()
	for _, typ := range forceVersionsForTypes {
		cacheSnapshot, ok := snapshot.(*cache_v2.Snapshot)
		if !ok {
			panic("invalid type of Snapshot")
		}
<<<<<<< HEAD
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
=======
		for resourceName := range cacheSnapshot.VersionMap[typ] {
			cacheSnapshot.VersionMap[typ][resourceName] = ""
>>>>>>> 4752f7b82 (fix(kds): fix retry on NACK and add backoff (#9736))
		}
	}
}

<<<<<<< HEAD
func (r *reconciler) meterConfigReadyForDelivery(new envoy_cache.ResourceSnapshot, old envoy_cache.ResourceSnapshot) {
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		if old == nil || old.GetVersion(typ) != new.GetVersion(typ) {
			r.statsCallbacks.ConfigReadyForDelivery(new.GetVersion(typ))
=======
func (r *reconciler) logChanges(logger logr.Logger, changedTypes []core_model.ResourceType, node *envoy_core.Node) {
	for _, typ := range changedTypes {
		client := node.Id
		if r.mode == config_core.Zone {
			// we need to override client name because Zone is always a client to Global (on gRPC level)
			client = "global"
>>>>>>> 4752f7b82 (fix(kds): fix retry on NACK and add backoff (#9736))
		}
		logger.Info("detected changes in the resources. Sending changes to the client.", "resourceType", typ, "client", client) // todo is client needed?
	}
}

func (r *reconciler) meterConfigReadyForDelivery(changedTypes []core_model.ResourceType, nodeID string) {
	for _, typ := range changedTypes {
		r.statsCallbacks.ConfigReadyForDelivery(nodeID + string(typ))
	}
}
