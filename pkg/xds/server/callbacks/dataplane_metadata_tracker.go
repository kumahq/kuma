package callbacks

import (
	"sync"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type DataplaneMetadataTracker struct {
	util_xds.NoopCallbacks
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

var _ util_xds.Callbacks = &DataplaneMetadataTracker{}

func (d *DataplaneMetadataTracker) OnStreamClosed(stream int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.metadataForStream, stream)
}

func (d *DataplaneMetadataTracker) OnStreamRequest(stream int64, req util_xds.DiscoveryRequest) error {
	if req.NodeId() == "" {
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
	d.metadataForStream[stream] = xds.DataplaneMetadataFromXdsMetadata(req.Metadata())
	return nil
}
