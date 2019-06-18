package client

import (
	"context"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model"
)

type ResourceClient interface {
	Create(context.Context, model.Resource, ...CreateOptionsFunc) error
	Update(context.Context, model.Resource, ...UpdateOptionsFunc) error
	Delete(context.Context, model.Resource, ...DeleteOptionsFunc) error
	Get(context.Context, model.Resource, ...GetOptionsFunc) error
	List(context.Context, model.ResourceList, ...ListOptionsFunc) error
}
