package providers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
)

type IdentityProviderManager struct {
	logger    logr.Logger
	providers IdentityProviders
}

func NewIdentityProviderManager(providers IdentityProviders) IdentityProviderManager {
	logger := core.Log.WithName("identity-provider")
	return IdentityProviderManager{
		logger:    logger,
		providers: providers,
	}
}

func (i *IdentityProviderManager) SelectedIdentity(dataplane *core_mesh.DataplaneResource, identities []*meshidentity_api.MeshIdentityResource) *meshidentity_api.MeshIdentityResource {
	identity, _ := meshidentity_api.Matched(dataplane.Meta.GetLabels(), identities)
	return identity
}

func (i *IdentityProviderManager) GetWorkloadIdentity(ctx context.Context, proxy *xds.Proxy, identity *meshidentity_api.MeshIdentityResource) (*xds.WorkloadIdentity, error) {
	if identity == nil {
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
	return workloadIdentity, nil
}
