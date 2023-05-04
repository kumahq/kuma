package access

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

type ResourceAccess interface {
	ValidateCreate(key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateUpdate(key model.ResourceKey, currentSpec model.ResourceSpec, newSpec model.ResourceSpec, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateDelete(key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateList(mesh string, desc model.ResourceTypeDescriptor, user user.User) error
	ValidateGet(key model.ResourceKey, desc model.ResourceTypeDescriptor, user user.User) error
}
