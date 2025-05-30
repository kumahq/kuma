package ratelimit

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type RateLimitValidator struct {
	Store store.ResourceStore
}

func (r *RateLimitValidator) ValidateCreate(ctx context.Context, mesh string, resource *core_mesh.RateLimitResource) error {
	return r.validateDestinations(ctx, mesh, resource.Destinations())
}

func (r *RateLimitValidator) ValidateUpdate(ctx context.Context, previousRateLimit *core_mesh.RateLimitResource, newRateLimit *core_mesh.RateLimitResource) error {
	return r.validateDestinations(ctx, previousRateLimit.GetMeta().GetMesh(), newRateLimit.Destinations())
}

func (r *RateLimitValidator) ValidateDelete(ctx context.Context, name string) error {
	return nil
}

func (r *RateLimitValidator) validateDestinations(ctx context.Context, mesh string, dests []*mesh_proto.Selector) error {
	validationErr := &validators.ValidationError{}
	// A ratelimit on an external service can only match kuma.io/service tags
	externalServices := &core_mesh.ExternalServiceResourceList{}
	err := r.Store.List(ctx, externalServices, store.ListByMesh(mesh))
	if err != nil {
		if store.IsNotFound(err) {
			return nil
		} else {
			validationErr.AddViolation("ratelimit", err.Error())
			return validationErr
		}
	}

	esLookup := map[string]bool{}
	for _, externalService := range externalServices.Items {
		spec := externalService.GetSpec().(*mesh_proto.ExternalService)
		for tag, value := range spec.GetTags() {
			if tag == mesh_proto.ServiceTag {
				esLookup[value] = true
			}
		}
	}

	if len(esLookup) == 0 {
		return nil
	}

	for _, dest := range dests {
		hasNonService := false
		hasExternalService := false
		for tag, value := range dest.GetMatch() {
			if tag == mesh_proto.ServiceTag {
				if _, exist := esLookup[value]; exist || value == mesh_proto.MatchAllTag {
					hasExternalService = true
				}
			} else {
				hasNonService = true
			}
		}
		if hasNonService && hasExternalService {
			validationErr.AddViolation("ratelimit", "RateLimit applied to external service only supports kuma.io/service as destination match")
		}
	}
	return validationErr.OrNil()
}
