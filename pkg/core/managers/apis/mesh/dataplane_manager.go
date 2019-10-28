package mesh

import (
	"context"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	validators "github.com/Kong/kuma/pkg/core/validators/apis/mesh"
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

type dataplaneManager struct {
	store core_store.ResourceStore
}

func NewDataplaneManager(store core_store.ResourceStore) core_manager.ResourceManager {
	return &dataplaneManager{store}
}

func (d *dataplaneManager) Create(ctx context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	dp, err := dataplane(resource)
	if err != nil {
		return err
	}
	if err := validators.ValidateDataplane(dp); err != nil {
		return err
	}
	return d.store.Create(ctx, dp, fs...)
}

func (d *dataplaneManager) Update(ctx context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	dp, err := dataplane(resource)
	if err != nil {
		return err
	}
	if err := validators.ValidateDataplane(dp); err != nil {
		return err
	}
	return d.store.Update(ctx, dp, fs...)
}

func (d *dataplaneManager) Delete(ctx context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	return d.store.Delete(ctx, resource, fs...)
}

func (d *dataplaneManager) DeleteAll(ctx context.Context, list model.ResourceList, fs ...store.DeleteAllOptionsFunc) error {
	return core_manager.DeleteAllResources(d, ctx, list, fs...)
}

func (d *dataplaneManager) Get(ctx context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	return d.store.Get(ctx, resource, fs...)
}

func (d *dataplaneManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	return d.store.List(ctx, list, fs...)
}

func dataplane(resource core_model.Resource) (*core_mesh.DataplaneResource, error) {
	dp, ok := resource.(*core_mesh.DataplaneResource)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.DataplaneResource)(nil), resource)
	}
	return dp, nil
}

var _ core_manager.ResourceManager = &dataplaneManager{}
