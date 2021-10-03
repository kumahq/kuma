package rbac

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/core/rbac"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

type adminResourceAccess struct {
	roleAssignments rbac.RoleAssignments
}

func NewAdminResourceAccess(roleAssignments rbac.RoleAssignments) ResourceAccess {
	return &adminResourceAccess{
		roleAssignments: roleAssignments,
	}
}

var _ ResourceAccess = &adminResourceAccess{}

func (a *adminResourceAccess) ValidateCreate(key model.ResourceKey, spec model.ResourceSpec, descriptor model.ResourceTypeDescriptor, user *user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateUpdate(key model.ResourceKey, spec model.ResourceSpec, descriptor model.ResourceTypeDescriptor, user *user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateDelete(key model.ResourceKey, spec model.ResourceSpec, descriptor model.ResourceTypeDescriptor, user *user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateList(descriptor model.ResourceTypeDescriptor, user *user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateGet(key model.ResourceKey, descriptor model.ResourceTypeDescriptor, user *user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (r *adminResourceAccess) validateAdminAccess(u *user.User, descriptor model.ResourceTypeDescriptor) error {
	if !descriptor.AdminOnly {
		return nil
	}
	if u == nil {
		return &AccessDeniedError{
			Reason: "user did not authenticate",
		}
	}
	role := r.roleAssignments.Role(*u)
	if role != rbac.AdminRole {
		return &AccessDeniedError{
			Reason: fmt.Sprintf("user %q of role %q cannot access the resource of type %q", u.String(), role.String(), descriptor.Name),
		}
	}
	return nil
}
