package rbac

import (
	"context"

	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
)

type rbacManager struct {
	core_manager.ResourceManager
	resourceAccess ResourceAccess
}

var _ core_manager.ResourceManager = &rbacManager{}

func NewRBACResourceManager(resourceManager core_manager.ResourceManager, resourceAccess ResourceAccess) core_manager.ResourceManager {
	return &rbacManager{
		ResourceManager: resourceManager,
		resourceAccess:  resourceAccess,
	}
}

func (r *rbacManager) Get(ctx context.Context, resource model.Resource, optionsFunc ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(optionsFunc...)
	key := model.ResourceKey{
		Mesh: opts.Name,
		Name: opts.Name,
	}
	if err := r.resourceAccess.ValidateGet(key, resource.Descriptor(), user.UserFromCtx(ctx)); err != nil {
		return err
	}
	return r.ResourceManager.Get(ctx, resource, optionsFunc...)
}

func (r *rbacManager) Create(ctx context.Context, resource model.Resource, optionsFunc ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(optionsFunc...)
	key := model.ResourceKey{
		Mesh: opts.Name,
		Name: opts.Name,
	}
	if err := r.resourceAccess.ValidateCreate(key, resource.GetSpec(), resource.Descriptor(), user.UserFromCtx(ctx)); err != nil {
		return err
	}
	return r.ResourceManager.Create(ctx, resource, optionsFunc...)
}

func (r *rbacManager) List(ctx context.Context, list model.ResourceList, optionsFunc ...store.ListOptionsFunc) error {
	if err := r.resourceAccess.ValidateList(list.NewItem().Descriptor(), user.UserFromCtx(ctx)); err != nil {
		return err
	}
	return r.ResourceManager.List(ctx, list, optionsFunc...)
}

func (r *rbacManager) Update(ctx context.Context, resource model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	key := model.ResourceKey{
		Mesh: resource.GetMeta().GetMesh(),
		Name: resource.GetMeta().GetName(),
	}
	if err := r.resourceAccess.ValidateUpdate(key, resource.GetSpec(), resource.Descriptor(), user.UserFromCtx(ctx)); err != nil {
		return err
	}
	return r.ResourceManager.Update(ctx, resource, optionsFunc...)
}

func (r *rbacManager) Delete(ctx context.Context, resource model.Resource, optionsFunc ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(optionsFunc...)
	key := model.ResourceKey{
		Mesh: opts.Name,
		Name: opts.Name,
	}
	if err := r.resourceAccess.ValidateDelete(key, resource.GetSpec(), resource.Descriptor(), user.UserFromCtx(ctx)); err != nil {
		return err
	}
	return r.ResourceManager.Delete(ctx, resource, optionsFunc...)
}

func (r *rbacManager) DeleteAll(ctx context.Context, list model.ResourceList, optionsFunc ...store.DeleteAllOptionsFunc) error {
	return core_manager.DeleteAllResources(r, ctx, list, optionsFunc...)
}
