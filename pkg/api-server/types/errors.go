package types

import (
	"strings"

	"github.com/pkg/errors"
)

func NewMaxPageSizeExceeded(pageSize, limit int) error {
	return errors.Errorf("Invalid page size of %d. Maximum page size is %d", pageSize, limit)
}

func IsMaxPageSizeExceeded(err error) bool {
	return strings.HasPrefix(err.Error(), "Invalid page size of")
}

var InvalidPageSize = errors.New("Invalid page size")
