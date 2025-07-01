package store

import (
	"context"
	"fmt"
	"io"

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
