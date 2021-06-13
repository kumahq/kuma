package callbacks

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

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
	proxyTypeForStream map[xds.StreamID]mesh_proto.ProxyType
	sync.RWMutex       // protects createdDpForStream
	shutdownCh         <-chan struct{}
}

var _ util_xds.Callbacks = &DataplaneLifecycle{}

func NewDataplaneLifecycle(resManager manager.ResourceManager, shutdownCh <-chan struct{}) *DataplaneLifecycle {
	return &DataplaneLifecycle{
		resManager:         resManager,
		createdDpForStream: map[xds.StreamID]*model.ResourceKey{},
		proxyTypeForStream: map[xds.StreamID]mesh_proto.ProxyType{},
		shutdownCh:         shutdownCh,
	}
}

func (d *DataplaneLifecycle) OnStreamClosed(streamID int64) {
	// OnStreamClosed method could be called either in case data plane proxy is down or
	// Kuma CP is gracefully shutting down. If Kuma CP is gracefully shutting down we
	// must not delete Dataplane resource, data plane proxy will be reconnected to another
	// instance of Kuma CP.
	select {
	case <-d.shutdownCh:
		lifecycleLog.Info("graceful shutdown, don't delete Dataplane resource")
		return
	default:
	}

	d.Lock()
	defer d.Unlock()
	key := d.createdDpForStream[streamID]
	proxyType := d.proxyTypeForStream[streamID]
	delete(d.createdDpForStream, streamID)
	delete(d.proxyTypeForStream, streamID)

	if key == nil {
		return
	}

	if proxyType == mesh_proto.DataplaneProxyType {
		lifecycleLog.Info("unregistering dataplane", "dataplaneKey", key, "streamID", streamID)
		if err := d.unregisterDataplane(*key); err != nil {
			lifecycleLog.Error(err, "could not unregister dataplane")
		}
		return
	}

	if proxyType == mesh_proto.IngressProxyType {
		lifecycleLog.Info("unregistering zone ingress", "zoneIngressKey", key, "streamID", streamID)
		if err := d.unregisterZoneIngress(*key); err != nil {
			lifecycleLog.Error(err, "could not unregister zone ingress")
		}
		return
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

	if md.GetProxyType() == mesh_proto.DataplaneProxyType && md.GetDataplaneResource() != nil {
		lifecycleLog.Info("registering dataplane", "dataplane", md.GetDataplaneResource(), "streamID", streamID, "nodeID", request.NodeId())
		if err := d.registerDataplane(md.GetDataplaneResource()); err != nil {
			return errors.Wrap(err, "could not register dataplane passed in kuma-dp run")
		}
		key := model.MetaToResourceKey(md.GetDataplaneResource().GetMeta())
		d.createdDpForStream[streamID] = &key
		d.proxyTypeForStream[streamID] = mesh_proto.DataplaneProxyType
		return nil
	}

	if md.GetProxyType() == mesh_proto.IngressProxyType && md.ZoneIngressResource != nil {
		lifecycleLog.Info("registering zone ingress", "zoneIngress", md.ZoneIngressResource, "streamID", streamID, "nodeID", request.NodeId())
		if err := d.registerZoneIngress(md.ZoneIngressResource); err != nil {
			return errors.Wrap(err, "could not register zone ingress passed in kuma-dp run")
		}
		key := model.MetaToResourceKey(md.ZoneIngressResource.GetMeta())
		d.createdDpForStream[streamID] = &key
		d.proxyTypeForStream[streamID] = mesh_proto.IngressProxyType
		return nil
	}

	d.createdDpForStream[streamID] = nil // put nil so we don't have to read metadata every time
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

func (d *DataplaneLifecycle) registerZoneIngress(zi *core_mesh.ZoneIngressResource) error {
	key := model.MetaToResourceKey(zi.GetMeta())
	existing := core_mesh.NewZoneIngressResource()
	return manager.Upsert(d.resManager, key, existing, func(resource model.Resource) {
		_ = existing.SetSpec(zi.GetSpec()) // ignore error because the spec type is the same
	})
}

func (d *DataplaneLifecycle) unregisterDataplane(key model.ResourceKey) error {
	return d.resManager.Delete(context.Background(), core_mesh.NewDataplaneResource(), store.DeleteBy(key))
}

func (d *DataplaneLifecycle) unregisterZoneIngress(key model.ResourceKey) error {
	return d.resManager.Delete(context.Background(), core_mesh.NewZoneIngressResource(), store.DeleteBy(key))
}
