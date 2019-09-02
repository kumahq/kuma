package store

import (
	"context"

	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	secret_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/model"
)

type SecretStore interface {
	Create(context.Context, *secret_model.Secret, ...core_store.CreateOptionsFunc) error
	Update(context.Context, *secret_model.Secret, ...core_store.UpdateOptionsFunc) error
	Delete(context.Context, *secret_model.Secret, ...core_store.DeleteOptionsFunc) error
	Get(context.Context, *secret_model.Secret, ...core_store.GetOptionsFunc) error
	List(context.Context, *secret_model.SecretList, ...core_store.ListOptionsFunc) error
}
