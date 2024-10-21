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
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/maps"
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
	resManager          manager.ResourceManager
	authenticator       xds_auth.Authenticator
	proxyInfos          maps.Sync[core_model.ResourceKey, *proxyInfo]
	appCtx              context.Context
	deregistrationDelay time.Duration
	cpInstanceID        string
	cacheExpirationTime time.Duration
}

type proxyInfo struct {
	mtx       sync.Mutex
	proxyType mesh_proto.ProxyType
	connected bool
	deleted   bool
}

var _ DataplaneCallbacks = &DataplaneLifecycle{}

func NewDataplaneLifecycle(
	appCtx context.Context,
	resManager manager.ResourceManager,
	authenticator xds_auth.Authenticator,
	deregistrationDelay time.Duration,
	cpInstanceID string,
	cacheExpirationTime time.Duration,
) *DataplaneLifecycle {
	return &DataplaneLifecycle{
		resManager:          resManager,
		authenticator:       authenticator,
		proxyInfos:          maps.Sync[core_model.ResourceKey, *proxyInfo]{},
		appCtx:              appCtx,
		deregistrationDelay: deregistrationDelay,
		cpInstanceID:        cpInstanceID,
		cacheExpirationTime: cacheExpirationTime,
	}
}

func (d *DataplaneLifecycle) OnProxyConnected(streamID core_xds.StreamID, proxyKey core_model.ResourceKey, ctx context.Context, md core_xds.DataplaneMetadata) error {
	if md.Resource == nil {
		return nil
	}
	if err := d.validateProxyKey(proxyKey, md.Resource); err != nil {
		return err
	}
	return d.register(ctx, streamID, proxyKey, md)
}

func (d *DataplaneLifecycle) OnProxyReconnected(streamID core_xds.StreamID, proxyKey core_model.ResourceKey, ctx context.Context, md core_xds.DataplaneMetadata) error {
	if md.Resource == nil {
		return nil
	}
	if err := d.validateProxyKey(proxyKey, md.Resource); err != nil {
		return err
	}
	return d.register(ctx, streamID, proxyKey, md)
}

func (d *DataplaneLifecycle) OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, proxyKey core_model.ResourceKey) {
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

	d.deregister(ctx, streamID, proxyKey)
}

func (d *DataplaneLifecycle) register(
	ctx context.Context,
	streamID core_xds.StreamID,
	proxyKey core_model.ResourceKey,
	md core_xds.DataplaneMetadata,
) error {
	log := lifecycleLog.
		WithValues("proxyType", md.GetProxyType()).
		WithValues("proxyKey", proxyKey).
		WithValues("streamID", streamID).
		WithValues("resource", md.Resource)

	info, loaded := d.proxyInfos.LoadOrStore(proxyKey, &proxyInfo{
		proxyType: md.GetProxyType(),
	})

	info.mtx.Lock()
	defer info.mtx.Unlock()

	if info.deleted {
		// we took info object that was deleted from proxyInfo map by other goroutine, return err so DPP retry registration
		return errors.Errorf("attempt to concurently register deleted DPP resource, needs retry")
	}

	log.Info("register proxy")

	err := manager.Upsert(ctx, d.resManager, core_model.MetaToResourceKey(md.Resource.GetMeta()), proxyResource(md.GetProxyType()), func(existing core_model.Resource) error {
		if err := d.validateUpsert(ctx, md.Resource); err != nil {
			return errors.Wrap(err, "you are trying to override existing proxy to which you don't have an access.")
		}
		return existing.SetSpec(md.Resource.GetSpec())
	})
	if err != nil {
		log.Info("cannot register proxy", "reason", err.Error())
		if !loaded {
			info.deleted = true
			d.proxyInfos.Delete(proxyKey)
		}
		return errors.Wrap(err, "could not register proxy passed in kuma-dp run")
	}

	// We should wait for Cache ExpirationTime to let MeshContext sync the latest data
	// in which the latter DataplaneSyncTracker callback relies on it to generate the XDS configuration.
	// This only happens on Universal Dataplane because the k8s Dataplane object is created by the controller.
	time.Sleep(d.cacheExpirationTime)
	info.connected = true

	return nil
}

