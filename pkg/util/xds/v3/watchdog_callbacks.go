package v3

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
)

type NewNodeWatchdogFunc func(ctx context.Context, node *envoy_core.Node, streamId int64) (util_watchdog.Watchdog, error)

func NewWatchdogCallbacks(newNodeWatchdog NewNodeWatchdogFunc) envoy_xds.Callbacks {
	return &watchdogCallbacks{
		newNodeWatchdog: newNodeWatchdog,
		streams:         make(map[int64]watchdogStreamState),
	}
}

type watchdogCallbacks struct {
	NoopCallbacks
	newNodeWatchdog NewNodeWatchdogFunc

	mu      sync.RWMutex // protects access to the fields below
	streams map[int64]watchdogStreamState
}

type watchdogStreamState struct {
	context context.Context
	cancel  context.CancelFunc
}

var _ envoy_xds.Callbacks = &watchdogCallbacks{}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *watchdogCallbacks) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	cb.streams[streamID] = watchdogStreamState{
		context: ctx,
	}

	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *watchdogCallbacks) OnStreamClosed(streamID int64) {
	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	defer delete(cb.streams, streamID)

	if watchdog := cb.streams[streamID]; watchdog.cancel != nil {
		watchdog.cancel()
	}
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *watchdogCallbacks) OnStreamRequest(streamID int64, req *envoy_discovery.DiscoveryRequest) error {
	cb.mu.RLock() // read access to the map of all ADS streams
	watchdog := cb.streams[streamID]
	cb.mu.RUnlock()

	if watchdog.cancel != nil {
		return nil
	}

	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	// create a stop chanel even if there wan't be an actual watchdog
	stopCh := make(chan struct{})
	watchdog.cancel = context.CancelFunc(func() {
		close(stopCh)
	})
	cb.streams[streamID] = watchdog

	runnable, err := cb.newNodeWatchdog(watchdog.context, req.Node, streamID)
	if err != nil {
		return err
	}

	if runnable != nil {
		// kick off watchdog for that stream
		go runnable.Start(stopCh)
	}
	return nil
}
