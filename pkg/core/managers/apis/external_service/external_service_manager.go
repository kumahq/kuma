package externalservice

import (
	"context"
	"time"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type externalServiceManager struct {
	store                    core_store.ResourceStore
	externalServiceValidator ExternalServiceValidator
}

func NewExternalServiceManager(
	store core_store.ResourceStore,
	validator ExternalServiceValidator,
) core_manager.ResourceManager {
	return &externalServiceManager{
		store:                    store,
		externalServiceValidator: validator,
	}
}

func (m *externalServiceManager) Get(ctx context.Context, resource core_model.Resource, fs ...core_store.GetOptionsFunc) error {
	externalService, err := m.externalService(resource)
	if err != nil {
		return err
	}
	return m.store.Get(ctx, externalService, fs...)
}

func (m *externalServiceManager) List(ctx context.Context, list core_model.ResourceList, fs ...core_store.ListOptionsFunc) error {
	externalServices, err := m.externalServices(list)
	if err != nil {
		return err
	}
	return m.store.List(ctx, externalServices, fs...)
}

func (m *externalServiceManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	opts := core_store.NewCreateOptions(fs...)
	externalService, err := m.externalService(resource)
	if err != nil {
		return err
	}
	if err := core_model.Validate(resource); err != nil {
		return err
	}
	if err := m.externalServiceValidator.ValidateCreate(ctx, opts.Mesh, externalService); err != nil {
		return err
	}

	if err := m.store.Create(ctx, externalService, append(fs, core_store.CreatedAt(time.Now()))...); err != nil {
		return err
	}
	return nil
}

func (m *externalServiceManager) Delete(ctx context.Context, resource core_model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	externalService, err := m.externalService(resource)
	if err != nil {
		return err
	}
	opts := core_store.NewDeleteOptions(fs...)

	if err := m.externalServiceValidator.ValidateDelete(ctx, opts.Name); err != nil {
		return err
	}
	if err := m.store.Delete(ctx, externalService, fs...); err != nil {
		return err
	}
	return nil
}

func (m *externalServiceManager) DeleteAll(ctx context.Context, list core_model.ResourceList, fs ...core_store.DeleteAllOptionsFunc) error {
	if _, err := m.externalServices(list); err != nil {
		return err
	}
	return core_manager.DeleteAllResources(m, ctx, list, fs...)
}

func (m *externalServiceManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	externalService, err := m.externalService(resource)
	if err != nil {
		return err
	}
	if err := core_model.Validate(resource); err != nil {
		return err
	}

	currentExternalService := core_mesh.NewExternalServiceResource()
	if err := m.Get(ctx, currentExternalService, core_store.GetBy(core_model.MetaToResourceKey(externalService.GetMeta())), core_store.GetByVersion(externalService.GetMeta().GetVersion())); err != nil {
		return err
	}
	if err := m.externalServiceValidator.ValidateUpdate(ctx, currentExternalService, externalService); err != nil {
		return err
	}

	return m.store.Update(ctx, externalService, append(fs, core_store.ModifiedAt(time.Now()))...)
}

func (m *externalServiceManager) externalService(resource core_model.Resource) (*core_mesh.ExternalServiceResource, error) {
	externalService, ok := resource.(*core_mesh.ExternalServiceResource)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.ExternalServiceResource)(nil), resource)
	}
	return externalService, nil
}

func (m *externalServiceManager) externalServices(list core_model.ResourceList) (*core_mesh.ExternalServiceResourceList, error) {
	externalServices, ok := list.(*core_mesh.ExternalServiceResourceList)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.ExternalServiceResourceList)(nil), list)
	}
	return externalServices, nil
}
