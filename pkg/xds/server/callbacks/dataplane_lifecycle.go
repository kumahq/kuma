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
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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
	resManager manager.ResourceManager

	sync.RWMutex         // protects createdDpByCallbacks
	createdDpByCallbacks map[model.ResourceKey]mesh_proto.ProxyType
	appCtx               context.Context
}

func (d *DataplaneLifecycle) OnProxyConnected(streamID core_xds.StreamID, dpKey model.ResourceKey, _ context.Context, md core_xds.DataplaneMetadata) error {
	return d.register(streamID, dpKey, md)
}

func (d *DataplaneLifecycle) OnProxyReconnected(streamID core_xds.StreamID, dpKey model.ResourceKey, _ context.Context, md core_xds.DataplaneMetadata) error {
	return d.register(streamID, dpKey, md)
}

func (d *DataplaneLifecycle) register(streamID core_xds.StreamID, dpKey model.ResourceKey, md core_xds.DataplaneMetadata) error {
	switch {
	case md.GetProxyType() == mesh_proto.DataplaneProxyType && md.GetDataplaneResource() != nil:
		dp := md.GetDataplaneResource()
		lifecycleLog.Info("registering dataplane", "dataplane", dp, "dataplaneKey", dpKey, "streamID", streamID)
		if err := d.registerDataplane(dp); err != nil {
			return errors.Wrap(err, "could not register dataplane passed in kuma-dp run")
		}
	case md.GetProxyType() == mesh_proto.IngressProxyType && md.GetZoneIngressResource() != nil:
		zi := md.GetZoneIngressResource()
		lifecycleLog.Info("registering zone ingress", "zoneIngress", zi, "zoneIngressKey", dpKey, "streamID", streamID)
		if err := d.registerZoneIngress(zi); err != nil {
			return errors.Wrap(err, "could not register zone ingress passed in kuma-dp run")
		}
	default:
		return nil
	}

	d.Lock()
	d.createdDpByCallbacks[dpKey] = md.GetProxyType()
	d.Unlock()

	return nil
}

func (d *DataplaneLifecycle) OnProxyDisconnected(streamID core_xds.StreamID, dpKey model.ResourceKey) {
	// OnStreamClosed method could be called either in case data plane proxy is down or
	// Kuma CP is gracefully shutting down. If Kuma CP is gracefully shutting down we
	// must not delete Dataplane resource, data plane proxy will be reconnected to another
	// instance of Kuma CP.
	select {
	case <-d.appCtx.Done():
		lifecycleLog.Info("graceful shutdown, don't delete Dataplane resource")
		return
	default:
	}

	d.Lock()
	proxyType, createdByCallbacks := d.createdDpByCallbacks[dpKey]
	if createdByCallbacks {
		delete(d.createdDpByCallbacks, dpKey)
	}
	d.Unlock()

	if !createdByCallbacks {
		return
	}

	switch proxyType {
	case mesh_proto.DataplaneProxyType:
		lifecycleLog.Info("unregistering dataplane", "dataplaneKey", dpKey, "streamID", streamID)
		if err := d.unregisterDataplane(dpKey); err != nil {
			lifecycleLog.Error(err, "could not unregister dataplane")
		}
	case mesh_proto.IngressProxyType:
		lifecycleLog.Info("unregistering zone ingress", "zoneIngressKey", dpKey, "streamID", streamID)
		if err := d.unregisterZoneIngress(dpKey); err != nil {
			lifecycleLog.Error(err, "could not unregister zone ingress")
		}
	}
}

var _ DataplaneCallbacks = &DataplaneLifecycle{}

func NewDataplaneLifecycle(appCtx context.Context, resManager manager.ResourceManager) *DataplaneLifecycle {
	return &DataplaneLifecycle{
		resManager:           resManager,
		createdDpByCallbacks: map[model.ResourceKey]mesh_proto.ProxyType{},
		appCtx:               appCtx,
	}
}

func (d *DataplaneLifecycle) registerDataplane(dp *core_mesh.DataplaneResource) error {
	key := model.MetaToResourceKey(dp.GetMeta())
	existing := core_mesh.NewDataplaneResource()
	return manager.Upsert(d.resManager, key, existing, func(resource model.Resource) error {
		return existing.SetSpec(dp.GetSpec())
	})
}

func (d *DataplaneLifecycle) registerZoneIngress(zi *core_mesh.ZoneIngressResource) error {
	key := model.MetaToResourceKey(zi.GetMeta())
	existing := core_mesh.NewZoneIngressResource()
	return manager.Upsert(d.resManager, key, existing, func(resource model.Resource) error {
		return existing.SetSpec(zi.GetSpec())
	})
}

func (d *DataplaneLifecycle) unregisterDataplane(key model.ResourceKey) error {
	return d.resManager.Delete(context.Background(), core_mesh.NewDataplaneResource(), store.DeleteBy(key))
}

func (d *DataplaneLifecycle) unregisterZoneIngress(key model.ResourceKey) error {
	return d.resManager.Delete(context.Background(), core_mesh.NewZoneIngressResource(), store.DeleteBy(key))
}
