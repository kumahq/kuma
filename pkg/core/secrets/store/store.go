package store

import (
	"context"

	secret_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

type SecretStore interface {
	Create(context.Context, *secret_model.SecretResource, ...core_store.CreateOptionsFunc) error
	Update(context.Context, *secret_model.SecretResource, ...core_store.UpdateOptionsFunc) error
	Delete(context.Context, *secret_model.SecretResource, ...core_store.DeleteOptionsFunc) error
	Get(context.Context, *secret_model.SecretResource, ...core_store.GetOptionsFunc) error
	List(context.Context, *secret_model.SecretResourceList, ...core_store.ListOptionsFunc) error
}
