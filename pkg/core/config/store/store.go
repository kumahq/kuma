package store

import (
	"context"

	config_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

type ConfigStore interface {
	Create(context.Context, *config_model.ConfigResource, ...core_store.CreateOptionsFunc) error
	Update(context.Context, *config_model.ConfigResource, ...core_store.UpdateOptionsFunc) error
	Delete(context.Context, *config_model.ConfigResource, ...core_store.DeleteOptionsFunc) error
	Get(context.Context, *config_model.ConfigResource, ...core_store.GetOptionsFunc) error
	List(context.Context, *config_model.ConfigResourceList, ...core_store.ListOptionsFunc) error
}
