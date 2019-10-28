package validators

import (
	"github.com/pkg/errors"
	"strings"
)

func NewValidationError(err error) error {
	return errors.Wrap(err, "validation error")
}

func IsValidationError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "validation error")
}
