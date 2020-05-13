package sync

import (
	"context"
	stdsync "sync"

	"github.com/Kong/kuma/pkg/core"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

var (
	dataplaneSyncTrackerLog = core.Log.WithName("xds-server").WithName("dataplane-sync-tracker")
)

type NewDataplaneWatchdogFunc func(dataplaneId core_model.ResourceKey, streamId core_xds.StreamID) util_watchdog.Watchdog

func NewDataplaneSyncTracker(factoryFunc NewDataplaneWatchdogFunc) envoy_xds.Callbacks {
	return &dataplaneSyncTracker{
		newDataplaneWatchdog: factoryFunc,
		streamsAssociation:   make(map[core_xds.StreamID]core_model.ResourceKey),
		dpStreams:            make(map[core_model.ResourceKey]streams),
	}
}

var _ envoy_xds.Callbacks = &dataplaneSyncTracker{}

type streams struct {
	watchdogCancel context.CancelFunc
	activeStreams  map[core_xds.StreamID]bool
}

// dataplaneSyncTracker tracks XDS streams that are connected to the CP and fire up a watchdog.
// Watchdog should be run only once for given dataplane regardless of the number of streams.
// For ADS there is only one stream for DP, but this is not the case with SDS
//
// Node info can be (but does not have to be) carried only on the first XDS request. That's why need streamsAssociation map
// that indicates that the stream was already associated
type dataplaneSyncTracker struct {
	newDataplaneWatchdog NewDataplaneWatchdogFunc

	stdsync.RWMutex    // protects access to the fields below
	streamsAssociation map[core_xds.StreamID]core_model.ResourceKey
	dpStreams          map[core_model.ResourceKey]streams
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (t *dataplaneSyncTracker) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, typ string) error {
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (t *dataplaneSyncTracker) OnStreamClosed(streamID core_xds.StreamID) {
	t.Lock()
	defer t.Unlock()

	dp, hasAssociation := t.streamsAssociation[streamID]
	if hasAssociation {
		delete(t.streamsAssociation, streamID)

		streams := t.dpStreams[dp]
		delete(streams.activeStreams, streamID)
		if len(streams.activeStreams) == 0 { // no stream is active, cancel watchdog
			if streams.watchdogCancel != nil {
				streams.watchdogCancel()
			}
			delete(t.dpStreams, dp)
		}
	}
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (t *dataplaneSyncTracker) OnStreamRequest(streamID core_xds.StreamID, req *envoy.DiscoveryRequest) error {
	t.RLock()
	_, alreadyAssociated := t.streamsAssociation[streamID]
	t.RUnlock()

	if alreadyAssociated {
		return nil
	}

	if id, err := core_xds.ParseProxyId(req.Node); err == nil {
		dataplaneKey := core_model.ResourceKey{Mesh: id.Mesh, Name: id.Name}
		t.Lock()
		defer t.Unlock()
		streams := t.dpStreams[dataplaneKey]
		if streams.activeStreams == nil {
			streams.activeStreams = map[core_xds.StreamID]bool{}
		}
		streams.activeStreams[streamID] = true
		if streams.watchdogCancel == nil { // watchdog was not started yet
			stopCh := make(chan struct{})
			streams.watchdogCancel = func() {
				close(stopCh)
			}
			// kick off watchdog for that Dataplane
			go t.newDataplaneWatchdog(dataplaneKey, streamID).Start(stopCh)
			dataplaneSyncTrackerLog.V(1).Info("started Watchdog for a Dataplane", "streamid", streamID, "proxyId", id, "dataplaneKey", dataplaneKey)
		}
		t.dpStreams[dataplaneKey] = streams
		t.streamsAssociation[streamID] = dataplaneKey
	} else {
		dataplaneSyncTrackerLog.Error(err, "failed to parse Dataplane Id out of DiscoveryRequest", "streamid", streamID, "req", req)
	}

	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (t *dataplaneSyncTracker) OnStreamResponse(streamID core_xds.StreamID, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (t *dataplaneSyncTracker) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (t *dataplaneSyncTracker) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}
