package types

import (
	"errors"
	"fmt"
)

type InvalidPageSizeError struct {
	Reason string
}

func (a *InvalidPageSizeError) Error() string {
	return a.Reason
}

func (a *InvalidPageSizeError) Is(err error) bool {
	var target *InvalidPageSizeError
	return errors.As(err, &target)
}

func NewMaxPageSizeExceeded(pageSize, limit int) error {
	return &InvalidPageSizeError{Reason: fmt.Sprintf("invalid page size of %d. Maximum page size is %d", pageSize, limit)}
}

var InvalidPageSize = &InvalidPageSizeError{Reason: "invalid format"}
