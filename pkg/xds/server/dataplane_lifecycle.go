package server

import (
	"context"
	"sync"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	go_cp_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
)

// DataplaneLifecycle is responsible for creating a deleting dataplanes that are passed through metadata
// There are two possible workflows
// 1) apply Dataplane resource before kuma-dp run and run kuma-dp
// 2) run kuma-dp and pass Dataplane resource as an argument to kuma-dp
// This component support second use case. When user passes Dataplane to kuma-dp it is attached to bootstrap request.
// Then, bootstrap server generates bootstrap configuration with Dataplane embedded in Envoy metadata.
// Here, we read Dataplane resource from metadata and a create resource on first DiscoveryRequest and remove on StreamClosed.
//
// This flow is optional, you may still want to go with 1. an example of this is Kubernetes deployment.
type DataplaneLifecycle struct {
	resManager manager.ResourceManager
	// createdDpForStream stores map from StreamID to created ResourceKey of Dataplane.
	// we store nil values for streams without Dataplane in metadata to avoid accessing metadata with every DiscoveryRequest
	createdDpForStream map[xds.StreamID]*model.ResourceKey
	sync.RWMutex       // protects createdDpForStream
}

func NewDataplaneLifecycle(resManager manager.ResourceManager) *DataplaneLifecycle {
	return &DataplaneLifecycle{
		resManager:         resManager,
		createdDpForStream: map[xds.StreamID]*model.ResourceKey{},
	}
}

func (d *DataplaneLifecycle) OnStreamOpen(_ context.Context, _ int64, _ string) error {
	return nil
}

func (d *DataplaneLifecycle) OnStreamClosed(streamID int64) {
	d.Lock()
	defer d.Unlock()
	key := d.createdDpForStream[streamID]
	delete(d.createdDpForStream, streamID)
	if key != nil {
		xdsServerLog.Info("unregistering dataplane", "dataplaneKey", key, "streamID", streamID)
		if err := d.unregisterDataplane(*key); err != nil {
			xdsServerLog.Error(err, "could not unregister dataplane")
		}
	}
}

func (d *DataplaneLifecycle) OnStreamRequest(streamID int64, request *envoy_api_v2.DiscoveryRequest) error {
	if request.Node == nil { // Only the first request on a stream is guaranteed to carry the node identifier.
		return nil
	}

	d.RLock()
	if _, ok := d.createdDpForStream[streamID]; ok {
		d.RUnlock()
		return nil
	}
	d.RUnlock()

	d.Lock()
	defer d.Unlock()
	md := xds.DataplaneMetadataFromNode(request.Node)
	if md.DataplaneResource != nil {
		xdsServerLog.Info("registering dataplane", "dataplane", md.DataplaneResource, "streamID", streamID, "nodeID", request.Node.Id)
		if err := d.registerDataplane(md.DataplaneResource); err != nil {
			return errors.Wrap(err, "could not register dataplane passed in kuma-dp run")
		}
		key := model.MetaToResourceKey(md.DataplaneResource.GetMeta())
		d.createdDpForStream[streamID] = &key
	} else {
		d.createdDpForStream[streamID] = nil // put nil so we don't have to read metadata every time
	}
	return nil
}

func (d *DataplaneLifecycle) registerDataplane(dp *core_mesh.DataplaneResource) error {
	key := model.MetaToResourceKey(dp.GetMeta())
	existing := &core_mesh.DataplaneResource{}
	return manager.Upsert(d.resManager, key, existing, func(resource model.Resource) {
		_ = existing.SetSpec(dp.GetSpec()) // ignore error because the spec type is the same
	})
}

func (d *DataplaneLifecycle) unregisterDataplane(key model.ResourceKey) error {
	return d.resManager.Delete(context.Background(), &core_mesh.DataplaneResource{}, store.DeleteBy(key))
}

func (d *DataplaneLifecycle) OnStreamResponse(_ int64, _ *envoy_api_v2.DiscoveryRequest, _ *envoy_api_v2.DiscoveryResponse) {
}

func (d *DataplaneLifecycle) OnFetchRequest(_ context.Context, _ *envoy_api_v2.DiscoveryRequest) error {
	return nil
}

func (d *DataplaneLifecycle) OnFetchResponse(request *envoy_api_v2.DiscoveryRequest, response *envoy_api_v2.DiscoveryResponse) {
}

var _ go_cp_server.Callbacks = &DataplaneLifecycle{}
