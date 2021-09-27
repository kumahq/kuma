package user

type User struct {
	Name  string
	Group string
}

func (u User) String() string {
	return u.Name + "/" + u.Group
}

// Admin is a static user that can be used when authn mechanism does not authenticate to specific user,
// but authenticate to admin without giving credential (ex. authenticate as localhost, authenticate via legacy client certs).
var Admin = User{
	Name:  "admin",
	Group: "admin",
}
