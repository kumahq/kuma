package callbacks

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/channels"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
)

// There are two possible workflows for managing lifecycles of Dataplane resources
// 1) apply Dataplane resource before kuma-dp run and run kuma-dp (this is the case with Kubernetes for example)
// 2) run kuma-dp and pass Dataplane resource as an argument to kuma-dp

// dataplaneLifecycle is responsible for creating and deleting dataplanes that are passed through metadata (option 2 above).
// When user passes Dataplane to kuma-dp it is attached to bootstrap request.
// Then, bootstrap server generates bootstrap configuration with Dataplane embedded in Envoy metadata.
// Here, we read Dataplane resource from metadata if it exists and then create or delete the Dataplane resource.
// This is not thread safe as it's meant to be called from DataplaneSyncCallback, it only has exceptions for DP reconnecting to a different CP.
type dataplaneLifecycle struct {
	resManager          manager.ResourceManager
	authenticator       xds_auth.Authenticator
	appCtx              context.Context
	deregistrationDelay time.Duration
	cpInstanceID        string
	cacheExpirationTime time.Duration
	proxyKey            core_model.ResourceKey
	tDesc               core_model.ResourceTypeDescriptor
	registered          bool
}

func DataplaneLifecycleFactory(
	appCtx context.Context,
	resManager manager.ResourceManager,
	authenticator xds_auth.Authenticator,
	deregistrationDelay time.Duration,
	cpInstanceID string,
	cacheExpirationTime time.Duration,
) DataplaneLifecycleManagerFactory {
	return DataplaneLifecycleManagerFunc(func(proxyKey core_model.ResourceKey) DataplaneLifecycleManager {
		return &dataplaneLifecycle{
			proxyKey:            proxyKey,
			resManager:          resManager,
			authenticator:       authenticator,
			appCtx:              appCtx,
			deregistrationDelay: deregistrationDelay,
			cpInstanceID:        cpInstanceID,
			cacheExpirationTime: cacheExpirationTime,
		}
	})
}

func (d *dataplaneLifecycle) Register(
	log logr.Logger,
	ctx context.Context,
	md *core_xds.DataplaneMetadata,
) error {
	if md.Resource == nil {
		// Noop when there's no resource in metadata
		return nil
	}
	d.registered = true
	l := log.WithName("lifecycle").
		WithValues("proxyType", md.GetProxyType()).
		WithValues("resource", md.Resource)

	tDesc, err := core_mesh.ResourceTypeDescriptorFromProxyType(md.GetProxyType())
	if err != nil {
		return err
	}
	d.tDesc = tDesc

	l.Info("register proxy")

	err = manager.Upsert(ctx, d.resManager, core_model.MetaToResourceKey(md.Resource.GetMeta()), d.tDesc.NewObject(), func(existing core_model.Resource) error {
		if err := d.validateUpsert(ctx, md.Resource); err != nil {
			return errors.Wrap(err, "you are trying to override existing proxy to which you don't have an access.")
		}
		return existing.SetSpec(md.Resource.GetSpec())
	})
	if err != nil {
		l.Info("cannot register proxy", "reason", err.Error())
		return errors.Wrap(err, "could not register proxy passed in kuma-dp run")
	}

	// Because in this case we've just created the dataplane, the MeshContext cache is not up to date.
	// So we block for the cache ExpirationTime to the MeshContext is refreshed before starting the
	// dataplane watchdog which will generate the XDS config.
	// See https://github.com/kumahq/kuma/pull/11180
	// This is not ideal but this cache is usually fairly short.
	time.Sleep(d.cacheExpirationTime)

	return nil
}

func (d *dataplaneLifecycle) Deregister(
	log logr.Logger,
	ctx context.Context,
) {
	if !d.registered {
		// We never registered this so don't care
		return
	}
	l := log.WithName("lifecycle").
		WithValues("proxyType", d.tDesc.Name)
	// Deregister() is called in 2 cases:
	// 1. connection between the DP to the CP is closed
	// 2. Kuma CP is gracefully shutting down.
	// This is for (2), in this case we must not delete the Dataplane resource because we assume the data plane proxy
	// will reconnect to another instance of Kuma CP.
	if channels.IsClosed(d.appCtx.Done()) {
		l.Info("graceful shutdown, don't delete Dataplane resource")
		return
	}

	// we avoid deleting the dataplane resource immediately because if the DP is connected to another CP.
	// a race could occur and we could end up deleting the resource used by the other CP
	l.Info("waiting for deregister proxy", "waitFor", d.deregistrationDelay)
	<-time.After(d.deregistrationDelay)

	insight := d.tDesc.NewInsight()
	err := d.resManager.Get(ctx, insight, store.GetBy(d.proxyKey))
	switch {
	case store.IsResourceNotFound(err):
		// If insight is missing it most likely means that it was not yet created, so DP just connected and now leaving the mesh.
		l.Info("insight is missing. Safe to deregister the proxy")
	case err != nil:
		l.Error(err, "could not get insight to determine if we can delete proxy object")
		return
	default: // There is an insight, let's check where it is connected to
		sub := insight.GetSpec().(generic.Insight).GetLastSubscription()
		if sub != nil && sub.(*mesh_proto.DiscoverySubscription).ControlPlaneInstanceId != d.cpInstanceID {
			l.Info("no need to deregister proxy. It has already connected to another instance", "lastSub", sub)
			return
		}
	}

	l.Info("deregister proxy")
	if err := d.resManager.Delete(ctx, d.tDesc.NewObject(), store.DeleteBy(d.proxyKey)); err != nil {
		l.Error(err, "could not deregister proxy")
	}
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
func (d *dataplaneLifecycle) validateUpsert(ctx context.Context, existing core_model.Resource) error {
	if core_model.IsEmpty(existing.GetSpec()) { // existing DP is empty, resource does not exist
		return nil
	}
	credential, err := xds_auth.ExtractCredential(ctx)
	if err != nil {
		return err
	}
	return d.authenticator.Authenticate(ctx, existing, credential)
}
