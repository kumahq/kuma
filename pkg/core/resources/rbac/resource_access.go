package rbac

import (
	"reflect"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

type AccessDeniedError struct {
	Reason string
}

func (a *AccessDeniedError) Error() string {
	return "access denied: " + a.Reason
}

func (a *AccessDeniedError) Is(err error) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(err)
}

type ResourceAccess interface {
	ValidateCreate(key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user *user.User) error
	ValidateUpdate(key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user *user.User) error
	ValidateDelete(key model.ResourceKey, spec model.ResourceSpec, desc model.ResourceTypeDescriptor, user *user.User) error
	ValidateList(desc model.ResourceTypeDescriptor, user *user.User) error
	ValidateGet(key model.ResourceKey, desc model.ResourceTypeDescriptor, user *user.User) error
}
