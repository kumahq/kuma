package store

import (
	"errors"
	"fmt"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var (
	ErrorInvalidOffset = errors.New("invalid offset")
	ErrIsAlreadyExists = errors.New("already exists")
	ErrConflict        = errors.New("conflict")
	ErrNotFound        = errors.New("not found")
	ErrInvalid         = errors.New("invalid")
	ErrBadRequest      = errors.New("bad request")
)

func ErrorResourceAlreadyExists(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrIsAlreadyExists, rt, name, mesh)
}

func IsResourceAlreadyExists(err error) bool {
	return err != nil && errors.Is(err, ErrIsAlreadyExists)
}

func ErrorResourceConflict(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrConflict, rt, name, mesh)
}

func IsResourceConflict(err error) bool {
	return err != nil && errors.Is(err, ErrConflict)
}

func ErrorResourceNotFound(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrNotFound, rt, name, mesh)
}

func IsResourceNotFound(err error) bool {
	return err != nil && errors.Is(err, ErrNotFound)
}

func ErrorResourceInvalid(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrInvalid, rt, name, mesh)
}

func IsResourceInvalid(err error) bool {
	return err != nil && errors.Is(err, ErrInvalid)
}

func ErrorBadRequest(msg string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, msg)
}

func IsBadRequest(err error) bool {
	return err != nil && errors.Is(err, ErrBadRequest)
}
