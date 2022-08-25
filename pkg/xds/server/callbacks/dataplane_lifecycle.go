package callbacks

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
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
	resManager    manager.ResourceManager
	authenticator xds_auth.Authenticator

	sync.RWMutex         // protects createdDpByCallbacks
	createdDpByCallbacks map[model.ResourceKey]mesh_proto.ProxyType
	appCtx               context.Context
	deregistrationDelay  time.Duration
	cpInstanceID         string
}

var _ DataplaneCallbacks = &DataplaneLifecycle{}

func NewDataplaneLifecycle(
	appCtx context.Context,
	resManager manager.ResourceManager,
	authenticator xds_auth.Authenticator,
	deregistrationDelay time.Duration,
	cpInstanceID string,
) *DataplaneLifecycle {
	return &DataplaneLifecycle{
		resManager:           resManager,
		authenticator:        authenticator,
		createdDpByCallbacks: map[model.ResourceKey]mesh_proto.ProxyType{},
		appCtx:               appCtx,
		deregistrationDelay:  deregistrationDelay,
		cpInstanceID:         cpInstanceID,
	}
}

func (d *DataplaneLifecycle) OnProxyConnected(streamID core_xds.StreamID, dpKey model.ResourceKey, ctx context.Context, md core_xds.DataplaneMetadata) error {
	return d.register(ctx, streamID, dpKey, md)
}

func (d *DataplaneLifecycle) OnProxyReconnected(streamID core_xds.StreamID, dpKey model.ResourceKey, ctx context.Context, md core_xds.DataplaneMetadata) error {
	return d.register(ctx, streamID, dpKey, md)
}

func (d *DataplaneLifecycle) register(ctx context.Context, streamID core_xds.StreamID, dpKey model.ResourceKey, md core_xds.DataplaneMetadata) error {
	switch {
	case md.GetProxyType() == mesh_proto.DataplaneProxyType && md.GetDataplaneResource() != nil:
		dp := md.GetDataplaneResource()
		log := lifecycleLog.WithValues("dataplane", dp, "dataplaneKey", dpKey, "streamID", streamID)
		log.Info("registering dataplane")
		if err := d.registerDataplane(ctx, dp); err != nil {
			log.Info("cannot register dataplane", "reason", err.Error())
			return errors.Wrap(err, "could not register dataplane passed in kuma-dp run")
		}
	case md.GetProxyType() == mesh_proto.IngressProxyType && md.GetZoneIngressResource() != nil:
		zi := md.GetZoneIngressResource()
		log := lifecycleLog.WithValues("zoneIngress", zi, "zoneIngressKey", dpKey, "streamID", streamID)
		log.Info("registering zone ingress")
		if err := d.registerZoneIngress(ctx, zi); err != nil {
			log.Info("cannot register zone ingress", "reason", err.Error())
			return errors.Wrap(err, "could not register zone ingress passed in kuma-dp run")
		}
	case md.GetProxyType() == mesh_proto.EgressProxyType && md.GetZoneEgressResource() != nil:
		zi := md.GetZoneEgressResource()
		log := lifecycleLog.WithValues("zoneEgress", zi, "zoneEgressKey", dpKey, "streamID", streamID)
		log.Info("registering zone egress")
		if err := d.registerZoneEgress(ctx, zi); err != nil {
			log.Info("cannot register zone egress", "reason", err.Error())
			return errors.Wrap(err, "could not register zone egress passed in kuma-dp run")
		}
	default:
		return nil
	}

	d.Lock()
	d.createdDpByCallbacks[dpKey] = md.GetProxyType()
	d.Unlock()

	return nil
}

func (d *DataplaneLifecycle) OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, dpKey model.ResourceKey) {
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

	log := lifecycleLog.WithValues("proxyType", proxyType, "proxyKey", dpKey, "streamID", streamID)
	switch proxyType {
	case mesh_proto.DataplaneProxyType:
		err := d.deregisterProxy(ctx, dpKey, core_mesh.NewDataplaneResource(), core_mesh.NewDataplaneInsightResource(), log)
		if err != nil {
			lifecycleLog.Error(err, "could not unregister dataplane")
		}
	case mesh_proto.IngressProxyType:
		err := d.deregisterProxy(ctx, dpKey, core_mesh.NewZoneIngressResource(), core_mesh.NewZoneIngressInsightResource(), log)
		if err != nil {
			lifecycleLog.Error(err, "could not unregister zone ingress")
		}
	case mesh_proto.EgressProxyType:
		err := d.deregisterProxy(ctx, dpKey, core_mesh.NewZoneEgressResource(), core_mesh.NewZoneEgressInsightResource(), log)
		if err != nil {
			lifecycleLog.Error(err, "could not unregister zone egress")
		}
	}
}

