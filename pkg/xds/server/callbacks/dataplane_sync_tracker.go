package callbacks

import (
	"context"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/pkg/errors"
	stdsync "sync"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

var dataplaneSyncTrackerLog = core.Log.WithName("xds").WithName("dataplane-sync-tracker")

type NewDataplaneWatchdogFunc func(key core_model.ResourceKey) util_xds_v3.Watchdog

func NewDataplaneSyncTracker(factoryFunc NewDataplaneWatchdogFunc, hasher envoy_cache.NodeHash, cache envoy_cache.SnapshotCache) DataplaneCallbacks {
	return &dataplaneSyncTracker{
		newDataplaneWatchdog: factoryFunc,
		nodeHasher:           hasher,
		snapshotCache:        cache,
		watchdogs:            map[core_model.ResourceKey]context.CancelFunc{},
	}
}

var _ DataplaneCallbacks = &dataplaneSyncTracker{}

// dataplaneSyncTracker tracks XDS streams that are connected to the CP and fire up a watchdog.
// Watchdog should be run only once for given dataplane regardless of the number of streams.
// For ADS there is only one stream for DP.
//
// Node info can be (but does not have to be) carried only on the first XDS request. That's why need streamsAssociation map
// that indicates that the stream was already associated
type dataplaneSyncTracker struct {
	NoopDataplaneCallbacks

	newDataplaneWatchdog NewDataplaneWatchdogFunc

	stdsync.RWMutex // protects access to the fields below
	watchdogs       map[core_model.ResourceKey]context.CancelFunc
	nodeHasher      envoy_cache.NodeHash
	snapshotCache   envoy_cache.SnapshotCache
}

func (t *dataplaneSyncTracker) OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, _ core_xds.DataplaneMetadata) error {
	// We use OnProxyConnected because there should be only one watchdog for given dataplane.
	t.Lock()
	defer t.Unlock()

	// when there is an existing cache to be cleaned up, we wait for it to be removed
	if err := t.waitForCleaningUpExisting(dpKey, ctx, 3*time.Second); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.watchdogs[dpKey] = func() {
		dataplaneSyncTrackerLog.V(1).Info("stopping Watchdog for a Dataplane", "dpKey", dpKey, "streamID", streamID)
		cancel()
	}
	dataplaneSyncTrackerLog.V(1).Info("starting Watchdog for a Dataplane", "dpKey", dpKey, "streamID", streamID)
	//nolint:contextcheck // it's not clear how the parent go-control-plane context lives
	go t.newDataplaneWatchdog(dpKey).Start(ctx)
	return nil
}

func (t *dataplaneSyncTracker) OnProxyDisconnected(_ context.Context, _ core_xds.StreamID, dpKey core_model.ResourceKey) {
	t.Lock()
	defer t.Unlock()
	if cancelFn := t.watchdogs[dpKey]; cancelFn != nil {
		cancelFn()
	}
	delete(t.watchdogs, dpKey)
}

func (t *dataplaneSyncTracker) waitForCleaningUpExisting(dpKey core_model.ResourceKey, ctx context.Context, maxWait time.Duration) error {
	timeout := time.After(maxWait)
	timer := time.NewTicker(50 * time.Millisecond)
	defer timer.Stop()

	if t.nodeHasher == nil || t.snapshotCache == nil {
		return nil
	}

	proxyId := core_xds.FromResourceKey(dpKey)
	nodeCacheKey := t.nodeHasher.ID(&envoy_core.Node{Id: proxyId.String()})
	cacheExists := func() error {
		snapshot, err := t.snapshotCache.GetSnapshot(nodeCacheKey)
		if err != nil || snapshot == nil {
			return nil
		}
		return errors.New("dataplane is still in the cache")
	}

	if cacheExists() == nil {
		return nil
	}

	for {
		select {
		case <-timer.C:
			if cacheExists() == nil {
				return nil
			}
		case <-timeout:
			// if the cache is still there, this means there is a performance issue on the CP
			// in this case, we kick out the connection and the DPP will reconnect later
			return errors.Errorf("timeout while waiting for dataplane %s to be removed from the cache", proxyId)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
