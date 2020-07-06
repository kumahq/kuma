package store

import (
	"context"

	"github.com/Kong/kuma/pkg/core/resources/model"
)

func NewCustomizableResourceStore(defaultStore ResourceStore, customStores map[model.ResourceType]ResourceStore) ResourceStore {
	return &customizableResourceStore{
		defaultStore: defaultStore,
		customStores: customStores,
	}
}

type customizableResourceStore struct {
	defaultStore ResourceStore
	customStores map[model.ResourceType]ResourceStore
}

func (m *customizableResourceStore) Get(ctx context.Context, resource model.Resource, fs ...GetOptionsFunc) error {
	return m.ResourceStore(resource.GetType()).Get(ctx, resource, fs...)
}

func (m *customizableResourceStore) List(ctx context.Context, list model.ResourceList, fs ...ListOptionsFunc) error {
	return m.ResourceStore(list.GetItemType()).List(ctx, list, fs...)
}

func (m *customizableResourceStore) Create(ctx context.Context, resource model.Resource, fs ...CreateOptionsFunc) error {
	return m.ResourceStore(resource.GetType()).Create(ctx, resource, fs...)
}

func (m *customizableResourceStore) Delete(ctx context.Context, resource model.Resource, fs ...DeleteOptionsFunc) error {
	return m.ResourceStore(resource.GetType()).Delete(ctx, resource, fs...)
}

func (m *customizableResourceStore) Update(ctx context.Context, resource model.Resource, fs ...UpdateOptionsFunc) error {
	return m.ResourceStore(resource.GetType()).Update(ctx, resource, fs...)
}

func (m *customizableResourceStore) ResourceStore(typ model.ResourceType) ResourceStore {
	if customManager, ok := m.customStores[typ]; ok {
		return customManager
	}
	return m.defaultStore
}