func (d *DataplaneLifecycle) deregister(
	ctx context.Context,
	streamID core_xds.StreamID,
	proxyKey core_model.ResourceKey,
) {
	info, ok := d.proxyInfos.Load(proxyKey)
	if !ok {
		// proxy was not registered with this callback
		return
	}

	info.mtx.Lock()
	if info.deleted {
		info.mtx.Unlock()
		return
	}

	info.connected = false
	proxyType := info.proxyType
	info.mtx.Unlock()

	log := lifecycleLog.
		WithValues("proxyType", proxyType).
		WithValues("proxyKey", proxyKey).
		WithValues("streamID", streamID)

	// if delete immediately we're more likely to have a race condition
	// when DPP is connected to another CP but proxy resource in the store is deleted
	log.Info("waiting for deregister proxy", "waitFor", d.deregistrationDelay)
	<-time.After(d.deregistrationDelay)

	info.mtx.Lock()
	defer info.mtx.Unlock()

	if info.deleted {
		return
	}

	if info.connected {
		log.Info("no need to deregister proxy. It has already connected to this instance")
		return
	}

	if connected, err := d.proxyConnectedToAnotherCP(ctx, proxyType, proxyKey, log); err != nil {
		log.Error(err, "could not check if proxy connected to another CP")
		return
	} else if connected {
		return
	}

	log.Info("deregister proxy")
	if err := d.resManager.Delete(ctx, proxyResource(proxyType), store.DeleteBy(proxyKey)); err != nil {
		log.Error(err, "could not unregister proxy")
	}

	d.proxyInfos.Delete(proxyKey)
	info.deleted = true
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
func (d *DataplaneLifecycle) validateUpsert(ctx context.Context, existing core_model.Resource) error {
	if core_model.IsEmpty(existing.GetSpec()) { // existing DP is empty, resource does not exist
		return nil
	}
	credential, err := xds_auth.ExtractCredential(ctx)
	if err != nil {
		return err
	}
	return d.authenticator.Authenticate(ctx, existing, credential)
}

func (d *DataplaneLifecycle) validateProxyKey(proxyKey core_model.ResourceKey, proxyResource core_model.Resource) error {
	if core_model.MetaToResourceKey(proxyResource.GetMeta()) != proxyKey {
		return errors.Errorf("proxyId %s does not match proxy resource %s", proxyKey, proxyResource.GetMeta())
	}
	return nil
}

func (d *DataplaneLifecycle) proxyConnectedToAnotherCP(
	ctx context.Context,
	pt mesh_proto.ProxyType,
	key core_model.ResourceKey,
	log logr.Logger,
) (bool, error) {
	insight := proxyInsight(pt)

	err := d.resManager.Get(ctx, insight, store.GetBy(key))
	switch {
	case store.IsResourceNotFound(err):
		// If insight is missing it most likely means that it was not yet created, so DP just connected and now leaving the mesh.
		log.Info("insight is missing. Safe to deregister the proxy")
		return false, nil
	case err != nil:
		return false, errors.Wrap(err, "could not get insight to determine if we can delete proxy object")
	}

	subs := insight.GetSpec().(generic.Insight).AllSubscriptions()
	if len(subs) == 0 {
		return false, nil
	}

	if sub := subs[len(subs)-1].(*mesh_proto.DiscoverySubscription); sub.ControlPlaneInstanceId != d.cpInstanceID {
		log.Info("no need to deregister proxy. It has already connected to another instance", "newCPInstanceID", sub.ControlPlaneInstanceId)
		return true, nil
	}

	return false, nil
}

func proxyResource(pt mesh_proto.ProxyType) core_model.Resource {
	switch pt {
	case mesh_proto.DataplaneProxyType:
		return core_mesh.NewDataplaneResource()
	case mesh_proto.IngressProxyType:
		return core_mesh.NewZoneIngressResource()
	case mesh_proto.EgressProxyType:
		return core_mesh.NewZoneEgressResource()
	default:
		return nil
	}
}

func proxyInsight(pt mesh_proto.ProxyType) core_model.Resource {
	switch pt {
	case mesh_proto.DataplaneProxyType:
		return core_mesh.NewDataplaneInsightResource()
	case mesh_proto.IngressProxyType:
		return core_mesh.NewZoneIngressInsightResource()
	case mesh_proto.EgressProxyType:
		return core_mesh.NewZoneEgressInsightResource()
	default:
		return nil
	}
}
