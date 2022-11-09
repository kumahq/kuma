package access

import (
	"fmt"

	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

type adminResourceAccess struct {
	usernames map[string]bool
	groups    map[string]bool
}

func NewAdminResourceAccess(cfg config_access.AdminResourcesStaticAccessConfig) ResourceAccess {
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

func (a *adminResourceAccess) ValidateCreate(_ model.ResourceKey, _ model.ResourceSpec, descriptor model.ResourceTypeDescriptor, user user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateUpdate(_ model.ResourceKey, _ model.ResourceSpec, _ model.ResourceSpec, descriptor model.ResourceTypeDescriptor, user user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateDelete(_ model.ResourceKey, _ model.ResourceSpec, descriptor model.ResourceTypeDescriptor, user user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateList(_ string, descriptor model.ResourceTypeDescriptor, user user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (a *adminResourceAccess) ValidateGet(_ model.ResourceKey, descriptor model.ResourceTypeDescriptor, user user.User) error {
	return a.validateAdminAccess(user, descriptor)
}

func (r *adminResourceAccess) validateAdminAccess(u user.User, descriptor model.ResourceTypeDescriptor) error {
	if !descriptor.AdminOnly {
		return nil
	}
	allowed := r.usernames[u.Name]
	for _, group := range u.Groups {
		if r.groups[group] {
			allowed = true
		}
	}
	if !allowed {
		return &access.AccessDeniedError{
			Reason: fmt.Sprintf("user %q cannot access the resource of type %q", u.String(), descriptor.Name),
		}
	}
	return nil
}
