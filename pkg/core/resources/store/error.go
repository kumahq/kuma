package store

import (
	"errors"
	"fmt"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var (
	ErrIsAlreadyExists = errors.New("already exists")
	ErrConflict        = errors.New("conflict")
	ErrNotFound        = errors.New("not found")
	ErrInvalid         = errors.New("invalid")
)

func ErrorResourceAlreadyExists(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrIsAlreadyExists, rt, name, mesh)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrIsAlreadyExists)
}

func ErrorResourceConflict(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrConflict, rt, name, mesh)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func ErrorResourceNotFound(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrNotFound, rt, name, mesh)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func ErrorInvalid(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalid, reason)
}

func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}
