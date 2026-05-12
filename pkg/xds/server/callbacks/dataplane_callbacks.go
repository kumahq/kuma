package callbacks

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/v2/pkg/util/xds"
	xds_metrics "github.com/kumahq/kuma/v2/pkg/xds/metrics"
)

// DataplaneCallbacks are XDS callbacks that keep the context of Kuma Dataplane.
// In the ideal world we could assume that one Dataplane has one xDS stream.
// Due to race network latencies etc. there might be a situation when one Dataplane has many xDS streams for the short period of time.
// Those callbacks helps us to deal with such situation.
//
// Keep in mind that it does not solve many xDS streams across many instances of the Control Plane.
// If there are many instances of the Control Plane and Dataplane reconnects, there might be an old stream
// in one instance of CP and a new stream in a new instance of CP.
type DataplaneCallbacks interface {
	// OnProxyConnected is executed when an active stream from a proxy is connected
	OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error
	// OnProxyDisconnected is executed only when an active stream of the proxy disconnects
	OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, dpKey core_model.ResourceKey)
}

type xdsCallbacks struct {
	callbacks DataplaneCallbacks
	metrics   *xds_metrics.Metrics
	util_xds.NoopCallbacks

	sync.RWMutex
	dpStreams      map[core_xds.StreamID]dpStream
	dpDeltaStreams map[core_xds.StreamID]dpStream
	// we don't need separate map for stream because we use here resource key
	activeStreams map[core_model.ResourceKey]core_xds.StreamID
	// connectingStream tracks streams that are currently in the process of registering
	// (i.e. OnProxyConnected has been called but has not yet returned). This prevents
	// a concurrent stream for the same dpKey from racing past the activeStreams guard
	// before registration completes, which would cause OnProxyConnected to be invoked
	// twice for the same dataplane.
	connectingStream map[core_model.ResourceKey]core_xds.StreamID
}

func DataplaneCallbacksToXdsCallbacks(callbacks DataplaneCallbacks, metrics *xds_metrics.Metrics) util_xds.MultiXDSCallbacks {
	return &xdsCallbacks{
		callbacks:        callbacks,
		metrics:          metrics,
		dpStreams:        map[core_xds.StreamID]dpStream{},
		dpDeltaStreams:   map[core_xds.StreamID]dpStream{},
		activeStreams:    map[core_model.ResourceKey]core_xds.StreamID{},
		connectingStream: map[core_model.ResourceKey]core_xds.StreamID{},
	}
}

type dpStream struct {
	dp  *core_model.ResourceKey
	ctx context.Context
}

func (d *xdsCallbacks) getDpStream() map[core_xds.StreamID]dpStream {
	return d.dpStreams
}

func (d *xdsCallbacks) getDpDeltaStream() map[core_xds.StreamID]dpStream {
	return d.dpDeltaStreams
}

var _ util_xds.MultiXDSCallbacks = &xdsCallbacks{}

func (d *xdsCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
	d.onStreamClosed(streamID, d.getDpStream)
}

func (d *xdsCallbacks) OnDeltaStreamClosed(streamID core_xds.StreamID) {
	d.onStreamClosed(streamID, d.getDpDeltaStream)
}

func (d *xdsCallbacks) OnStreamRequest(streamID core_xds.StreamID, request util_xds.DiscoveryRequest) error {
	return d.onStreamRequest(streamID, request, d.getDpStream)
}

func (d *xdsCallbacks) OnStreamDeltaRequest(streamID core_xds.StreamID, request util_xds.DeltaDiscoveryRequest) error {
	return d.onStreamRequest(streamID, request, d.getDpDeltaStream)
}

func (d *xdsCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	return d.onStreamOpen(ctx, streamID, d.getDpStream)
}

func (d *xdsCallbacks) OnDeltaStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	return d.onStreamOpen(ctx, streamID, d.getDpDeltaStream)
}

func (d *xdsCallbacks) onStreamClosed(streamID core_xds.StreamID, getDpStream func() map[core_xds.StreamID]dpStream) {
	var streamDpKey *core_model.ResourceKey
	d.RLock()
	dpStream := getDpStream()[streamID]
	streamDpKey = dpStream.dp
	d.RUnlock()

	if streamDpKey != nil {
		// execute callback after lock is freed, so heavy callback implementation won't block every callback for every DPP.
		d.callbacks.OnProxyDisconnected(dpStream.ctx, streamID, *streamDpKey)
	}

	d.Lock()
	if streamDpKey != nil {
		// Guard by stream ID: a stale-owner takeover may have already replaced
		// this stream as the active owner. Only remove the entry if we are still
		// the current owner, so the new owner's registration is not clobbered.
		if d.activeStreams[*streamDpKey] == streamID {
			delete(d.activeStreams, *streamDpKey)
		}
		// Safety cleanup: remove any in-progress connecting entry for this stream.
		if d.connectingStream[*streamDpKey] == streamID {
			delete(d.connectingStream, *streamDpKey)
		}
	}
	delete(getDpStream(), streamID)
	d.Unlock()
}

