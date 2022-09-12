package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-retry"
	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type ReadOnlyResourceManager interface {
	Get(context.Context, model.Resource, ...store.GetOptionsFunc) error
	List(context.Context, model.ResourceList, ...store.ListOptionsFunc) error
}

type ResourceManager interface {
	ReadOnlyResourceManager
	Create(context.Context, model.Resource, ...store.CreateOptionsFunc) error
	Update(context.Context, model.Resource, ...store.UpdateOptionsFunc) error
	Delete(context.Context, model.Resource, ...store.DeleteOptionsFunc) error
	DeleteAll(context.Context, model.ResourceList, ...store.DeleteAllOptionsFunc) error
}

func NewResourceManager(store store.ResourceStore) ResourceManager {
	return &resourcesManager{
		Store: store,
	}
}

var _ ResourceManager = &resourcesManager{}

type resourcesManager struct {
	Store store.ResourceStore
}

func (r *resourcesManager) Get(ctx context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	return r.Store.Get(ctx, resource, fs...)
}

func (r *resourcesManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	return r.Store.List(ctx, list, fs...)
}

func (r *resourcesManager) Create(ctx context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	if err := model.Validate(resource); err != nil {
		return err
	}
	opts := store.NewCreateOptions(fs...)

	var owner model.Resource
	if resource.Descriptor().Scope == model.ScopeMesh {
		owner = core_mesh.NewMeshResource()
		if err := r.Store.Get(ctx, owner, store.GetByKey(opts.Mesh, model.NoMesh)); err != nil {
			return MeshNotFound(opts.Mesh)
		}
	}
	if resource.Descriptor().Name == core_mesh.MeshInsightType {
		owner = core_mesh.NewMeshResource()
		if err := r.Store.Get(ctx, owner, store.GetByKey(opts.Name, model.NoMesh)); err != nil {
			return MeshNotFound(opts.Name)
		}
	}

	return r.Store.Create(ctx, resource, append(fs, store.CreatedAt(core.Now()), store.CreateWithOwner(owner))...)
}

func (r *resourcesManager) Delete(ctx context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	return r.Store.Delete(ctx, resource, fs...)
}

func (r *resourcesManager) DeleteAll(ctx context.Context, list model.ResourceList, fs ...store.DeleteAllOptionsFunc) error {
	return DeleteAllResources(r, ctx, list, fs...)
}

func DeleteAllResources(manager ResourceManager, ctx context.Context, list model.ResourceList, fs ...store.DeleteAllOptionsFunc) error {
	opts := store.NewDeleteAllOptions(fs...)
	if err := manager.List(ctx, list, store.ListByMesh(opts.Mesh)); err != nil {
		return err
	}
	for _, item := range list.GetItems() {
		if err := manager.Delete(ctx, item, store.DeleteBy(model.MetaToResourceKey(item.GetMeta()))); err != nil && !store.IsResourceNotFound(err) {
			return err
		}
	}
	return nil
}

func (r *resourcesManager) Update(ctx context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	if err := model.Validate(resource); err != nil {
		return err
	}
	return r.Store.Update(ctx, resource, append(fs, store.ModifiedAt(time.Now()))...)
}

type ConflictRetry struct {
	BaseBackoff time.Duration
	MaxTimes    uint
}

type UpsertOpts struct {
	ConflictRetry ConflictRetry
}

type UpsertFunc func(opts *UpsertOpts)

func WithConflictRetry(baseBackoff time.Duration, maxTimes uint) UpsertFunc {
	return func(opts *UpsertOpts) {
		opts.ConflictRetry.BaseBackoff = baseBackoff
		opts.ConflictRetry.MaxTimes = maxTimes
	}
}

func NewUpsertOpts(fs ...UpsertFunc) UpsertOpts {
	opts := UpsertOpts{}
	for _, f := range fs {
		f(&opts)
	}
	return opts
}

func Upsert(ctx context.Context, manager ResourceManager, key model.ResourceKey, resource model.Resource, fn func(resource model.Resource) error, fs ...UpsertFunc) error {
	upsert := func(ctx context.Context) error {
		create := false
		err := manager.Get(ctx, resource, store.GetBy(key))
		if err != nil {
			if store.IsResourceNotFound(err) {
				create = true
			} else {
				return err
			}
		}
		if err := fn(resource); err != nil {
			return err
		}
		if create {
			return manager.Create(ctx, resource, store.CreateBy(key))
		} else {
			return manager.Update(ctx, resource)
		}
	}

	opts := NewUpsertOpts(fs...)
	if opts.ConflictRetry.BaseBackoff <= 0 || opts.ConflictRetry.MaxTimes == 0 {
		return upsert(ctx)
	}
	backoff := retry.WithMaxRetries(uint64(opts.ConflictRetry.MaxTimes), retry.NewExponential(opts.ConflictRetry.BaseBackoff))
	return retry.Do(ctx, backoff, func(ctx context.Context) error {
		resource.SetMeta(nil)
		proto.Reset(resource.GetSpec())
		err := upsert(ctx)
		if store.IsResourceConflict(err) {
			return retry.RetryableError(err)
		}
		return err
	})
}

type MeshNotFoundError struct {
	Mesh string
}

func (m *MeshNotFoundError) Error() string {
	return fmt.Sprintf("mesh of name %s is not found", m.Mesh)
}

func MeshNotFound(meshName string) error {
	return &MeshNotFoundError{meshName}
}

func IsMeshNotFound(err error) bool {
	_, ok := err.(*MeshNotFoundError)
	return ok
}
