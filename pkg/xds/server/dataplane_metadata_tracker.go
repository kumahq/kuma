package server

import (
	"context"
	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"sync"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	go_cp_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"

	"github.com/kumahq/kuma/pkg/core/xds"
)

type DataplaneMetadataTracker struct {
	mutex             sync.RWMutex
	metadataForStream map[int64]*xds.DataplaneMetadata
	rm                manager.ResourceManager
	environmentType   core.EnvironmentType
}

func NewDataplaneMetadataTracker(rm manager.ResourceManager, environmentType core.EnvironmentType) *DataplaneMetadataTracker {
	return &DataplaneMetadataTracker{
		mutex:             sync.RWMutex{},
		metadataForStream: map[int64]*xds.DataplaneMetadata{},
		rm:                rm,
		environmentType:   environmentType,
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
	md, ok := d.metadataForStream[stream]
	if !ok {
		return
	}
	if err := d.unregisterDataplane(md.DataplaneResource); err != nil {
		xdsServerLog.Error(err, "unable to delete Dataplane resource on stream closing")
	}
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
	md := xds.DataplaneMetadataFromNode(req.Node)
	if _, exist := d.metadataForStream[stream]; !exist {
		if err := d.registerDataplane(md.DataplaneResource); err != nil {
			return err
		}
	}
	d.metadataForStream[stream] = md
	return nil
}

func (d *DataplaneMetadataTracker) OnStreamResponse(int64, *envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}

func (d *DataplaneMetadataTracker) OnFetchRequest(context.Context, *envoy.DiscoveryRequest) error {
	return nil
}

func (d *DataplaneMetadataTracker) OnFetchResponse(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse) {
}

func (d *DataplaneMetadataTracker) registerDataplane(dp *core_mesh.DataplaneResource) error {
	if d.environmentType != core.UniversalEnvironment {
		return nil
	}
	existing := &core_mesh.DataplaneResource{}
	err := d.rm.Get(context.Background(), existing, store.GetBy(model.MetaToResourceKey(dp.GetMeta())))
	if err == nil {
		return errors.Errorf("provided Dataplane %s already exists in %s mesh", dp.GetMeta().GetName(), dp.GetMeta().GetMesh())
	}
	if !store.IsResourceNotFound(err) {
		return err
	}
	return d.rm.Create(context.Background(), dp, store.CreateBy(model.MetaToResourceKey(dp.GetMeta())))
}

func (d *DataplaneMetadataTracker) unregisterDataplane(dp *core_mesh.DataplaneResource) error {
	if d.environmentType != core.UniversalEnvironment {
		return nil
	}
	return d.rm.Delete(context.Background(), &core_mesh.DataplaneResource{}, store.DeleteBy(model.MetaToResourceKey(dp.GetMeta())))
}
