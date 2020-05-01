package server

import (
	"context"
	"sync"
	"time"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var (
	// overridable by unit tests
	now     = time.Now
	newUUID = core.NewUUID
)

type DataplaneStatusTracker interface {
	envoy_xds.Callbacks
	GetStatusAccessor(streamID int64) (SubscriptionStatusAccessor, bool)
}

type SubscriptionStatusAccessor interface {
	GetStatus() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription)
}

type DataplaneInsightSinkFactoryFunc = func(SubscriptionStatusAccessor) DataplaneInsightSink

func NewDataplaneStatusTracker(runtimeInfo core_runtime.RuntimeInfo,
	createStatusSink DataplaneInsightSinkFactoryFunc) DataplaneStatusTracker {
	return &dataplaneStatusTracker{
		runtimeInfo:      runtimeInfo,
		createStatusSink: createStatusSink,
		streams:          make(map[int64]*streamState),
	}
}

var _ DataplaneStatusTracker = &dataplaneStatusTracker{}

type dataplaneStatusTracker struct {
	runtimeInfo      core_runtime.RuntimeInfo
	createStatusSink DataplaneInsightSinkFactoryFunc
	mu               sync.RWMutex // protects access to the fields below
	streams          map[int64]*streamState
}

type streamState struct {
	stop         chan struct{} // is used for stopping a goroutine that flushes Dataplane status periodically
	mu           sync.RWMutex  // protects access to the fields below
	dataplaneId  core_model.ResourceKey
	subscription *mesh_proto.DiscoverySubscription
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *dataplaneStatusTracker) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	// initialize subscription
	subscription := &mesh_proto.DiscoverySubscription{
		Id:                     newUUID(),
		ControlPlaneInstanceId: c.runtimeInfo.GetInstanceId(),
		ConnectTime:            util_proto.MustTimestampProto(now()),
		Status:                 mesh_proto.NewSubscriptionStatus(),
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

	state := c.streams[streamID]

	delete(c.streams, streamID)

	// finilize subscription
	state.mu.Lock() // write access to the per Dataplane info
	subscription := state.subscription
	subscription.DisconnectTime = util_proto.MustTimestampProto(now())
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

	state := c.streams[streamID]

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// infer Dataplane id
	if state.dataplaneId == (core_model.ResourceKey{}) {
		if id, err := core_xds.ParseProxyId(req.Node); err == nil {
			state.dataplaneId = core_model.ResourceKey{Mesh: id.Mesh, Name: id.Name}
			// kick off async Dataplane status flusher
			go c.createStatusSink(state).Start(state.stop)
		} else {
			xdsServerLog.Error(err, "failed to parse Dataplane Id out of DiscoveryRequest", "streamid", streamID, "req", req)
		}
	}

	// update Dataplane status
	subscription := state.subscription
	if req.ResponseNonce != "" {
		subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(now())
		if req.ErrorDetail != nil {
			subscription.Status.Total.ResponsesRejected++
			subscription.Status.StatsOf(req.TypeUrl).ResponsesRejected++
		} else {
			subscription.Status.Total.ResponsesAcknowledged++
			subscription.Status.StatsOf(req.TypeUrl).ResponsesAcknowledged++
		}
	}

	xdsServerLog.V(1).Info("OnStreamRequest", "streamid", streamID, "request", req, "subscription", subscription)
	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (c *dataplaneStatusTracker) OnStreamResponse(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	state := c.streams[streamID]

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// update Dataplane status
	subscription := state.subscription
	subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(now())
	subscription.Status.Total.ResponsesSent++
	subscription.Status.StatsOf(resp.TypeUrl).ResponsesSent++

	xdsServerLog.V(1).Info("OnStreamResponse", "streamid", streamID, "request", req, "response", resp, "subscription", subscription)
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (c *dataplaneStatusTracker) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (c *dataplaneStatusTracker) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {}

func (c *dataplaneStatusTracker) GetStatusAccessor(streamID int64) (SubscriptionStatusAccessor, bool) {
	state, ok := c.streams[streamID]
	return state, ok
}

var _ SubscriptionStatusAccessor = &streamState{}

func (s *streamState) GetStatus() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription) {
	s.mu.RLock() // read access to the per Dataplane info
	defer s.mu.RUnlock()
	return s.dataplaneId, proto.Clone(s.subscription).(*mesh_proto.DiscoverySubscription)
}

func (s *streamState) Close() {
	close(s.stop)
}
