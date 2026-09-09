package providers

import (
	"context"

	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
)

// This interface is a subject to changes based on other providers to find one that serves for all
type IdentityProvider interface {
	Validate(context.Context, *meshidentity_api.MeshIdentityResource) error
	Initialize(context.Context, *meshidentity_api.MeshIdentityResource) error
	// CreateIdentity issues an identity in the given trust domain. The trust domain
	// is resolved by the caller so that it can never run ahead of the MeshTrust that
	// publishes the matching CA bundle.
	CreateIdentity(context.Context, *meshidentity_api.MeshIdentityResource, *xds.Proxy, string) (*xds.WorkloadIdentity, error)
	// GetMeshTrustCA returns the CA bytes for MeshTrust creation.
	// Returning (nil, nil) means no MeshTrust should be created.
	GetMeshTrustCA(context.Context, *meshidentity_api.MeshIdentityResource) ([]byte, error)
}

type IdentityProviders = map[string]IdentityProvider
