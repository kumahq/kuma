package store

import (
	"errors"
	"fmt"

	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

var (
	ErrIsAlreadyExists = errors.New("already exists")
	ErrConflict        = errors.New("conflict")
	ErrNotFound        = errors.New("not found")
	// ErrInvalid is the store's verdict on one specific resource: the resource is
	// unacceptable on its own merits, so replaying it unchanged can never succeed.
	// It is the opposite of a transient failure - a database outage, a lost
	// connection, a timeout - which the caller should retry.
	//
	// Callers that apply many resources at once (KDS sync) skip an invalid resource
	// and carry on with the rest, so a store must only return it when dropping the
	// resource is the correct outcome. A failure that says nothing about the
	// resource has to stay a plain error, so it fails the whole operation.
	ErrInvalid = errors.New("invalid")
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
