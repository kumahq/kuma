package rbac

import "github.com/kumahq/kuma/pkg/core/user"

type Role int

const (
	MemberRole Role = iota
	AdminRole
)

func (r Role) String() string {
	switch r {
	case MemberRole:
		return "Member"
	case AdminRole:
		return "Admin"
	default:
		return "unknown"
	}
}

type RoleAssignments interface {
	Role(user user.User) Role
}
