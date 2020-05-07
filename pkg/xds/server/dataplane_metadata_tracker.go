package server

import (
	"context"
	"sync"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	go_cp_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"

	"github.com/Kong/kuma/pkg/core/xds"
)

type DataplaneMetadataTracker struct {
	mutex             sync.RWMutex
	metadataForStream map[int64]*xds.DataplaneMetadata
}

func NewDataplaneMetadataTracker() *DataplaneMetadataTracker {
	return &DataplaneMetadataTracker{
		mutex:             sync.RWMutex{},
		metadataForStream: map[int64]*xds.DataplaneMetadata{},
	}
}

func (d *DataplaneMetadataTracker) Metadata(streamId int64) *xds.DataplaneMetadata {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	metadata, found := d.metadataForStream[streamId]
	if found {
		return metadata
	} else {
		return &xds.DataplaneMetadata{}
	}
}

var _ go_cp_server.Callbacks = &DataplaneMetadataTracker{}

func (d *DataplaneMetadataTracker) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (d *DataplaneMetadataTracker) OnStreamClosed(stream int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.metadataForStream, stream)
}

func (d *DataplaneMetadataTracker) OnStreamRequest(stream int64, req *envoy.DiscoveryRequest) error {
	if req.Node == nil {
		// from https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#ack-nack-and-versioning:
		// Only the first request on a stream is guaranteed to carry the node identifier.
		// The subsequent discovery requests on the same stream may carry an empty node identifier.
		// This holds true regardless of the acceptance of the discovery responses on the same stream.
		// The node identifier should always be identical if present more than once on the stream.
		// It is sufficient to only check the first message for the node identifier as a result.
		return nil
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.metadataForStream[stream] = xds.DataplaneMetadataFromNode(req.Node)
	return nil
}

func (d *DataplaneMetadataTracker) OnStreamResponse(int64, *envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}

func (d *DataplaneMetadataTracker) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

func (d *DataplaneMetadataTracker) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}
