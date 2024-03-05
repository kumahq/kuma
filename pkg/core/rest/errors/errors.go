package errors

import (
	"fmt"
	"reflect"
)

type Unauthenticated struct{}

func (e *Unauthenticated) Error() string {
	return "Unauthenticated"
}

type MethodNotAllowed struct{}

func (e *MethodNotAllowed) Error() string {
	return "Method not allowed"
}

type Conflict struct{}

func (e *Conflict) Error() string {
	return "Conflict"
}

type ServiceUnavailable struct{}

func (e *ServiceUnavailable) Error() string {
	return "Service unavailable"
}

type BadRequest struct {
	msg string
}

func NewBadRequestError(msg string) error {
	return &BadRequest{msg: msg}
}

func (e *BadRequest) Error() string {
	if e.msg == "" {
		return "bad request"
	}
	return fmt.Sprintf("bad request: %s", e.msg)
}

func (e *BadRequest) Is(err error) bool {
	return reflect.TypeOf(e) == reflect.TypeOf(err)
}
