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
	return r.validateDestinations(ctx, mesh, resource.Spec.Destinations)
}

func (r *RateLimitValidator) ValidateUpdate(ctx context.Context, previousRateLimit *core_mesh.RateLimitResource, newRateLimit *core_mesh.RateLimitResource) error {
	return r.validateDestinations(ctx, previousRateLimit.GetMeta().GetMesh(), newRateLimit.Spec.Destinations)
}

func (r *RateLimitValidator) ValidateDelete(ctx context.Context, name string) error {
	return nil
}

func (r *RateLimitValidator) validateDestinations(ctx context.Context, mesh string, dests []*mesh_proto.Selector) error {
	validationErr := &validators.ValidationError{}
	// A ratelimit on an external service can only match kuma.io/service tags
	for _, dest := range dests {
		hasNonService := false
		hasExternalService := false
		for tag, value := range dest.Match {
			if tag == "kuma.io/service" {
				svc := core_mesh.NewExternalServiceResource()
				err := r.Store.Get(ctx, svc, store.GetByKey(value, mesh))
				if err == nil {
					hasExternalService = true
				}
			} else {
				hasNonService = true
			}
		}
		if hasNonService && hasExternalService {
			validationErr.AddViolation("ratelimit", "rate limit applied to external service only supports kuma.io/service as destination match")
		}
	}
	return validationErr.OrNil()
}
