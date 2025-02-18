package callbacks

import (
	"context"
	stdsync "sync"
	"sync/atomic"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

var dataplaneSyncTrackerLog = core.Log.WithName("xds").WithName("dataplane-sync-tracker")

func NewDataplaneSyncTracker(factoryFunc sync.DataplaneWatchdogFactory) DataplaneCallbacks {
	return &dataplaneSyncTracker{
		newDataplaneWatchdog: factoryFunc,
		watchdogs:            map[core_model.ResourceKey]entry{},
	}
}

var _ DataplaneCallbacks = &dataplaneSyncTracker{}

// dataplaneSyncTracker tracks XDS streams that are connected to the CP and fire up a watchdog.
// Watchdog should be run only once for given dataplane regardless of the number of streams.
// For ADS there is only one stream for DP.
//
// Node info can be (but does not have to be) carried only on the first XDS request. That's why need streamsAssociation map
// that indicates that the stream was already associated

type entry struct {
	cancelFunc context.CancelFunc
	meta       *atomic.Pointer[core_xds.DataplaneMetadata]
}
type dataplaneSyncTracker struct {
	NoopDataplaneCallbacks

	newDataplaneWatchdog sync.DataplaneWatchdogFactory

	stdsync.RWMutex // protects access to the fields below
	watchdogs       map[core_model.ResourceKey]entry
}

func (t *dataplaneSyncTracker) OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, _ context.Context, meta core_xds.DataplaneMetadata) error {
	// We use OnProxyConnected because there should be only one watchdog for given dataplane.
	t.Lock()
	defer t.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	t.watchdogs[dpKey] = entry{
		cancelFunc: func() {
			dataplaneSyncTrackerLog.V(1).Info("stopping Watchdog for a Dataplane", "dpKey", dpKey, "streamID", streamID)
			cancel()
		},
		meta: &atomic.Pointer[core_xds.DataplaneMetadata]{},
	}
	t.watchdogs[dpKey].meta.Store(&meta)
	dataplaneSyncTrackerLog.V(1).Info("starting Watchdog for a Dataplane", "dpKey", dpKey, "streamID", streamID)
	//nolint:contextcheck // it's not clear how the parent go-control-plane context lives
	go t.newDataplaneWatchdog.New(dpKey, t.watchdogs[dpKey].meta.Load).Start(ctx)
	return nil
}

func (t *dataplaneSyncTracker) OnProxyReconnected(_ core_xds.StreamID, dpKey core_model.ResourceKey, _ context.Context, meta core_xds.DataplaneMetadata) error {
	t.RLock()
	defer t.RUnlock()
	if e, ok := t.watchdogs[dpKey]; ok {
		e.meta.Store(&meta)
	}
	return nil
}

func (t *dataplaneSyncTracker) OnProxyDisconnected(_ context.Context, _ core_xds.StreamID, dpKey core_model.ResourceKey) {
	t.Lock()
	defer t.Unlock()
	if e, exists := t.watchdogs[dpKey]; exists {
		e.cancelFunc()
	}
	delete(t.watchdogs, dpKey)
}
