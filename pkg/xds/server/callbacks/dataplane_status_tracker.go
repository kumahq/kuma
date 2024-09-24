package callbacks

import (
	"context"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"

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
	util_xds.DeltaCallbacks
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
		deltaStreams:     make(map[int64]*streamState),
	}
}

var _ DataplaneStatusTracker = &dataplaneStatusTracker{}

type dataplaneStatusTracker struct {
	util_xds.NoopCallbacks
	runtimeInfo      core_runtime.RuntimeInfo
	createStatusSink DataplaneInsightSinkFactoryFunc
	mu               sync.RWMutex // protects access to the fields below
	streams          map[int64]*streamState
	deltaStreams     map[int64]*streamState
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
	return c.onStreamOpen(streamID, typ, false)
}

// OnDeltaStreamOpen is called once an Delta xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnDeltaStreamOpen will still be called.
func (c *dataplaneStatusTracker) OnDeltaStreamOpen(_ context.Context, streamID int64, typ string) error {
	return c.onStreamOpen(streamID, typ, true)
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (c *dataplaneStatusTracker) OnStreamClosed(streamID int64) {
	c.onStreamClose(streamID, false)
}

// OnDeltaStreamClosed is called immediately prior to closing an Delta xDS stream with a stream ID.
func (c *dataplaneStatusTracker) OnDeltaStreamClosed(streamID int64) {
	c.onStreamClose(streamID, true)
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *dataplaneStatusTracker) OnStreamRequest(streamID int64, req util_xds.DiscoveryRequest) error {
	return c.onStreamRequest(streamID, req, false)
}

// OnStreamDeltaRequest is called once a request is received on a delta stream.
// Returning an error will end processing and close the stream. OnStreamDeltaRequest will still be called.
func (c *dataplaneStatusTracker) OnStreamDeltaRequest(streamID int64, req util_xds.DeltaDiscoveryRequest) error {
	return c.onStreamRequest(streamID, req, true)
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (c *dataplaneStatusTracker) OnStreamResponse(streamID int64, req util_xds.DiscoveryRequest, resp util_xds.DiscoveryResponse) {
	c.onStreamResponse(streamID, req, resp, false)
}

// OnStreamDeltaResponse is called immediately prior to sending a response on a delta stream.
func (c *dataplaneStatusTracker) OnStreamDeltaResponse(streamID int64, req util_xds.DeltaDiscoveryRequest, resp util_xds.DeltaDiscoveryResponse) {
	c.onStreamResponse(streamID, req, resp, true)
}

// To keep logs short, we want to log "Listeners" instead of full qualified Envoy type url name
func shortEnvoyType(typeURL string) string {
	segments := strings.Split(typeURL, ".")
	if len(segments) <= 1 {
		return typeURL
	}
	return segments[len(segments)-1]
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

func (c *dataplaneStatusTracker) onStreamRequest(streamID int64, req util_xds.Request, isDelta bool) error {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	var state *streamState
	if isDelta {
		state = c.deltaStreams[streamID]
	} else {
		state = c.streams[streamID]
	}

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	if state.dataplaneId == (core_model.ResourceKey{}) {
		// Infer the Dataplane ID.
		if proxyId, err := core_xds.ParseProxyIdFromString(req.NodeId()); err == nil {
			state.dataplaneId = proxyId.ToResourceKey()
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

			log := statusTrackerLog.WithValues(
				"proxyName", state.dataplaneId.Name,
				"mesh", state.dataplaneId.Mesh,
				"streamID", streamID,
				"type", md.GetProxyType(),
				"dpVersion", md.GetVersion().GetKumaDp().GetVersion(),
				"subscriptionID", state.subscription.Id,
			)
			if statusTrackerLog.V(1).Enabled() {
				log = log.WithValues("node", req.Node())
			}
			log.Info("proxy connected")

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

	subscription := state.subscription
	log := statusTrackerLog.WithValues(
		"proxyName", state.dataplaneId.Name,
		"mesh", state.dataplaneId.Mesh,
		"streamID", streamID,
		"type", shortEnvoyType(req.GetTypeUrl()),
		"resourceVersion", req.VersionInfo(),
	)
	if statusTrackerLog.V(1).Enabled() {
		log = log.WithValues(
			"resourceNames", req.GetResourceNames(),
			"subscriptionID", subscription.Id,
			"nonce", req.GetResponseNonce(),
		)
	}

	// update Dataplane status
	if req.GetResponseNonce() != "" {
		subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(core.Now())
		if req.HasErrors() {
			log.Info("config rejected")
			subscription.Status.Total.ResponsesRejected++
			subscription.Status.StatsOf(req.GetTypeUrl()).ResponsesRejected++
		} else {
			log.V(1).Info("config accepted")
			subscription.Status.Total.ResponsesAcknowledged++
			subscription.Status.StatsOf(req.GetTypeUrl()).ResponsesAcknowledged++
		}
	} else {
		if !statusTrackerLog.V(1).Enabled() { // it was already added, no need to add it twice
			log = log.WithValues("resourceNames", req.GetResourceNames())
		}
		log.Info("config requested")
	}
	return nil
}

func (c *dataplaneStatusTracker) onStreamResponse(streamID int64, req util_xds.Request, resp util_xds.Response, isDelta bool) {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	var state *streamState
	if isDelta {
		state = c.deltaStreams[streamID]
	} else {
		state = c.streams[streamID]
	}

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// update Dataplane status
	subscription := state.subscription
	subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(core.Now())
	subscription.Status.Total.ResponsesSent++
	subscription.Status.StatsOf(resp.GetTypeUrl()).ResponsesSent++

	log := statusTrackerLog.WithValues(
		"proxyName", state.dataplaneId.Name,
		"mesh", state.dataplaneId.Mesh,
		"streamID", streamID,
		"type", shortEnvoyType(req.GetTypeUrl()),
		"resourceVersion", resp.VersionInfo(),
		"requestedResourceNames", req.GetResourceNames(),
		"resourceCount", resp.GetNumberOfResources(),
	)
	if statusTrackerLog.V(1).Enabled() {
		log = log.WithValues(
			"subscriptionID", subscription.Id,
			"nonce", resp.GetNonce(),
		)
	}

	log.V(1).Info("config sent")
}

func (c *dataplaneStatusTracker) onStreamOpen(streamID int64, typ string, isDelta bool) error {
	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	// initialize subscription
	now := core.Now()
	subscription := &mesh_proto.DiscoverySubscription{
		Id:                     core.NewUUID(),
		ControlPlaneInstanceId: c.runtimeInfo.GetInstanceId(),
		ConnectTime:            util_proto.MustTimestampProto(now),
		Status:                 mesh_proto.NewSubscriptionStatus(now),
		Version:                mesh_proto.NewVersion(),
	}
	// initialize state per ADS stream
	state := &streamState{
		stop:         make(chan struct{}),
		subscription: subscription,
	}
	// save
	if isDelta {
		c.deltaStreams[streamID] = state
	} else {
		c.streams[streamID] = state
	}

	statusTrackerLog.V(1).Info("proxy connecting", "streamID", streamID, "type", typ, "subscriptionID", subscription.Id)
	return nil
}

func (c *dataplaneStatusTracker) onStreamClose(streamID int64, isDelta bool) {
	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	var state *streamState
	if isDelta {
		state = c.deltaStreams[streamID]
	} else {
		state = c.streams[streamID]
	}
	if state == nil {
		statusTrackerLog.Info("[WARNING] proxy disconnected but no state in the status_tracker", "streamID", streamID)
		return
	}

	if isDelta {
		delete(c.deltaStreams, streamID)
	} else {
		delete(c.streams, streamID)
	}
	// finilize subscription
	state.mu.Lock() // write access to the per Dataplane info
	subscription := state.subscription
	subscription.DisconnectTime = util_proto.MustTimestampProto(core.Now())
	state.mu.Unlock()

	// trigger final flush
	state.Close()

	log := statusTrackerLog.WithValues(
		"streamID", streamID,
		"proxyName", state.dataplaneId.Name,
		"mesh", state.dataplaneId.Mesh,
		"subscriptionID", state.subscription.Id,
	)

	if statusTrackerLog.V(1).Enabled() {
		log = log.WithValues("subscription", subscription)
	}

	log.Info("proxy disconnected")
}
