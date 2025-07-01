package externalservice

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type ExternalServiceValidator struct {
	Store store.ResourceStore
}

func (r *ExternalServiceValidator) ValidateCreate(ctx context.Context, mesh string, resource *core_mesh.ExternalServiceResource) error {
	return r.validateRateLimits(ctx, mesh, resource.Spec.GetTags()[mesh_proto.ServiceTag])
}

func (r *ExternalServiceValidator) ValidateUpdate(ctx context.Context, previousExternalService *core_mesh.ExternalServiceResource, newExternalService *core_mesh.ExternalServiceResource) error {
	return r.validateRateLimits(ctx, previousExternalService.GetMeta().GetMesh(), newExternalService.Spec.GetTags()[mesh_proto.ServiceTag])
}

func (r *ExternalServiceValidator) ValidateDelete(ctx context.Context, name string) error {
	return nil
}

func (r *ExternalServiceValidator) validateRateLimits(ctx context.Context, mesh string, service string) error {
	validationErr := &validators.ValidationError{}
	// If referenced by a ratelimit, that ratelimit must only match kuma.io/service destinations

	rateLimits := &core_mesh.RateLimitResourceList{}
	err := r.Store.List(ctx, rateLimits, store.ListByMesh(mesh))
	if err != nil {
		if store.IsNotFound(err) {
			return nil
		} else {
			validationErr.AddViolation("externalservice", err.Error())
			return validationErr
		}
	}

	for _, rateLimit := range rateLimits.Items {
		for _, dest := range rateLimit.Destinations() {
			matchesThisService := false
			matchesNonService := false
			for tag, value := range dest.GetMatch() {
				if tag == mesh_proto.ServiceTag {
					if value == service || value == mesh_proto.MatchAllTag {
						matchesThisService = true
					}
				} else {
					matchesNonService = true
				}
			}
			if matchesThisService && matchesNonService {
				rlName := rateLimit.GetMeta().GetName()
				validationErr.AddViolation("externalservice", "ExternalService would match RateLimit '"+rlName+"' which includes incompatible destination matches")
			}
		}
	}

	return validationErr.OrNil()
}
