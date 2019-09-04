package store

import (
	"context"

	secret_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/system"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

func NewSecretStore(resourceStore core_store.ResourceStore) SecretStore {
	return &secretStore{
		resourceStore: resourceStore,
	}
}

var _ SecretStore = &secretStore{}

type secretStore struct {
	resourceStore core_store.ResourceStore
}

func (r *secretStore) Get(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.GetOptionsFunc) error {
	return r.resourceStore.Get(ctx, secret, fs...)
}

func (r *secretStore) List(ctx context.Context, secrets *secret_model.SecretResourceList, fs ...core_store.ListOptionsFunc) error {
	return r.resourceStore.List(ctx, secrets, fs...)
}

func (r *secretStore) Create(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.CreateOptionsFunc) error {
	return r.resourceStore.Create(ctx, secret, fs...)
}

func (r *secretStore) Delete(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.DeleteOptionsFunc) error {
	return r.resourceStore.Delete(ctx, secret, fs...)
}

func (r *secretStore) Update(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.UpdateOptionsFunc) error {
	return r.resourceStore.Update(ctx, secret, fs...)
}
