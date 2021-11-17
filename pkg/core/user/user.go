package user

import "strings"

const AuthenticatedGroup = "mesh-system:authenticated"

type User struct {
	Name   string
	Groups []string
}

func (u User) String() string {
	return u.Name + "/" + strings.Join(u.Groups, ",")
}

func (u User) Authenticated() User {
	u.Groups = append(u.Groups, AuthenticatedGroup)
	return u
}

// Admin is a static user that can be used when authn mechanism does not authenticate to specific user,
// but authenticate to admin without giving credential (ex. authenticate as localhost, authenticate via legacy client certs).
var Admin = User{
	Name:   "mesh-system:admin",
	Groups: []string{"mesh-system:admin"},
}

var Anonymous = User{
	Name:   "mesh-system:anonymous",
	Groups: []string{"mesh-system:unauthenticated"},
}
