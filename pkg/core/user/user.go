package user

import "strings"

type User struct {
	Name   string
	Groups []string
}

func (u User) String() string {
	return u.Name + "/" + strings.Join(u.Groups, ",")
}

// Admin is a static user that can be used when authn mechanism does not authenticate to specific user,
// but authenticate to admin without giving credential (ex. authenticate as localhost, authenticate via legacy client certs).
var Admin = User{
	Name:   "admin",
	Groups: []string{"admin"},
}
