package errors

type Unauthenticated struct{}

func (u *Unauthenticated) Error() string {
	return "Unauthenticated"
}

type MethodNotAllowed struct{}

func (m *MethodNotAllowed) Error() string {
	return "Method not allowed"
}
