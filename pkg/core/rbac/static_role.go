package rbac

import (
	"github.com/kumahq/kuma/pkg/config/rbac"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticRoleAssignments struct {
	adminUsers  map[string]bool
	adminGroups map[string]bool
}

var _ RoleAssignments = &staticRoleAssignments{}

// NewStaticRoleAssignments returns RoleAssignments that assigns a role for a user based on the static Kuma CP configuration
func NewStaticRoleAssignments(cfg rbac.RBACStaticConfig) RoleAssignments {
	s := staticRoleAssignments{
		adminUsers:  map[string]bool{},
		adminGroups: map[string]bool{},
	}
	for _, user := range cfg.AdminUsers {
		s.adminUsers[user] = true
	}
	for _, group := range cfg.AdminGroups {
		s.adminGroups[group] = true
	}
	return &s
}

func (s *staticRoleAssignments) Role(user user.User) Role {
	if s.adminUsers[user.Name] || s.adminGroups[user.Group] {
		return AdminRole
	}
	return MemberRole
}
