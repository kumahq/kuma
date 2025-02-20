package callbacks

import (
	"context"
	stdsync "sync"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

var dataplaneSyncTrackerLog = core.Log.WithName("xds").WithName("dataplane-sync-tracker")

type NewDataplaneWatchdogFunc func(key core_model.ResourceKey, onDisconnectDone func(key core_model.ResourceKey)) util_xds_v3.Watchdog

func NewDataplaneSyncTracker(factoryFunc NewDataplaneWatchdogFunc) DataplaneCallbacks {
	return &dataplaneSyncTracker{
		newDataplaneWatchdog: factoryFunc,
		watchdogs:            map[core_model.ResourceKey]*trackerCleanup{},
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
	newDataplaneWatchdog NewDataplaneWatchdogFunc

	stdsync.RWMutex // protects access to the fields below
	watchdogs       map[core_model.ResourceKey]*trackerCleanup
}
type trackerCleanup struct {
	cancelFunc     context.CancelFunc
	disconnectDone chan<- struct{}
}

func (t *dataplaneSyncTracker) OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, _ context.Context, _ core_xds.DataplaneMetadata) error {
	// We use OnProxyConnected because there should be only one watchdog for given dataplane.
	t.Lock()
	defer t.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	t.watchdogs[dpKey] = &trackerCleanup{
		cancelFunc: func() {
			dataplaneSyncTrackerLog.V(1).Info("stopping Watchdog for a Dataplane", "dpKey", dpKey, "streamID", streamID)
			cancel()
		},
	}
	dataplaneSyncTrackerLog.V(1).Info("starting Watchdog for a Dataplane", "dpKey", dpKey, "streamID", streamID)
	//nolint:contextcheck // it's not clear how the parent go-control-plane context lives
	go t.newDataplaneWatchdog(dpKey, t.onWatchDogStopped).Start(ctx)
	return nil
}

func (t *dataplaneSyncTracker) OnProxyDisconnected(_ context.Context, _ core_xds.StreamID, dpKey core_model.ResourceKey, done chan<- struct{}) {
	t.RLock()
	defer t.RUnlock()
	if cleanup := t.watchdogs[dpKey]; cleanup != nil {
		cleanup.disconnectDone = done
		// kick off stopping of the watchdog
		cleanup.cancelFunc()
	}
}

func (t *dataplaneSyncTracker) onWatchDogStopped(dpKey core_model.ResourceKey) {
	t.Lock()
	defer t.Unlock()
	if cleanup := t.watchdogs[dpKey]; cleanup != nil && cleanup.disconnectDone != nil {
		cleanup.disconnectDone <- struct{}{}
	}
	delete(t.watchdogs, dpKey)
}
