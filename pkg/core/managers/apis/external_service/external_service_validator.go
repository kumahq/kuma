package externalservice

import (
	"context"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
)

type ExternalServiceValidator struct {
	Store store.ResourceStore
}

func (r *ExternalServiceValidator) ValidateCreate(ctx context.Context, mesh string, resource *core_mesh.ExternalServiceResource) error {
	return nil
}

func (r *ExternalServiceValidator) ValidateUpdate(ctx context.Context, previousExternalService *core_mesh.ExternalServiceResource, newExternalService *core_mesh.ExternalServiceResource) error {
	return nil
}

func (r *ExternalServiceValidator) ValidateDelete(ctx context.Context, name string) error {
	return nil
}
