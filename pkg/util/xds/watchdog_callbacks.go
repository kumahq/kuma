package xds

import (
	"context"
	"sync"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"

	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"
)

type NewNodeWatchdogFunc func(node *envoy_core.Node, streamId int64) (util_watchdog.Watchdog, error)

func NewWatchdogCallbacks(newNodeWatchdog NewNodeWatchdogFunc) envoy_xds.Callbacks {
	return &watchdogCallbacks{
		newNodeWatchdog: newNodeWatchdog,
		streams:         make(map[int64]context.CancelFunc),
	}
}

type watchdogCallbacks struct {
	newNodeWatchdog NewNodeWatchdogFunc

	mu      sync.RWMutex // protects access to the fields below
	streams map[int64]context.CancelFunc
}

var _ envoy_xds.Callbacks = &watchdogCallbacks{}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *watchdogCallbacks) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *watchdogCallbacks) OnStreamClosed(streamID int64) {
	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	defer delete(cb.streams, streamID)

	if stopWatchdog := cb.streams[streamID]; stopWatchdog != nil {
		stopWatchdog()
	}
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *watchdogCallbacks) OnStreamRequest(streamID int64, req *envoy.DiscoveryRequest) error {
	cb.mu.RLock() // read access to the map of all ADS streams
	_, hasWatchdog := cb.streams[streamID]
	cb.mu.RUnlock()

	if hasWatchdog {
		return nil
	}

	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	// create a stop chanel even if there wan't be an actual watchdog
	stopCh := make(chan struct{})
	cb.streams[streamID] = context.CancelFunc(func() {
		close(stopCh)
	})

	watchdog, err := cb.newNodeWatchdog(req.Node, streamID)
	if err != nil {
		return err
	}

	if watchdog != nil {
		// kick off watchdag for that stream
		go watchdog.Start(stopCh)
	}

	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (cb *watchdogCallbacks) OnStreamResponse(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb *watchdogCallbacks) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (cb *watchdogCallbacks) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}
