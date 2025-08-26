package access

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

type ResourceAccess interface {
	ValidateCreate(ctx context.Context, key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateUpdate(ctx context.Context, key model.ResourceKey, currentSpec, newSpec model.ResourceSpec, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateDelete(ctx context.Context, key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateList(ctx context.Context, mesh string, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateGet(ctx context.Context, key model.ResourceKey, desc model.ResourceTypeDescriptor, user user.User) error
}
