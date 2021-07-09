package callbacks

import (
	"context"
	"sync"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type DataplaneMetadataTracker struct {
	NoopDataplaneCallbacks
	sync.RWMutex
	metadataForDp map[core_model.ResourceKey]*core_xds.DataplaneMetadata
}

var _ DataplaneCallbacks = &DataplaneMetadataTracker{}

func NewDataplaneMetadataTracker() *DataplaneMetadataTracker {
	return &DataplaneMetadataTracker{
		metadataForDp: map[core_model.ResourceKey]*core_xds.DataplaneMetadata{},
	}
}

func (d *DataplaneMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	d.RLock()
	defer d.RUnlock()
	return d.metadataForDp[dpKey]
}

func (d *DataplaneMetadataTracker) OnStreamConnected(_ core_xds.StreamID, dpKey core_model.ResourceKey, _ context.Context, metadata core_xds.DataplaneMetadata) error {
	// We use OnStreamConnected, not OnFirstStreamConnected because if there are many xDS streams, we want to follow metadata from the newest stream.
	d.Lock()
	defer d.Unlock()
	d.metadataForDp[dpKey] = &metadata
	return nil
}

func (d *DataplaneMetadataTracker) OnLastStreamDisconnected(_ core_xds.StreamID, dpKey core_model.ResourceKey) {
	d.Lock()
	defer d.Unlock()
	delete(d.metadataForDp, dpKey)
}