func (d *DataplaneLifecycle) registerDataplane(ctx context.Context, dp *core_mesh.DataplaneResource) error {
	key := model.MetaToResourceKey(dp.GetMeta())
	existing := core_mesh.NewDataplaneResource()
	return manager.Upsert(ctx, d.resManager, key, existing, func(resource model.Resource) error {
		if err := d.validateUpsert(ctx, existing); err != nil {
			return errors.Wrap(err, "you are trying to override existing dataplane to which you don't have an access.")
		}
		return existing.SetSpec(dp.GetSpec())
	})
}

// validateUpsert checks if a new data plane proxy can replace the old one.
// We cannot just upsert a new data plane proxy over the old one, because if you generate token bound to mesh + kuma.io/service
// then you would be able to just replace any other data plane proxy in the system.
// Ideally, when starting a new data plane proxy, the old one should be deleted, so we could do Create instead of Upsert, but this may not be the case.
// For example, if you spin down CP, then DP, then start CP, the old DP is still there for a couple of minutes (see pkg/gc).
// We could check if Dataplane is identical, but CP may alter Dataplane resource after is connected.
// What we do instead is that we use current data plane proxy credential to check if we can manage already registered Dataplane.
// The assumption is that if you have a token that can manage the old Dataplane.
// You can delete it and create a new one, so we can simplify this manual process by just replacing it.
func (d *DataplaneLifecycle) validateUpsert(ctx context.Context, existing model.Resource) error {
	if util_proto.IsEmpty(existing.GetSpec()) { // existing DP is empty, resource does not exist
		return nil
	}
	credential, err := xds_auth.ExtractCredential(ctx)
	if err != nil {
		return err
	}
	return d.authenticator.Authenticate(ctx, existing, credential)
}

func (d *DataplaneLifecycle) registerZoneIngress(ctx context.Context, zi *core_mesh.ZoneIngressResource) error {
	key := model.MetaToResourceKey(zi.GetMeta())
	existing := core_mesh.NewZoneIngressResource()
	return manager.Upsert(ctx, d.resManager, key, existing, func(resource model.Resource) error {
		if err := d.validateUpsert(ctx, existing); err != nil {
			return errors.Wrap(err, "you are trying to override existing zone ingress to which you don't have an access.")
		}
		return existing.SetSpec(zi.GetSpec())
	})
}

func (d *DataplaneLifecycle) registerZoneEgress(ctx context.Context, ze *core_mesh.ZoneEgressResource) error {
	key := model.MetaToResourceKey(ze.GetMeta())
	existing := core_mesh.NewZoneEgressResource()
	return manager.Upsert(ctx, d.resManager, key, existing, func(resource model.Resource) error {
		if err := d.validateUpsert(ctx, existing); err != nil {
			return errors.Wrap(err, "you are trying to override existing zone egress to which you don't have an access.")
		}
		return existing.SetSpec(ze.GetSpec())
	})
}

func (d *DataplaneLifecycle) deregisterProxy(
	ctx context.Context,
	key model.ResourceKey,
	obj model.Resource,
	insight model.Resource,
	log logr.Logger,
) error {
	log.Info("waiting for deregister proxy", "waitFor", d.deregistrationDelay)
	time.After(d.deregistrationDelay)
	// stream context is canceled at this point, because DP has disconnected. Use background context instead.
	err := d.resManager.Get(ctx, insight, store.GetBy(key))
	switch {
	case store.IsResourceNotFound(err):
		// If insight is missing it most likely means that it was not yet created, so DP just connected and now leaving the mesh.
		log.V(1).Info("insight is missing. Safe to deregister the proxy")
	case err != nil:
		return errors.Wrap(err, "could not get insight to determine if we can delete proxy object")
	default:
		sub := insight.GetSpec().(generic.Insight).GetLastSubscription().(*mesh_proto.DiscoverySubscription)
		if sub != nil && sub.ControlPlaneInstanceId != d.cpInstanceID {
			log.Info("no need to deregister proxy. It has already connected to another instance", "newCPInstanceID", sub.ControlPlaneInstanceId)
			return nil
		}
	}
	log.Info("deregister proxy")
	return d.resManager.Delete(ctx, obj, store.DeleteBy(key))
}
