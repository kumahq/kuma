package ratelimit

import (
	"context"
	"time"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type rateLimitManager struct {
	store              core_store.ResourceStore
	rateLimitValidator RateLimitValidator
}

func NewRateLimitManager(
	store core_store.ResourceStore,
	validator RateLimitValidator,
) core_manager.ResourceManager {
	return &rateLimitManager{
		store:              store,
		rateLimitValidator: validator,
	}
}

func (m *rateLimitManager) Get(ctx context.Context, resource core_model.Resource, fs ...core_store.GetOptionsFunc) error {
	rateLimit, err := m.rateLimit(resource)
	if err != nil {
		return err
	}
	return m.store.Get(ctx, rateLimit, fs...)
}

func (m *rateLimitManager) List(ctx context.Context, list core_model.ResourceList, fs ...core_store.ListOptionsFunc) error {
	rateLimits, err := m.rateLimits(list)
	if err != nil {
		return err
	}
	return m.store.List(ctx, rateLimits, fs...)
}

func (m *rateLimitManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	opts := core_store.NewCreateOptions(fs...)
	rateLimit, err := m.rateLimit(resource)
	if err != nil {
		return err
	}
	if err := core_model.Validate(resource); err != nil {
		return err
	}
	if err := m.rateLimitValidator.ValidateCreate(ctx, opts.Mesh, rateLimit); err != nil {
		return err
	}

	if err := m.store.Create(ctx, rateLimit, append(fs, core_store.CreatedAt(time.Now()))...); err != nil {
		return err
	}
	return nil
}

func (m *rateLimitManager) Delete(ctx context.Context, resource core_model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	rateLimit, err := m.rateLimit(resource)
	if err != nil {
		return err
	}
	opts := core_store.NewDeleteOptions(fs...)

	if err := m.rateLimitValidator.ValidateDelete(ctx, opts.Name); err != nil {
		return err
	}
	if err := m.store.Delete(ctx, rateLimit, fs...); err != nil {
		return err
	}
	return nil
}

func (m *rateLimitManager) DeleteAll(ctx context.Context, list core_model.ResourceList, fs ...core_store.DeleteAllOptionsFunc) error {
	if _, err := m.rateLimits(list); err != nil {
		return err
	}
	return core_manager.DeleteAllResources(m, ctx, list, fs...)
}

func (m *rateLimitManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	rateLimit, err := m.rateLimit(resource)
	if err != nil {
		return err
	}
	if err := core_model.Validate(resource); err != nil {
		return err
	}

	currentRateLimit := core_mesh.NewRateLimitResource()
	if err := m.Get(ctx, currentRateLimit, core_store.GetBy(core_model.MetaToResourceKey(rateLimit.GetMeta())), core_store.GetByVersion(rateLimit.GetMeta().GetVersion())); err != nil {
		return err
	}
	if err := m.rateLimitValidator.ValidateUpdate(ctx, currentRateLimit, rateLimit); err != nil {
		return err
	}

	return m.store.Update(ctx, rateLimit, append(fs, core_store.ModifiedAt(time.Now()))...)
}

func (m *rateLimitManager) rateLimit(resource core_model.Resource) (*core_mesh.RateLimitResource, error) {
	rateLimit, ok := resource.(*core_mesh.RateLimitResource)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.RateLimitResource)(nil), resource)
	}
	return rateLimit, nil
}

func (m *rateLimitManager) rateLimits(list core_model.ResourceList) (*core_mesh.RateLimitResourceList, error) {
	rateLimits, ok := list.(*core_mesh.RateLimitResourceList)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.RateLimitResourceList)(nil), list)
	}
	return rateLimits, nil
}
