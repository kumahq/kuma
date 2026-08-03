package util

import (
	"errors"
	"strings"
)

var (
	ErrUserNack     = errors.New("user error")
	ErrInternalNack = errors.New("internal error")
)

func IsUserErrorMessage(message string) bool {
	return strings.HasPrefix(message, ErrUserNack.Error())
}

func IsInternalErrorMessage(message string) bool {
	return strings.HasPrefix(message, ErrInternalNack.Error())
}

func IsUserError(err error) bool {
	return errors.Is(err, ErrUserNack)
}

func IsInternalError(err error) bool {
	return errors.Is(err, ErrInternalNack)
}
