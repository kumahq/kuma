package access

import "reflect"

type AccessDeniedError struct {
	Reason string
}

func (a *AccessDeniedError) Error() string {
	return "access denied: " + a.Reason
}

func (a *AccessDeniedError) Is(err error) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(err)
}
