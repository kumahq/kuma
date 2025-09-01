package providers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/events"
)

type IdentityProviderManager struct {
	logger      logr.Logger
	eventWriter events.Emitter
	providers   IdentityProviders
}

func NewIdentityProviderManager(providers IdentityProviders, eventWriter events.Emitter) IdentityProviderManager {
	logger := core.Log.WithName("identity-provider")
	return IdentityProviderManager{
		logger:      logger,
		eventWriter: eventWriter,
		providers:   providers,
	}
}

func (i *IdentityProviderManager) SelectedIdentity(dataplane *core_mesh.DataplaneResource, identities []*meshidentity_api.MeshIdentityResource) *meshidentity_api.MeshIdentityResource {
	identity, _ := meshidentity_api.Matched(dataplane.Meta.GetLabels(), identities)
	return identity
}

func (i *IdentityProviderManager) GetWorkloadIdentity(ctx context.Context, proxy *xds.Proxy, identity *meshidentity_api.MeshIdentityResource) (*xds.WorkloadIdentity, error) {
	if identity == nil {
		i.eventWriter.Send(events.WorkloadIdentityChangedEvent{
			ResourceKey: model.MetaToResourceKey(proxy.Dataplane.GetMeta()),
			Operation:   events.Delete,
		})
		return nil, nil
	}

	if !identity.Status.IsInitialized() {
		i.logger.V(1).Info("identity hasn't been initialized yet", "identity", identity.Meta.GetName())
		i.eventWriter.Send(events.WorkloadIdentityChangedEvent{
			ResourceKey: model.MetaToResourceKey(proxy.Dataplane.GetMeta()),
			Operation:   events.Delete,
		})
		return nil, nil
	}
	i.logger.V(1).Info("providing identity", "identity", identity.Meta.GetName(), "dataplane", proxy.Dataplane.Meta.GetName())
	provider, found := i.providers[string(identity.Spec.Provider.Type)]
	if !found {
		return nil, fmt.Errorf("identity provider %s not found", identity.Spec.Provider.Type)
	}
	workloadIdentity, err := provider.CreateIdentity(ctx, identity, proxy)
	if err != nil {
		return nil, err
	}
	if workloadIdentity == nil {
		return nil, nil
	}
	event := events.WorkloadIdentityChangedEvent{
		ResourceKey: model.MetaToResourceKey(proxy.Dataplane.GetMeta()),
		Operation:   events.Create,
		Origin:      workloadIdentity.KRI,
	}
	if workloadIdentity.ManagementMode == xds.KumaManagementMode {
		event.ExpirationTime = workloadIdentity.ExpirationTime
		event.GenerationTime = workloadIdentity.GenerationTime
	}
	i.eventWriter.Send(event)
	return workloadIdentity, nil
}