func (d *xdsCallbacks) onStreamRequest(streamID core_xds.StreamID, request util_xds.Request, getDpStream func() map[core_xds.StreamID]dpStream) error {
	if request.NodeId() == "" {
		// from https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#ack-nack-and-versioning:
		// Only the first request on a stream is guaranteed to carry the node identifier.
		// The subsequent discovery requests on the same stream may carry an empty node identifier.
		// This holds true regardless of the acceptance of the discovery responses on the same stream.
		// The node identifier should always be identical if present more than once on the stream.
		// It is sufficient to only check the first message for the node identifier as a result.
		return nil
	}

	d.RLock()
	alreadyProcessed := getDpStream()[streamID].dp != nil
	d.RUnlock()
	if alreadyProcessed {
		return nil
	}

	proxyId, err := core_xds.ParseProxyIdFromString(request.NodeId())
	if err != nil {
		return errors.Wrap(err, "invalid node ID")
	}
	dpKey := proxyId.ToResourceKey()
	metadata := core_xds.DataplaneMetadataFromXdsMetadata(request.Metadata())
	if metadata == nil {
		return errors.New("metadata in xDS Node cannot be nil")
	}

	d.Lock()
	// in case client will open 2 concurrent request for the same streamID then
	// we don't want to increment the counter twice, so checking once again that stream
	// wasn't processed
	alreadyProcessed = getDpStream()[streamID].dp != nil
	if alreadyProcessed {
		d.Unlock()
		return nil
	}

	// If another stream is already registering for this dpKey, return RESOURCE_EXHAUSTED
	// so Envoy applies backoff before retrying, rather than hammering the CP.
	if connectingStreamID, registering := d.connectingStream[dpKey]; registering && connectingStreamID != streamID {
		d.incrementInProgressRetries(dpKey.Mesh, string(metadata.GetProxyType()))
		d.Unlock()
		return status.Error(codes.ResourceExhausted, "registration in progress for this node, try again later")
	}

	var streamInfo dpStream
	ownerStreamID, alreadyConnected := d.activeStreams[dpKey]
	if alreadyConnected {
		// Check whether the existing owner's stream context is already cancelled.
		// Under high dataplane churn a narrow window exists between gRPC GOAWAY/FIN
		// (which cancels the stream context) and the CP processing OnStreamClosed
		// (which removes the activeStreams entry). If we hit that window a new stream
		// would otherwise be rejected even though the old one is truly gone.
		ownerCtx := getDpStream()[ownerStreamID].ctx
		if ownerCtx == nil || ownerCtx.Err() == nil {
			// Owner is still alive — reject the incoming stream to avoid races.
			d.Unlock()
			// we don't allow more than one active stream from a data plane as there can be race conditions
			return errors.New("there is already an active stream from this node, try again later")
		}
		// Owner's context is cancelled — it is stale. Allow the new stream to take
		// over. The stale owner's OnStreamClosed will still fire and call
		// OnProxyDisconnected only if stream state still exists. Evict stale stream
		// state here so OnStreamClosed for the stale owner does not disconnect the
		// replacement stream's watchdog lifecycle.
		delete(getDpStream(), ownerStreamID)
		d.incrementStaleTakeovers(dpKey.Mesh, string(metadata.GetProxyType()))
	}

	streamInfo = getDpStream()[streamID]
	streamInfo.dp = &dpKey
	getDpStream()[streamID] = streamInfo
	d.connectingStream[dpKey] = streamID
	d.Unlock()

	err = d.callbacks.OnProxyConnected(streamID, dpKey, streamInfo.ctx, *metadata)

	d.Lock()
	if d.connectingStream[dpKey] == streamID {
		delete(d.connectingStream, dpKey)
	}
	if err == nil {
		d.activeStreams[dpKey] = streamID
	}
	d.Unlock()

	return err
}

func (d *xdsCallbacks) onStreamOpen(ctx context.Context, streamID core_xds.StreamID, getDpStream func() map[core_xds.StreamID]dpStream) error {
	d.Lock()
	defer d.Unlock()
	dps := dpStream{
		ctx: ctx,
	}
	getDpStream()[streamID] = dps
	return nil
}

// NoopDataplaneCallbacks are empty callbacks that helps to implement DataplaneCallbacks without need to implement every function.
type NoopDataplaneCallbacks struct{}

func (n *NoopDataplaneCallbacks) OnProxyConnected(core_xds.StreamID, core_model.ResourceKey, context.Context, core_xds.DataplaneMetadata) error {
	return nil
}

func (n *NoopDataplaneCallbacks) OnProxyDisconnected(_ context.Context, _ core_xds.StreamID, _ core_model.ResourceKey) {
}

var _ DataplaneCallbacks = &NoopDataplaneCallbacks{}

func (d *xdsCallbacks) incrementStaleTakeovers(mesh, proxyType string) {
	if d.metrics == nil {
		return
	}
	d.metrics.XdsStreamStaleOwnerTakeovers.WithLabelValues(mesh, proxyType).Inc()
}

func (d *xdsCallbacks) incrementInProgressRetries(mesh, proxyType string) {
	if d.metrics == nil {
		return
	}
	d.metrics.XdsStreamRegistrationInProgressRetries.WithLabelValues(mesh, proxyType).Inc()
}
