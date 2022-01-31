package callbacks

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

var statusTrackerLog = core.Log.WithName("xds").WithName("status-tracker")

type DataplaneStatusTracker interface {
	util_xds.Callbacks
	GetStatusAccessor(streamID int64) (SubscriptionStatusAccessor, bool)
}

type SubscriptionStatusAccessor interface {
	GetStatus() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription)
}

type DataplaneInsightSinkFactoryFunc = func(core_model.ResourceType, SubscriptionStatusAccessor) DataplaneInsightSink

func NewDataplaneStatusTracker(
	runtimeInfo core_runtime.RuntimeInfo,
	createStatusSink DataplaneInsightSinkFactoryFunc,
) DataplaneStatusTracker {
	return &dataplaneStatusTracker{
		runtimeInfo:      runtimeInfo,
		createStatusSink: createStatusSink,
		streams:          make(map[int64]*streamState),
	}
}

var _ DataplaneStatusTracker = &dataplaneStatusTracker{}

type dataplaneStatusTracker struct {
	util_xds.NoopCallbacks
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
		Id:                     core.NewUUID(),
		ControlPlaneInstanceId: c.runtimeInfo.GetInstanceId(),
		ConnectTime:            util_proto.MustTimestampProto(core.Now()),
		Status:                 mesh_proto.NewSubscriptionStatus(),
		Version:                mesh_proto.NewVersion(),
	}
	// initialize state per ADS stream
	state := &streamState{
		stop:         make(chan struct{}),
		subscription: subscription,
	}
	// save
	c.streams[streamID] = state

	statusTrackerLog.V(1).Info("OnStreamOpen", "context", ctx, "streamid", streamID, "type", typ, "subscription", subscription)
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
	subscription.DisconnectTime = util_proto.MustTimestampProto(core.Now())
	state.mu.Unlock()

	// trigger final flush
	state.Close()

	statusTrackerLog.V(1).Info("OnStreamClosed", "streamid", streamID, "subscription", subscription)
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *dataplaneStatusTracker) OnStreamRequest(streamID int64, req util_xds.DiscoveryRequest) error {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	state := c.streams[streamID]

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	if state.dataplaneId == (core_model.ResourceKey{}) {
		var dpType core_model.ResourceType
		md := core_xds.DataplaneMetadataFromXdsMetadata(req.Metadata())

		// If the dataplane was started with a resource YAML, then it
		// will be serialized in the node metadata and we would know
		// the underlying type directly. Since that is optional, we
		// can't depend on it here, so we map from the proxy type,
		// which is guaranteed.
		switch md.GetProxyType() {
		case mesh_proto.IngressProxyType:
			dpType = core_mesh.ZoneIngressType
		case mesh_proto.DataplaneProxyType:
			dpType = core_mesh.DataplaneType
		case mesh_proto.EgressProxyType:
			dpType = core_mesh.ZoneEgressType
		}

		// Infer the Dataplane ID.
		if proxyId, err := core_xds.ParseProxyIdFromString(req.NodeId()); err == nil {
			state.dataplaneId = proxyId.ToResourceKey()
			if md.GetVersion() != nil {
				state.subscription.Version = md.GetVersion()
			} else {
				statusTrackerLog.Error(err, "failed to extract version out of the Envoy metadata", "streamid", streamID, "metadata", req.Metadata())
			}
			// Kick off the async Dataplane status flusher.
			go c.createStatusSink(dpType, state).Start(state.stop)
		} else {
			statusTrackerLog.Error(err, "failed to parse Dataplane Id out of DiscoveryRequest", "streamid", streamID, "req", req)
		}
	}

	// update Dataplane status
	subscription := state.subscription
	if req.GetResponseNonce() != "" {
		subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(core.Now())
		if req.HasErrors() {
			subscription.Status.Total.ResponsesRejected++
			subscription.Status.StatsOf(req.GetTypeUrl()).ResponsesRejected++
		} else {
			subscription.Status.Total.ResponsesAcknowledged++
			subscription.Status.StatsOf(req.GetTypeUrl()).ResponsesAcknowledged++
		}
	}

	statusTrackerLog.V(1).Info("OnStreamRequest", "streamid", streamID, "request", req, "subscription", subscription)
	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (c *dataplaneStatusTracker) OnStreamResponse(streamID int64, req util_xds.DiscoveryRequest, resp util_xds.DiscoveryResponse) {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	state := c.streams[streamID]

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// update Dataplane status
	subscription := state.subscription
	subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(core.Now())
	subscription.Status.Total.ResponsesSent++
	subscription.Status.StatsOf(resp.GetTypeUrl()).ResponsesSent++

	statusTrackerLog.V(1).Info("OnStreamResponse", "streamid", streamID, "request", req, "response", resp, "subscription", subscription)
}

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
