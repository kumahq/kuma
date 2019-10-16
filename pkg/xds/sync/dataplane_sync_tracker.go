package sync

import (
	"context"
	stdsync "sync"

	"github.com/Kong/kuma/pkg/core"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

var (
	dataplaneSyncTrackerLog = core.Log.WithName("xds-server").WithName("dataplane-sync-tracker")
)

type NewDataplaneWatchdogFunc func(dataplaneId core_model.ResourceKey, streamId int64) util_watchdog.Watchdog

func NewDataplaneSyncTracker(factoryFunc NewDataplaneWatchdogFunc) envoy_xds.Callbacks {
	return &dataplaneSyncTracker{
		newDataplaneWatchdog: factoryFunc,
		streams:              make(map[int64]context.CancelFunc),
	}
}

var _ envoy_xds.Callbacks = &dataplaneSyncTracker{}

type dataplaneSyncTracker struct {
	newDataplaneWatchdog NewDataplaneWatchdogFunc

	mu      stdsync.RWMutex // protects access to the fields below
	streams map[int64]context.CancelFunc
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (t *dataplaneSyncTracker) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (t *dataplaneSyncTracker) OnStreamClosed(streamID int64) {
	t.mu.Lock() // write access to the map of all ADS streams
	defer t.mu.Unlock()

	defer delete(t.streams, streamID)

	if stopWatchdog := t.streams[streamID]; stopWatchdog != nil {
		stopWatchdog()
	}
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (t *dataplaneSyncTracker) OnStreamRequest(streamID int64, req *envoy.DiscoveryRequest) error {
	t.mu.RLock() // read access to the map of all ADS streams
	_, hasWatchdog := t.streams[streamID]
	t.mu.RUnlock()

	if hasWatchdog {
		return nil
	}

	t.mu.Lock() // write access to the map of all ADS streams
	defer t.mu.Unlock()

	if id, err := core_xds.ParseProxyId(req.Node); err == nil {
		dataplaneKey := core_model.ResourceKey{Mesh: id.Mesh, Namespace: id.Namespace, Name: id.Name}

		// kick off watchdag for that Dataplane
		stopCh := make(chan struct{})
		t.streams[streamID] = context.CancelFunc(func() {
			close(stopCh)
		})
		go t.newDataplaneWatchdog(dataplaneKey, streamID).Start(stopCh)
		dataplaneSyncTrackerLog.V(1).Info("started Watchdog for a Dataplane", "streamid", streamID, "proxyId", id, "dataplaneKey", dataplaneKey)
	} else {
		dataplaneSyncTrackerLog.Error(err, "failed to parse Dataplane Id out of DiscoveryRequest", "streamid", streamID, "req", req)
	}

	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (t *dataplaneSyncTracker) OnStreamResponse(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (t *dataplaneSyncTracker) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (t *dataplaneSyncTracker) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}
