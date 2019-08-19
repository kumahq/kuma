package resources

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/pkg/errors"
	"strings"
)

type Resources struct {
	Store store.ResourceStore
}

func (f *Resources) Get(ctx context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	return f.Store.Get(ctx, resource, fs...)
}

func (f *Resources) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	return f.Store.List(ctx, list, fs...)
}

func (f *Resources) Create(ctx context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)
	if resource.GetType() != mesh.MeshType {
		if err := f.ensureMeshExists(ctx, opts.Mesh, opts.Namespace); err != nil {
			return err
		}
	}
	return f.Store.Create(ctx, resource, fs...)
}

func (f *Resources) Delete(ctx context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	return f.Store.Delete(ctx, resource, fs...)
}

func (f *Resources) Update(ctx context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	if resource.GetType() != mesh.MeshType {
		opts := store.NewUpdateOptions(fs...)
		var meshName string
		if opts.Mesh != "" {
			meshName = opts.Mesh
		} else {
			meshName = resource.GetMeta().GetMesh()
		}
		if err := f.ensureMeshExists(ctx, meshName, resource.GetMeta().GetNamespace()); err != nil {
			return err
		}
	}
	return f.Store.Update(ctx, resource, fs...)
}

func (f *Resources) ensureMeshExists(ctx context.Context, meshName string, namespace string) error {
	if err := f.Store.Get(ctx, &mesh.MeshResource{}, store.GetByKey(namespace, meshName, meshName)); err != nil { // todo namespace
		if store.IsResourceNotFound(err) {
			return MeshNotFound(meshName)
		}
		return err
	}
	return nil
}

func MeshNotFound(meshName string) error {
	return errors.Errorf("mesh of name %v is not found", meshName)
}

func IsMeshNotFound(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "mesh of name") && strings.HasSuffix(err.Error(), "is not found")
}
