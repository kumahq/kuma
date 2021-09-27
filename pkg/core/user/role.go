package user

type Role int

const (
	UserRole Role = iota
	AdminRole
)

func (r Role) String() string {
	switch r {
	case UserRole:
		return "User"
	case AdminRole:
		return "Admin"
	default:
		return "unknown"
	}
}

type RoleAssignments interface {
	Role(user User) Role
}
