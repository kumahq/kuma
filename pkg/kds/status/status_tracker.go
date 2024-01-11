package status

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type StatusTracker interface {
	envoy_xds.Callbacks
	GetStatusAccessor(streamID int64) (StatusAccessor, bool)
}

type StatusAccessor interface {
	GetStatus() (string, *system_proto.KDSSubscription)
}

type ZoneInsightSinkFactoryFunc = func(StatusAccessor, logr.Logger) ZoneInsightSink

func NewStatusTracker(runtimeInfo core_runtime.RuntimeInfo,
	createStatusSink ZoneInsightSinkFactoryFunc, log logr.Logger,
) StatusTracker {
	return &statusTracker{
		runtimeInfo:      runtimeInfo,
		createStatusSink: createStatusSink,
		streams:          make(map[int64]*streamState),
		log:              log,
	}
}

type statusTracker struct {
	util_xds_v3.NoopCallbacks
	runtimeInfo      core_runtime.RuntimeInfo
	createStatusSink ZoneInsightSinkFactoryFunc
	mu               sync.RWMutex // protects access to the fields below
	streams          map[int64]*streamState
	log              logr.Logger
}

type streamState struct {
	stop         chan struct{} // is used for stopping a goroutine that flushes Dataplane status periodically
	mu           sync.RWMutex  // protects access to the fields below
	zone         string
	subscription *system_proto.KDSSubscription
	ctx          context.Context
}

func (c *statusTracker) onStreamOpen(ctx context.Context, streamID int64, typ string) error {
	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	// initialize subscription
	now := core.Now()
	subscription := &system_proto.KDSSubscription{
		Id:                core.NewUUID(),
		ConnectTime:       util_proto.MustTimestampProto(now),
		Status:            system_proto.NewSubscriptionStatus(now),
		Version:           system_proto.NewVersion(),
		AuthTokenProvided: len(md.Get("authorization")) == 1,
	}
	switch c.runtimeInfo.GetMode() {
	case config_core.Global:
		subscription.GlobalInstanceId = c.runtimeInfo.GetInstanceId()
	case config_core.Zone:
		subscription.ZoneInstanceId = c.runtimeInfo.GetInstanceId()
	}
	// initialize state per ADS stream
	state := &streamState{
		stop:         make(chan struct{}),
		subscription: subscription,
		ctx:          ctx,
	}
	// save
	c.streams[streamID] = state

	c.log.V(1).Info("onStreamOpen", "context", ctx, "streamid", streamID, "type", typ, "subscription", subscription)
	return nil
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *statusTracker) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	return c.onStreamOpen(ctx, streamID, typ)
}

// OnDeltaStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnDeltaStreamClosed will still be called.
func (c *statusTracker) OnDeltaStreamOpen(ctx context.Context, streamID int64, typ string) error {
	return c.onStreamOpen(ctx, streamID, typ)
}

func (c *statusTracker) onStreamClosed(streamID int64, _ *envoy_core.Node) {
	c.mu.Lock() // write access to the map of all ADS streams
	defer c.mu.Unlock()

	state := c.streams[streamID]
	if state == nil {
		c.log.Info("[WARNING] OnStreamClosed but no state in the status_tracker", "streamid", streamID)
		return
	}

	delete(c.streams, streamID)

	// finilize subscription
	state.mu.Lock() // write access to the per Dataplane info
	subscription := state.subscription
	subscription.DisconnectTime = util_proto.MustTimestampProto(core.Now())
	state.mu.Unlock()

	// trigger final flush
	state.Close()

	c.log.V(1).Info("onStreamClosed", "streamid", streamID, "subscription", subscription)
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (c *statusTracker) OnStreamClosed(streamID int64, node *envoy_core.Node) {
	c.onStreamClosed(streamID, node)
}

// OnDeltaStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (c *statusTracker) OnDeltaStreamClosed(streamID int64, node *envoy_core.Node) {
	c.onStreamClosed(streamID, node)
}

