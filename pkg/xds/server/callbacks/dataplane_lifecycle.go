package callbacks

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

var lifecycleLog = core.Log.WithName("xds").WithName("dp-lifecycle")

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
	util_xds.NoopCallbacks
	resManager manager.ResourceManager
	// createdDpForStream stores map from StreamID to created ResourceKey of Dataplane.
	// we store nil values for streams without Dataplane in metadata to avoid accessing metadata with every DiscoveryRequest
	createdDpForStream map[xds.StreamID]*model.ResourceKey
	sync.RWMutex       // protects createdDpForStream
}

var _ util_xds.Callbacks = &DataplaneLifecycle{}

func NewDataplaneLifecycle(resManager manager.ResourceManager) *DataplaneLifecycle {
	return &DataplaneLifecycle{
		resManager:         resManager,
		createdDpForStream: map[xds.StreamID]*model.ResourceKey{},
	}
}

func (d *DataplaneLifecycle) OnStreamClosed(streamID int64) {
	d.Lock()
	defer d.Unlock()
	key := d.createdDpForStream[streamID]
	delete(d.createdDpForStream, streamID)
	if key != nil {
		lifecycleLog.Info("unregistering dataplane", "dataplaneKey", key, "streamID", streamID)
		if err := d.unregisterDataplane(*key); err != nil {
			lifecycleLog.Error(err, "could not unregister dataplane")
		}
	}
}

func (d *DataplaneLifecycle) OnStreamRequest(streamID int64, request util_xds.DiscoveryRequest) error {
	if request.NodeId() == "" { // Only the first request on a stream is guaranteed to carry the node identifier.
		return nil
	}

	if d.streamProcessed(streamID) {
		return nil
	}

	d.Lock()
	defer d.Unlock()
	md := xds.DataplaneMetadataFromXdsMetadata(request.Metadata())
	if md.DataplaneResource != nil {
		lifecycleLog.Info("registering dataplane", "dataplane", md.DataplaneResource, "streamID", streamID, "nodeID", request.NodeId())
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

func (d *DataplaneLifecycle) streamProcessed(streamID int64) bool {
	d.RLock()
	defer d.RUnlock()
	_, ok := d.createdDpForStream[streamID]
	return ok
}

func (d *DataplaneLifecycle) registerDataplane(dp *core_mesh.DataplaneResource) error {
	key := model.MetaToResourceKey(dp.GetMeta())
	existing := core_mesh.NewDataplaneResource()
	return manager.Upsert(d.resManager, key, existing, func(resource model.Resource) {
		_ = existing.SetSpec(dp.GetSpec()) // ignore error because the spec type is the same
	})
}

func (d *DataplaneLifecycle) unregisterDataplane(key model.ResourceKey) error {
	return d.resManager.Delete(context.Background(), core_mesh.NewDataplaneResource(), store.DeleteBy(key))
}
