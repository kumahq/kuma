package access

import (
	"fmt"
	"reflect"

	"github.com/kumahq/kuma/pkg/core/user"
)

type AccessDeniedError struct {
	Reason string
}

func (a *AccessDeniedError) Error() string {
	return "access denied: " + a.Reason
}

func (a *AccessDeniedError) Is(err error) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(err)
}

func Validate(usernames, groups map[string]struct{}, user user.User, action string) error {
	if _, ok := usernames[user.Name]; ok {
		return nil
	}
	for _, group := range user.Groups {
		if _, ok := groups[group]; ok {
			return nil
		}
	}
	return &AccessDeniedError{Reason: fmt.Sprintf("user %q cannot access %s", user, action)}
}
