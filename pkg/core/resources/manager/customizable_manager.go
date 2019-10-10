package manager

import (
	"context"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func NewCustomizableResourceManager(defaultManager ResourceManager, customManagers map[model.ResourceType]ResourceManager, typRegistry registry.TypeRegistry) ResourceManager {
	return &customizableResourceManager{
		defaultManager: defaultManager,
		customManagers: customManagers,
		registry:       typRegistry,
	}
}

type customizableResourceManager struct {
	defaultManager ResourceManager
	customManagers map[model.ResourceType]ResourceManager
	registry       registry.TypeRegistry
}

func (m *customizableResourceManager) Get(ctx context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	return m.resourceManager(resource.GetType()).Get(ctx, resource, fs...)
}

func (m *customizableResourceManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	return m.resourceManager(list.GetItemType()).List(ctx, list, fs...)
}

func (m *customizableResourceManager) Create(ctx context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	return m.resourceManager(resource.GetType()).Create(ctx, resource, fs...)
}

func (m *customizableResourceManager) Delete(ctx context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	return m.resourceManager(resource.GetType()).Delete(ctx, resource, fs...)
}

func (m *customizableResourceManager) DeleteAll(ctx context.Context, fs ...store.DeleteAllOptionsFunc) error {
	for _, typ := range m.registry.ListTypes() {
		if err := m.resourceManager(typ).DeleteAll(ctx, fs...); err != nil {
			return err
		}
	}
	return nil
}

func (m *customizableResourceManager) Update(ctx context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	return m.resourceManager(resource.GetType()).Update(ctx, resource, fs...)
}

func (m *customizableResourceManager) resourceManager(typ model.ResourceType) ResourceManager {
	if customManager, ok := m.customManagers[typ]; ok {
		return customManager
	}
	return m.defaultManager
}
