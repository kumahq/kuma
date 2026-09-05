package util

import (
	"errors"
	"strings"
)

var ErrUserNack = errors.New("user error")

func IsUserErrorMessage(message string) bool {
	return strings.HasPrefix(message, ErrUserNack.Error())
}

func IsUserError(err error) bool {
	return errors.Is(err, ErrUserNack)
}
