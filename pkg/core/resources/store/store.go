package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceStore interface {
	Create(context.Context, model.Resource, ...CreateOptionsFunc) error
	Update(context.Context, model.Resource, ...UpdateOptionsFunc) error
	Delete(context.Context, model.Resource, ...DeleteOptionsFunc) error
	Get(context.Context, model.Resource, ...GetOptionsFunc) error
	List(context.Context, model.ResourceList, ...ListOptionsFunc) error
}

type ClosableResourceStore interface {
	ResourceStore
	io.Closer
}

func NewStrictResourceStore(c ResourceStore) ClosableResourceStore {
	return &strictResourceStore{delegate: c}
}

var _ ResourceStore = &strictResourceStore{}

// strictResourceStore encapsulates a contract between ResourceStore and its users.
type strictResourceStore struct {
	delegate ResourceStore
}

func (s *strictResourceStore) Create(ctx context.Context, r model.Resource, fs ...CreateOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceStore.Create() requires a non-nil resource")
	}
	if r.GetMeta() != nil {
		return fmt.Errorf("ResourceStore.Create() ignores resource.GetMeta() but the argument has a non-nil value")
	}
	opts := NewCreateOptions(fs...)
	if opts.Name == "" {
		return fmt.Errorf("ResourceStore.Create() requires options.Name to be a non-empty value")
	}
	if r.Descriptor().Scope == model.ScopeMesh && opts.Mesh == "" {
		return fmt.Errorf("ResourceStore.Create() requires options.Mesh to be a non-empty value")
	}
	return s.delegate.Create(ctx, r, fs...)
}

func (s *strictResourceStore) Update(ctx context.Context, r model.Resource, fs ...UpdateOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceStore.Update() requires a non-nil resource")
	}
	if r.GetMeta() == nil {
		return fmt.Errorf("ResourceStore.Update() requires resource.GetMeta() to be a non-nil value previously returned by ResourceStore.Get()")
	}
	return s.delegate.Update(ctx, r, fs...)
}

func (s *strictResourceStore) Delete(ctx context.Context, r model.Resource, fs ...DeleteOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceStore.Delete() requires a non-nil resource")
	}
	opts := NewDeleteOptions(fs...)
	if opts.Name == "" {
		return fmt.Errorf("ResourceStore.Delete() requires options.Name to be a non-empty value")
	}
	if r.Descriptor().Scope == model.ScopeMesh && opts.Mesh == "" {
		return fmt.Errorf("ResourceStore.Delete() requires options.Mesh to be a non-empty value")
	}
	if r.GetMeta() != nil {
		if opts.Name != r.GetMeta().GetName() {
			return fmt.Errorf("ResourceStore.Delete() requires resource.GetMeta() either to be a nil or resource.GetMeta().GetName() == options.Name")
		}
		if opts.Mesh != r.GetMeta().GetMesh() {
			return fmt.Errorf("ResourceStore.Delete() requires resource.GetMeta() either to be a nil or resource.GetMeta().GetMesh() == options.Mesh")
		}
	}
	return s.delegate.Delete(ctx, r, fs...)
}

func (s *strictResourceStore) Get(ctx context.Context, r model.Resource, fs ...GetOptionsFunc) error {
	if r == nil {
		return fmt.Errorf("ResourceStore.Get() requires a non-nil resource")
	}
	if r.GetMeta() != nil {
		return fmt.Errorf("ResourceStore.Get() ignores resource.GetMeta() but the argument has a non-nil value")
	}
	opts := NewGetOptions(fs...)
	if opts.Name == "" {
		return fmt.Errorf("ResourceStore.Get() requires options.Name to be a non-empty value")
	}
	if r.Descriptor().Scope == model.ScopeMesh && opts.Mesh == "" {
		return fmt.Errorf("ResourceStore.Get() requires options.Mesh to be a non-empty value")
	}
	return s.delegate.Get(ctx, r, fs...)
}

func (s *strictResourceStore) List(ctx context.Context, rs model.ResourceList, fs ...ListOptionsFunc) error {
	if rs == nil {
		return fmt.Errorf("ResourceStore.List() requires a non-nil resource list")
	}
	return s.delegate.List(ctx, rs, fs...)
}

func (s *strictResourceStore) Close() error {
	closable, ok := s.delegate.(io.Closer)
	if ok {
		return closable.Close()
	}
	return nil
}

type ResourceConflictError struct {
	rType model.ResourceType
	name  string
	mesh  string
	msg   string
}

func (e *ResourceConflictError) Error() string {
	return fmt.Sprintf("%s: type=%q name=%q mesh=%q", e.msg, e.rType, e.name, e.mesh)
}

func (e *ResourceConflictError) Is(err error) bool {
	return reflect.TypeOf(e) == reflect.TypeOf(err)
}

func ErrorResourceAlreadyExists(rt model.ResourceType, name, mesh string) error {
	return &ResourceConflictError{msg: "resource already exists", rType: rt, name: name, mesh: mesh}
}

func ErrorResourceConflict(rt model.ResourceType, name, mesh string) error {
	return &ResourceConflictError{msg: "resource conflict", rType: rt, name: name, mesh: mesh}
}

func ErrorResourceNotFound(rt model.ResourceType, name, mesh string) error {
	return fmt.Errorf("resource not found: type=%q name=%q mesh=%q", rt, name, mesh)
}

var ErrorInvalidOffset = errors.New("invalid offset")

func IsResourceNotFound(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "resource not found")
}

func IsResourceAlreadyExists(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "resource already exists")
}

// AssertionError
type AssertionError struct {
	msg string
	err error
}

func ErrorResourceAssertion(msg string, rt model.ResourceType, name, mesh string) error {
	return &AssertionError{
		msg: fmt.Sprintf("%s: type=%q name=%q mesh=%q", msg, rt, name, mesh),
	}
}

func (e *AssertionError) Unwrap() error {
	return e.err
}

func (e *AssertionError) Error() string {
	msg := "store assertion failed"
	if e.msg != "" {
		msg += " " + e.msg
	}
	if e.err != nil {
		msg += fmt.Sprintf("error: %s", e.err)
	}
	return msg
}

func (e *AssertionError) Is(err error) bool {
	return reflect.TypeOf(e) == reflect.TypeOf(err)
}

type PreconditionError struct {
	Reason string
}

func (a *PreconditionError) Error() string {
	return a.Reason
}

func (a *PreconditionError) Is(err error) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(err)
}

func PreconditionFormatError(reason string) *PreconditionError {
	return &PreconditionError{Reason: "invalid format: " + reason}
}
