package store

import (
	"context"

	config_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func NewConfigStore(resourceStore core_store.ResourceStore) ConfigStore {
	return &configStore{
		resourceStore: resourceStore,
	}
}

var _ ConfigStore = &configStore{}

type configStore struct {
	resourceStore core_store.ResourceStore
}

func (r *configStore) Get(ctx context.Context, config *config_model.ConfigResource, fs ...core_store.GetOptionsFunc) error {
	return r.resourceStore.Get(ctx, config, fs...)
}

func (r *configStore) List(ctx context.Context, configs *config_model.ConfigResourceList, fs ...core_store.ListOptionsFunc) error {
	return r.resourceStore.List(ctx, configs, fs...)
}

func (r *configStore) Create(ctx context.Context, config *config_model.ConfigResource, fs ...core_store.CreateOptionsFunc) error {
	return r.resourceStore.Create(ctx, config, fs...)
}

func (r *configStore) Delete(ctx context.Context, config *config_model.ConfigResource, fs ...core_store.DeleteOptionsFunc) error {
	return r.resourceStore.Delete(ctx, config, fs...)
}

func (r *configStore) Update(ctx context.Context, config *config_model.ConfigResource, fs ...core_store.UpdateOptionsFunc) error {
	return r.resourceStore.Update(ctx, config, fs...)
}
