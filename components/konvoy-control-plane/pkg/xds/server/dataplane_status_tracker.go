package server

import (
	"context"
	"sync"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

type DataplaneStatusTracker interface {
	envoy_xds.Callbacks
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) DataplaneStatusTracker {
	return NewDataplaneStatusTracker(rt, func() DataplaneStatusSink {
		return DefaultDataplaneStatusSink(rt.ResourceStore())
	})
}

type DataplaneStatusSinkFactoryFunc = func() DataplaneStatusSink

func NewDataplaneStatusTracker(runtimeInfo core_runtime.RuntimeInfo,
	createStatusSink DataplaneStatusSinkFactoryFunc) DataplaneStatusTracker {
	return &dataplaneStatusTracker{
		runtimeInfo:      runtimeInfo,
		createStatusSink: createStatusSink,
		streams:          make(map[int64]*streamState),
	}
}

var _ DataplaneStatusTracker = &dataplaneStatusTracker{}

type dataplaneStatusTracker struct {
	runtimeInfo      core_runtime.RuntimeInfo
	createStatusSink DataplaneStatusSinkFactoryFunc
	mu               sync.RWMutex // protects access to the fields below
	streams          map[int64]*streamState
}

type streamState struct {
	stop         chan struct{} // is used for stopping a goroutine that flushes Dataplane status periodically
	mu           sync.RWMutex  // protects access to the fields below
	dataplaneId  *core_model.ResourceKey
	subscription *mesh_proto.DiscoverySubscription
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *dataplaneStatusTracker) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	if typ != envoy_cache.AnyType {
		// status of a Dataplane can only be tracked as part of ADS subscription
		return nil
	}

	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	// initialize subscription
	subscription := &mesh_proto.DiscoverySubscription{
		Id:                     core.NewUUID(),
		ControlPlaneInstanceId: c.runtimeInfo.InstanceId(),
		ConnectTime:            types.TimestampNow(),
		Status:                 &mesh_proto.DiscoverySubscriptionStatus{},
	}
	// initialize state per ADS stream
	state := &streamState{
		stop:         make(chan struct{}),
		subscription: subscription,
	}
	// save
	c.streams[streamID] = state

	xdsServerLog.V(1).Info("OnStreamOpen", "context", ctx, "streamid", streamID, "type", typ, "subscription", subscription)
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (c *dataplaneStatusTracker) OnStreamClosed(streamID int64) {
	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	state, ok := c.streams[streamID]
	if !ok {
		// non ADS stream
		return
	}

	delete(c.streams, streamID)

	// finilize subscription
	state.mu.Lock() // write access to the per Dataplane info
	subscription := state.subscription
	subscription.DisconnectTime = types.TimestampNow()
	state.mu.Unlock()

	// trigger final flush
	state.Close()

	xdsServerLog.V(1).Info("OnStreamClosed", "streamid", streamID, "subscription", subscription)
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *dataplaneStatusTracker) OnStreamRequest(streamID int64, req *envoy.DiscoveryRequest) error {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	state, ok := c.streams[streamID]
	if !ok {
		// non ADS stream
		return nil
	}

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// infer Dataplane id
	if state.dataplaneId == nil {
		if id, err := core_xds.ParseDataplaneId(req.Node); err == nil {
			state.dataplaneId = id
			// kick off async Dataplane status flusher
			go c.createStatusSink().Start(state, state.stop)
		} else {
			xdsServerLog.Error(err, "failed to parse Dataplane Id out of DiscoveryRequest", "streamid", streamID, "req", req)
		}
	}

	// update Dataplane status
	subscription := state.subscription
	if req.ResponseNonce != "" {
		subscription.LastStatusUpdateTime = types.TimestampNow()
		if req.ErrorDetail != nil {
			subscription.Status.TotalResponsesRejected++
		} else {
			subscription.Status.TotalResponsesAcknowledged++
		}
		switch req.TypeUrl {
		case envoy_cache.ClusterType:
			if req.ErrorDetail != nil {
				subscription.Status.CdsResponsesRejected++
			} else {
				subscription.Status.CdsResponsesAcknowledged++
			}
		case envoy_cache.EndpointType:
			if req.ErrorDetail != nil {
				subscription.Status.EdsResponsesRejected++
			} else {
				subscription.Status.EdsResponsesAcknowledged++
			}
		case envoy_cache.ListenerType:
			if req.ErrorDetail != nil {
				subscription.Status.LdsResponsesRejected++
			} else {
				subscription.Status.LdsResponsesAcknowledged++
			}
		case envoy_cache.RouteType:
			if req.ErrorDetail != nil {
				subscription.Status.RdsResponsesRejected++
			} else {
				subscription.Status.RdsResponsesAcknowledged++
			}
		}
	}

	xdsServerLog.V(1).Info("OnStreamRequest", "streamid", streamID, "request", req, "subscription", subscription)
	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (c *dataplaneStatusTracker) OnStreamResponse(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	state, ok := c.streams[streamID]
	if !ok {
		// non ADS stream
		return
	}

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// update Dataplane status
	subscription := state.subscription
	subscription.LastStatusUpdateTime = types.TimestampNow()
	subscription.Status.TotalResponsesSent++
	switch resp.TypeUrl {
	case envoy_cache.ClusterType:
		subscription.Status.CdsResponsesSent++
	case envoy_cache.EndpointType:
		subscription.Status.EdsResponsesSent++
	case envoy_cache.ListenerType:
		subscription.Status.LdsResponsesSent++
	case envoy_cache.RouteType:
		subscription.Status.RdsResponsesSent++
	}

	xdsServerLog.V(1).Info("OnStreamResponse", "streamid", streamID, "request", req, "response", resp, "subscription", subscription)
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (c *dataplaneStatusTracker) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (c *dataplaneStatusTracker) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {}

var _ DataplaneStatusAccessor = &streamState{}

func (s *streamState) GetStatusSnapshot() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription) {
	s.mu.RLock() // read access to the per Dataplane info
	defer s.mu.RUnlock()
	return *s.dataplaneId, proto.Clone(s.subscription).(*mesh_proto.DiscoverySubscription)
}

func (s *streamState) Close() {
	close(s.stop)
}
