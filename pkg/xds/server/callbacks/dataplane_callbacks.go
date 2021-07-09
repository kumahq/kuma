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
//
// Those callbacks may be also used with SDS. In case of SDS, at this moment Envoy creates many SDS streams to the Control Plane.
type DataplaneCallbacks interface {
	// OnStreamConnected is executed when Dataplane is connected to the control plane (go-control-plane's OnStreamOpen)
	// and Dataplane executes the first request (go-control-plane's OnStreamRequest, only in this phase we can extract Dataplane and DataplaneMetadata).
	// If there are many xDS stream this callback is executed for every new xDS stream.
	OnStreamConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error
	// OnFirstStreamConnected is similar to OnStreamConnected but if there are many xDS streams, it is only executed for the first stream.
	OnFirstStreamConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error

	// OnStreamDisconnected is executed when Dataplane stream disconnects.
	OnStreamDisconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey)
	// OnLastStreamDisconnected is executed only when the last stream of Dataplane disconnects.
	OnLastStreamDisconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey)
}

type dpCallbacks struct {
	callbacks DataplaneCallbacks
	util_xds.NoopCallbacks

	sync.RWMutex
	dpStreams     map[core_xds.StreamID]dpStream
	activeStreams map[core_model.ResourceKey]int
}

func DataplaneCallbacksToXdsCallbacks(callbacks DataplaneCallbacks) util_xds.Callbacks {
	return &dpCallbacks{
		callbacks:     callbacks,
		dpStreams:     map[core_xds.StreamID]dpStream{},
		activeStreams: map[core_model.ResourceKey]int{},
	}
}

type dpStream struct {
	dp  *core_model.ResourceKey
	ctx context.Context
}

var _ util_xds.Callbacks = &dpCallbacks{}

func (d *dpCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
	d.Lock()
	defer d.Unlock()
	dpStream := d.dpStreams[streamID]
	if dpKey := dpStream.dp; dpKey != nil {
		d.activeStreams[*dpKey]--
		d.callbacks.OnStreamDisconnected(streamID, *dpKey)
		if d.activeStreams[*dpKey] == 0 {
			d.callbacks.OnLastStreamDisconnected(streamID, *dpKey)
			delete(d.activeStreams, *dpKey)
		}
	}
	delete(d.dpStreams, streamID)
}

func (d *dpCallbacks) OnStreamRequest(streamID core_xds.StreamID, request util_xds.DiscoveryRequest) error {
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
	alreadyProcessed := d.dpStreams[streamID].dp != nil
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
	dpStream := d.dpStreams[streamID]
	dpStream.dp = &dpKey
	d.dpStreams[streamID] = dpStream

	activeStreams := d.activeStreams[dpKey]
	d.activeStreams[dpKey]++
	d.Unlock()

	if err := d.callbacks.OnStreamConnected(streamID, dpKey, dpStream.ctx, *metadata); err != nil {
		return err
	}

	if activeStreams == 0 {
		if err := d.callbacks.OnFirstStreamConnected(streamID, dpKey, dpStream.ctx, *metadata); err != nil {
			return err
		}
	}
	return nil
}

func (d *dpCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	d.Lock()
	defer d.Unlock()
	dps := dpStream{
		ctx: ctx,
	}
	d.dpStreams[streamID] = dps
	return nil
}

// NoopDataplaneCallbacks are empty callbacks that helps to implement DataplaneCallbacks without need to implement every function.
type NoopDataplaneCallbacks struct {
}

func (n *NoopDataplaneCallbacks) OnStreamConnected(core_xds.StreamID, core_model.ResourceKey, context.Context, core_xds.DataplaneMetadata) error {
	return nil
}

func (n *NoopDataplaneCallbacks) OnFirstStreamConnected(core_xds.StreamID, core_model.ResourceKey, context.Context, core_xds.DataplaneMetadata) error {
	return nil
}

func (n *NoopDataplaneCallbacks) OnStreamDisconnected(core_xds.StreamID, core_model.ResourceKey) {
}

func (n *NoopDataplaneCallbacks) OnLastStreamDisconnected(core_xds.StreamID, core_model.ResourceKey) {
}

var _ DataplaneCallbacks = &NoopDataplaneCallbacks{}
