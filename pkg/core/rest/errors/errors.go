package errors

type Unauthenticated struct {
}

func (u *Unauthenticated) Error() string {
	return "Unauthenticated"
}

type AccessDenied struct {
}

func (u *AccessDenied) Error() string {
	return "Access Denied"
}
