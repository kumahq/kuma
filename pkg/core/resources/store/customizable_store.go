package store

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceStoreWrapper = func(delegate ResourceStore) ResourceStore

type CustomizableResourceStore interface {
	ResourceStore
	ResourceStore(model.ResourceType) ResourceStore
	WrapAll(ResourceStoreWrapper)
}

func NewCustomizableResourceStore(defaultStore ResourceStore, customStores map[model.ResourceType]ResourceStore) CustomizableResourceStore {
	return &customizableResourceStore{
		defaultStore: defaultStore,
		customStores: customStores,
	}
}

var _ CustomizableResourceStore = &customizableResourceStore{}

type customizableResourceStore struct {
	defaultStore ResourceStore
	customStores map[model.ResourceType]ResourceStore
}

func (m *customizableResourceStore) Get(ctx context.Context, resource model.Resource, fs ...GetOptionsFunc) error {
	return m.ResourceStore(resource.Descriptor().Name).Get(ctx, resource, fs...)
}

func (m *customizableResourceStore) List(ctx context.Context, list model.ResourceList, fs ...ListOptionsFunc) error {
	return m.ResourceStore(list.GetItemType()).List(ctx, list, fs...)
}

func (m *customizableResourceStore) Create(ctx context.Context, resource model.Resource, fs ...CreateOptionsFunc) error {
	return m.ResourceStore(resource.Descriptor().Name).Create(ctx, resource, fs...)
}

func (m *customizableResourceStore) Delete(ctx context.Context, resource model.Resource, fs ...DeleteOptionsFunc) error {
	return m.ResourceStore(resource.Descriptor().Name).Delete(ctx, resource, fs...)
}

func (m *customizableResourceStore) Update(ctx context.Context, resource model.Resource, fs ...UpdateOptionsFunc) error {
	return m.ResourceStore(resource.Descriptor().Name).Update(ctx, resource, fs...)
}

func (m *customizableResourceStore) ResourceStore(typ model.ResourceType) ResourceStore {
	if customManager, ok := m.customStores[typ]; ok {
		return customManager
	}
	return m.defaultStore
}

func (m *customizableResourceStore) WrapAll(wrapper ResourceStoreWrapper) {
	m.defaultStore = wrapper(m.defaultStore)
	for typ, store := range m.customStores {
		m.customStores[typ] = wrapper(store)
	}
}
