package providers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/issuer"
	"github.com/kumahq/kuma/v3/pkg/events"
)

type IdentityProviderManager struct {
	logger      logr.Logger
	eventWriter events.Emitter
	providers   IdentityProviders
	limiter     issuer.Limiter
}

func NewIdentityProviderManager(providers IdentityProviders, eventWriter events.Emitter, limiter issuer.Limiter) IdentityProviderManager {
	logger := core.Log.WithName("identity-provider")
	return IdentityProviderManager{
		logger:      logger,
		eventWriter: eventWriter,
		providers:   providers,
		limiter:     limiter,
	}
}

func (i *IdentityProviderManager) SelectedIdentity(dataplane *core_mesh.DataplaneResource, identities []*meshidentity_api.MeshIdentityResource) *meshidentity_api.MeshIdentityResource {
	identity, _ := meshidentity_api.BestMatched(dataplane.Meta.GetLabels(), identities)
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

	// Throttle issuance: per-proxy backoff plus a per-MeshIdentity circuit
	// breaker, so a failing provider (e.g. a misconfigured ACM Private CA) can't
	// be hammered on every DP sync tick.
	backend := kri.From(identity)
	source := model.MetaToResourceKey(proxy.Dataplane.GetMeta())
	if ok, retryAfter := i.limiter.Allow(backend, source); !ok {
		return nil, fmt.Errorf("backing off identity generation for %q after a previous failure (retry after %s)", identity.Meta.GetName(), retryAfter)
	}
	workloadIdentity, err := provider.CreateIdentity(ctx, identity, proxy)
	i.limiter.Record(backend, source, err == nil)
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

// Cleanup drops the proxy's issuance backoff state, e.g. when a dataplane
// disconnects. The dataplane watchdog calls this so MeshIdentity backoff is
// cleaned up explicitly, not just incidentally via the shared limiter and the
// legacy mTLS Cleanup path.
func (i *IdentityProviderManager) Cleanup(dpKey model.ResourceKey) {
	i.limiter.Forget(dpKey)
}
