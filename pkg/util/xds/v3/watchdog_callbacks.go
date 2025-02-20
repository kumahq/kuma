package v3

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

// Watchdog used to run in the background and update at a fixed rhythm
type Watchdog interface {
	Start(ctx context.Context)
}

type NewNodeWatchdogFunc func(ctx context.Context, node *envoy_core.Node, streamId int64) (Watchdog, error)

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

// RestStreamID is used in the non-streaming REST context
const RestStreamID = int64(-1)

func (cb *watchdogCallbacks) hasStream(streamID int64) bool {
	cb.mu.RLock() // read access to the map of all ADS streams
	defer cb.mu.RUnlock()
	_, ok := cb.streams[streamID]
	return ok
}

func (cb *watchdogCallbacks) OnFetchRequest(ctx context.Context, req *envoy_discovery.DiscoveryRequest) error {
	// Open up a new "stream" state, which all REST requests use, if one doesn't already exist
	if cb.hasStream(RestStreamID) {
		return nil
	}

	if err := cb.OnStreamOpen(ctx, RestStreamID, req.TypeUrl); err != nil {
		return err
	}
	// TODO: could also register a TTL on the REST stream to clean it up if there is no activity over a certain period,
	// 		 since it will currently never be closed once opened
	//nolint:contextcheck // `OnStreamRequest` is a go-control-plane interface
	return cb.OnStreamRequest(RestStreamID, req)
}

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
func (cb *watchdogCallbacks) OnStreamClosed(streamID int64, node *envoy_core.Node) {
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

	// The request context shouldn't influence watchdogs.
	ctx, cancel := context.WithCancel(context.WithoutCancel(watchdog.context))
	watchdog.cancel = cancel
	cb.streams[streamID] = watchdog

	runnable, err := cb.newNodeWatchdog(ctx, req.Node, streamID)
	if err != nil {
		return err
	}

	if runnable != nil {
		// kick off watchdog for that stream
		go runnable.Start(ctx)
	}
	return nil
}

// OnDeltaStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnDeltaStreamClosed will still be called.
func (cb *watchdogCallbacks) OnDeltaStreamOpen(ctx context.Context, streamID int64, typ string) error {
	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	cb.streams[streamID] = watchdogStreamState{
		context: ctx,
	}

	return nil
}

// OnDeltaStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *watchdogCallbacks) OnDeltaStreamClosed(streamID int64, node *envoy_core.Node) {
	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	defer delete(cb.streams, streamID)

	if watchdog := cb.streams[streamID]; watchdog.cancel != nil {
		watchdog.cancel()
	}
}

// OnStreamDeltaRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnDeltaStreamClosed will still be called.
func (cb *watchdogCallbacks) OnStreamDeltaRequest(streamID int64, req *envoy_discovery.DeltaDiscoveryRequest) error {
	cb.mu.RLock() // read access to the map of all ADS streams
	watchdog := cb.streams[streamID]
	cb.mu.RUnlock()

	if watchdog.cancel != nil {
		return nil
	}

	cb.mu.Lock() // write access to the map of all ADS streams
	defer cb.mu.Unlock()

	ctx, cancel := context.WithCancel(context.WithoutCancel(watchdog.context))
	watchdog.cancel = cancel
	cb.streams[streamID] = watchdog

	runnable, err := cb.newNodeWatchdog(ctx, req.Node, streamID)
	if err != nil {
		return err
	}

	if runnable != nil {
		// kick off watchdog for that stream
		go runnable.Start(ctx)
	}
	return nil
}
