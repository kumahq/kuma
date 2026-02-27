package providers

import (
	"context"

	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
)

// This interface is a subject to changes based on other providers to find one that serves for all
type IdentityProvider interface {
	Validate(context.Context, *meshidentity_api.MeshIdentityResource) error
	Initialize(context.Context, *meshidentity_api.MeshIdentityResource) error
	CreateIdentity(context.Context, *meshidentity_api.MeshIdentityResource, *xds.Proxy) (*xds.WorkloadIdentity, error)
	// GetMeshTrustCA returns the CA bytes for MeshTrust creation.
	// Returning (nil, nil) means no MeshTrust should be created.
	GetMeshTrustCA(context.Context, *meshidentity_api.MeshIdentityResource) ([]byte, error)
}

type IdentityProviders = map[string]IdentityProvider
