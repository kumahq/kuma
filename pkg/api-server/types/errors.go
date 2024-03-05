package types

import (
	"fmt"
	"reflect"
)

type InvalidPageSizeError struct {
	Reason string
}

func (a *InvalidPageSizeError) Error() string {
	return a.Reason
}

func (a *InvalidPageSizeError) Is(err error) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(err)
}

func NewMaxPageSizeExceeded(pageSize, limit int) error {
	return &InvalidPageSizeError{Reason: fmt.Sprintf("invalid page size of %d. Maximum page size is %d", pageSize, limit)}
}

var InvalidPageSize = &InvalidPageSizeError{Reason: "invalid format"}
