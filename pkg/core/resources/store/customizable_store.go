package store

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// ResourceStoreWrapper is a function that takes a ResourceStore and returns a wrapped ResourceStore.
// The wrapped ResourceStore can be used to modify or augment the behavior of the original ResourceStore.
type ResourceStoreWrapper = func(delegate ResourceStore) ResourceStore

type CustomizableResourceStore interface {
	ResourceStore
	ResourceStore(typ model.ResourceType) ResourceStore
	DefaultResourceStore() ResourceStore
	Customize(typ model.ResourceType, store ResourceStore)
	WrapAll(wrapper ResourceStoreWrapper)
}

func NewCustomizableResourceStore(defaultStore ResourceStore) CustomizableResourceStore {
	return &customizableResourceStore{
		defaultStore: defaultStore,
		customStores: map[model.ResourceType]ResourceStore{},
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
	if customStore, ok := m.customStores[typ]; ok {
		return customStore
	}
	return m.defaultStore
}

func (m *customizableResourceStore) DefaultResourceStore() ResourceStore {
	return m.defaultStore
}

// Customize installs a new store for the given type. If a store of the specified type already exists, it is overwritten.
func (m *customizableResourceStore) Customize(typ model.ResourceType, store ResourceStore) {
	m.customStores[typ] = store
}

// WrapAll function wraps the default and all custom ResourceStores with the provided ResourceStoreWrapper function.
// This means that all future accesses to these ResourceStores will go through the ResourceStoreWrapper function,
// which can be used to modify or augment the behavior of the ResourceStores.
func (m *customizableResourceStore) WrapAll(wrapper ResourceStoreWrapper) {
	m.defaultStore = wrapper(m.defaultStore)
	for typ, store := range m.customStores {
		m.customStores[typ] = wrapper(store)
	}
}
