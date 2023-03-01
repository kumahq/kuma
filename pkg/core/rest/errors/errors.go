package errors

type Unauthenticated struct{}

func (u *Unauthenticated) Error() string {
	return "Unauthenticated"
}