type DiscoveryRequestInfo interface {
	GetNode() *envoy_core.Node
	GetTypeUrl() string
	GetResponseNonce() string
	GetErrorDetail() *status.Status
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *statusTracker) onStreamRequest(streamID int64, req DiscoveryRequestInfo) error {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()
	node := req.GetNode()
	if node == nil {
		return errors.New("Node not set, this should never happen")
	}

	state := c.streams[streamID]

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// infer zone
	if state.zone == "" {
		state.zone = node.Id
		if err := readVersion(node.GetMetadata(), state.subscription.Version); err != nil {
			c.log.Error(err, "failed to extract version out of the Envoy metadata", "streamid", streamID, "metadata", node.GetMetadata())
		}
		go c.createStatusSink(state, c.log).Start(state.ctx, state.stop)
	}

	// update Dataplane status
	subscription := state.subscription
	if req.GetResponseNonce() != "" {
		subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(core.Now())
		if req.GetErrorDetail() != nil {
			subscription.Status.Total.ResponsesRejected++
			util.StatsOf(subscription.Status, model.ResourceType(req.GetTypeUrl())).ResponsesRejected++
		} else {
			subscription.Status.Total.ResponsesAcknowledged++
			util.StatsOf(subscription.Status, model.ResourceType(req.GetTypeUrl())).ResponsesAcknowledged++
		}
	}
	remoteInstanceId := ""
	if node.Metadata != nil {
		if subscription.Config == "" && node.Metadata.Fields[kds.MetadataFieldConfig] != nil {
			subscription.Config = node.Metadata.Fields[kds.MetadataFieldConfig].GetStringValue()
		}
		if node.Metadata.Fields[kds.MetadataControlPlaneId] != nil {
			remoteInstanceId = node.Metadata.Fields[kds.MetadataControlPlaneId].GetStringValue()
		}
	}
	switch c.runtimeInfo.GetMode() {
	case config_core.Global:
		subscription.GlobalInstanceId = c.runtimeInfo.GetInstanceId()
		if remoteInstanceId != "" {
			subscription.ZoneInstanceId = remoteInstanceId
		}
	case config_core.Zone:
		subscription.ZoneInstanceId = c.runtimeInfo.GetInstanceId()
		if remoteInstanceId != "" {
			subscription.GlobalInstanceId = remoteInstanceId
		}
	}

	c.log.V(1).Info("OnStreamRequest", "streamid", streamID, "request", req, "subscription", subscription)
	return nil
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (c *statusTracker) OnStreamRequest(streamID int64, req *envoy_sd.DiscoveryRequest) error {
	return c.onStreamRequest(streamID, req)
}

// OnStreamDeltaRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnDeltaStreamClosed will still be called.
func (c *statusTracker) OnStreamDeltaRequest(streamID int64, req *envoy_sd.DeltaDiscoveryRequest) error {
	return c.onStreamRequest(streamID, req)
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (c *statusTracker) onStreamResponse(streamID int64, req DiscoveryRequestInfo, resp interface{}) {
	c.mu.RLock() // read access to the map of all ADS streams
	defer c.mu.RUnlock()

	state := c.streams[streamID]

	state.mu.Lock() // write access to the per Dataplane info
	defer state.mu.Unlock()

	// update Dataplane status
	subscription := state.subscription
	subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(core.Now())
	subscription.Status.Total.ResponsesSent++
	util.StatsOf(subscription.Status, model.ResourceType(req.GetTypeUrl())).ResponsesSent++

	c.log.V(1).Info("OnStreamResponse", "streamid", streamID, "request", req, "response", resp, "subscription", subscription)
}

func (c *statusTracker) OnStreamResponse(_ context.Context, streamID int64, req *envoy_sd.DiscoveryRequest, resp *envoy_sd.DiscoveryResponse) {
	c.onStreamResponse(streamID, req, resp)
}

// OnStreamDeltaResponse is called immediately prior to sending a response on a stream.
func (c *statusTracker) OnStreamDeltaResponse(streamID int64, req *envoy_sd.DeltaDiscoveryRequest, resp *envoy_sd.DeltaDiscoveryResponse) {
	c.onStreamResponse(streamID, req, resp)
}

func (c *statusTracker) GetStatusAccessor(streamID int64) (StatusAccessor, bool) {
	state, ok := c.streams[streamID]
	return state, ok
}

var _ StatusAccessor = &streamState{}

func (s *streamState) GetStatus() (string, *system_proto.KDSSubscription) {
	s.mu.RLock() // read access to the per Dataplane info
	defer s.mu.RUnlock()
	return s.zone, proto.Clone(s.subscription).(*system_proto.KDSSubscription)
}

func (s *streamState) Close() {
	close(s.stop)
}

func readVersion(metadata *structpb.Struct, version *system_proto.Version) error {
	if metadata == nil {
		return nil
	}
	rawVersion := metadata.Fields[kds.MetadataFieldVersion].GetStructValue()
	if rawVersion != nil {
		err := util_proto.ToTyped(rawVersion, version)
		if err != nil {
			return err
		}
		version.KumaCp.KumaCpGlobalCompatible = kuma_version.DeploymentVersionCompatible(kuma_version.Build.Version, version.KumaCp.GetVersion())
	}
	return nil
}
