package dataplane

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type Validator interface {
	ValidateCreate(ctx context.Context, key model.ResourceKey, newDp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error
	ValidateUpdate(ctx context.Context, newDp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error
}
