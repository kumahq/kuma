package manager

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"strings"
)

type ResourceManager interface {
	Create(context.Context, model.Resource, ...store.CreateOptionsFunc) error
	Update(context.Context, model.Resource, ...store.UpdateOptionsFunc) error
	Delete(context.Context, model.Resource, ...store.DeleteOptionsFunc) error
	DeleteAll(context.Context, ...store.DeleteAllOptionsFunc) error
	Get(context.Context, model.Resource, ...store.GetOptionsFunc) error
	List(context.Context, model.ResourceList, ...store.ListOptionsFunc) error
}

func NewResourceManager(store store.ResourceStore, typeRegistry registry.TypeRegistry) ResourceManager {
	return &resourcesManager{
		Store:    store,
		Registry: typeRegistry,
	}
}

var _ ResourceManager = &resourcesManager{}

type resourcesManager struct {
	Store    store.ResourceStore
	Registry registry.TypeRegistry
}

func (r *resourcesManager) Get(ctx context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	return r.Store.Get(ctx, resource, fs...)
}

func (r *resourcesManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	return r.Store.List(ctx, list, fs...)
}

func (r *resourcesManager) Create(ctx context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)
	if resource.GetType() != mesh.MeshType {
		if err := r.ensureMeshExists(ctx, opts.Mesh, opts.Namespace); err != nil {
			return err
		}
	}
	return r.Store.Create(ctx, resource, fs...)
}

func (r *resourcesManager) ensureMeshExists(ctx context.Context, meshName string, namespace string) error {
	list := mesh.MeshResourceList{}
	if err := r.Store.List(ctx, &list, store.ListByMesh(meshName)); err != nil {
		return err
	}
	if len(list.Items) != 1 {
		return MeshNotFound(meshName)
	}
	return nil
}

func (r *resourcesManager) Delete(ctx context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	return r.Store.Delete(ctx, resource, fs...)
}

func (r *resourcesManager) DeleteAll(ctx context.Context, fs ...store.DeleteAllOptionsFunc) error {
	opts := store.NewDeleteAllOptions(fs...)

	for _, typ := range r.Registry.ListTypes() {
		list, err := r.Registry.NewList(typ)
		if err != nil {
			return err
		}
		if err := r.List(ctx, list, store.ListByMesh(opts.Mesh)); err != nil {
			return err
		}
		for _, obj := range list.GetItems() {
			if err := r.Delete(ctx, obj, store.DeleteBy(model.MetaToResourceKey(obj.GetMeta()))); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *resourcesManager) Update(ctx context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	return r.Store.Update(ctx, resource, fs...)
}

func MeshNotFound(meshName string) error {
	return errors.Errorf("mesh of name %v is not found", meshName)
}

func IsMeshNotFound(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "mesh of name") && strings.HasSuffix(err.Error(), "is not found")
}
