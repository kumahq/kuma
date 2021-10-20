package rbac

import (
	"fmt"

	config_rbac "github.com/kumahq/kuma/pkg/config/rbac"
	"github.com/kumahq/kuma/pkg/core/rbac"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

type adminResourceAccess struct {
	usernames map[string]bool
	groups    map[string]bool
}

func NewAdminResourceAccess(cfg config_rbac.AdminResourcesRBACStaticConfig) ResourceAccess {
	a := &adminResourceAccess{
		usernames: map[string]bool{},
		groups:    map[string]bool{},
	}
	for _, user := range cfg.Users {
		a.usernames[user] = true
	}
	for _, group := range cfg.Groups {
		a.groups[group] = true
	}
	return a
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
		return &rbac.AccessDeniedError{
			Reason: "user did not authenticate",
		}
	}
	allowed := r.usernames[u.Name]
	for _, group := range u.Groups {
		if r.groups[group] {
			allowed = true
		}
	}
	if !allowed {
		return &rbac.AccessDeniedError{
			Reason: fmt.Sprintf("user %q cannot access the resource of type %q", u.String(), descriptor.Name),
		}
	}
	return nil
}
