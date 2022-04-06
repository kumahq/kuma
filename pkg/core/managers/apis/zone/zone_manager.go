package zone

import (
	"context"

	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewZoneManager(store core_store.ResourceStore, validator Validator, unsafeDelete bool) core_manager.ResourceManager {
	return &zoneManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		validator:       validator,
		unsafeDelete:    unsafeDelete,
	}
}

type zoneManager struct {
	core_manager.ResourceManager
	store        core_store.ResourceStore
	validator    Validator
	unsafeDelete bool
}

func (z *zoneManager) Delete(ctx context.Context, r model.Resource, opts ...core_store.DeleteOptionsFunc) error {
	options := core_store.NewDeleteOptions(opts...)
	if !z.unsafeDelete {
		if err := z.validator.ValidateDelete(ctx, options.Name); err != nil {
			return err
		}
	}
	return z.ResourceManager.Delete(ctx, r, opts...)
}

func (z *zoneManager) DeleteAll(ctx context.Context, rl model.ResourceList, opts ...core_store.DeleteAllOptionsFunc) error {
	return core_manager.DeleteAllResources(z, ctx, rl, opts...)
}
