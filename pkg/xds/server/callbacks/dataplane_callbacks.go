package callbacks

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
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
	// OnProxyConnected is executed when proxy is connected after it was disconnected before.
	OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error
	// OnProxyReconnected is executed when proxy is already connected, but there is another stream.
	// This can happen when there is a delay with closing the old connection from the proxy to the control plane.
	OnProxyReconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error
	// OnProxyDisconnected is executed only when the last stream of the proxy disconnects.
	OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, dpKey core_model.ResourceKey)
}

type xdsCallbacks struct {
	callbacks DataplaneCallbacks
	util_xds.NoopCallbacks

	sync.RWMutex
	dpStreams      map[core_xds.StreamID]dpStream
	dpDeltaStreams map[core_xds.StreamID]dpStream
	// we don't need separate map for stream because we use here resource key
	activeStreams map[core_model.ResourceKey]int
}

func DataplaneCallbacksToXdsCallbacks(callbacks DataplaneCallbacks) util_xds.MultiXDSCallbacks {
	return &xdsCallbacks{
		callbacks:      callbacks,
		dpStreams:      map[core_xds.StreamID]dpStream{},
		dpDeltaStreams: map[core_xds.StreamID]dpStream{},
		activeStreams:  map[core_model.ResourceKey]int{},
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
	var lastStreamDpKey *core_model.ResourceKey
	d.Lock()
	dpStream := getDpStream()[streamID]
	if dpKey := dpStream.dp; dpKey != nil {
		d.activeStreams[*dpKey]--
		if d.activeStreams[*dpKey] == 0 {
			lastStreamDpKey = dpKey
			delete(d.activeStreams, *dpKey)
		}
	}
	delete(getDpStream(), streamID)
	d.Unlock()
	if lastStreamDpKey != nil {
		// execute callback after lock is freed, so heavy callback implementation won't block every callback for every DPP.
		d.callbacks.OnProxyDisconnected(dpStream.ctx, streamID, *lastStreamDpKey)
	}
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
	// we don't to increment the counter twice, so checking once again that stream
	// wasn't processed
	alreadyProcessed = getDpStream()[streamID].dp != nil
	if alreadyProcessed {
		return nil
	}

	dpStream := getDpStream()[streamID]
	dpStream.dp = &dpKey
	getDpStream()[streamID] = dpStream

	activeStreams := d.activeStreams[dpKey]
	d.activeStreams[dpKey]++
	d.Unlock()

	if activeStreams == 0 {
		if err := d.callbacks.OnProxyConnected(streamID, dpKey, dpStream.ctx, *metadata); err != nil {
			return err
		}
	} else {
		if err := d.callbacks.OnProxyReconnected(streamID, dpKey, dpStream.ctx, *metadata); err != nil {
			return err
		}
	}
	return nil
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

func (n *NoopDataplaneCallbacks) OnProxyReconnected(core_xds.StreamID, core_model.ResourceKey, context.Context, core_xds.DataplaneMetadata) error {
	return nil
}

func (n *NoopDataplaneCallbacks) OnProxyConnected(core_xds.StreamID, core_model.ResourceKey, context.Context, core_xds.DataplaneMetadata) error {
	return nil
}

func (n *NoopDataplaneCallbacks) OnProxyDisconnected(context.Context, core_xds.StreamID, core_model.ResourceKey) {
}

var _ DataplaneCallbacks = &NoopDataplaneCallbacks{}
