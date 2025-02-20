package callbacks

import (
	"context"
	"sync"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type DataplaneMetadataTracker struct {
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

func (d *DataplaneMetadataTracker) OnProxyConnected(_ core_xds.StreamID, dpKey core_model.ResourceKey, _ context.Context, metadata core_xds.DataplaneMetadata) error {
	d.storeMetadata(dpKey, metadata)
	return nil
}

func (d *DataplaneMetadataTracker) storeMetadata(dpKey core_model.ResourceKey, metadata core_xds.DataplaneMetadata) {
	d.Lock()
	defer d.Unlock()
	d.metadataForDp[dpKey] = &metadata
}

func (d *DataplaneMetadataTracker) OnProxyDisconnected(_ context.Context, _ core_xds.StreamID, dpKey core_model.ResourceKey, done chan<- struct{}) {
	d.Lock()
	defer d.Unlock()
	delete(d.metadataForDp, dpKey)
	done <- struct{}{}
}
