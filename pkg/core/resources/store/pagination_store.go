package store

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func NewPaginationStore(delegate ResourceStore) ResourceStore {
	return &PaginationStore{
		delegate: delegate,
	}
}

type PaginationStore struct {
	delegate ResourceStore
}

func (m *PaginationStore) Create(ctx context.Context, resource model.Resource, optionsFunc ...CreateOptionsFunc) error {
	return m.delegate.Create(ctx, resource, optionsFunc...)
}

func (m *PaginationStore) Update(ctx context.Context, resource model.Resource, optionsFunc ...UpdateOptionsFunc) error {
	return m.delegate.Update(ctx, resource, optionsFunc...)
}

func (m *PaginationStore) Delete(ctx context.Context, resource model.Resource, optionsFunc ...DeleteOptionsFunc) error {
	return m.delegate.Delete(ctx, resource, optionsFunc...)
}

func (m *PaginationStore) Get(ctx context.Context, resource model.Resource, optionsFunc ...GetOptionsFunc) error {
	return m.delegate.Get(ctx, resource, optionsFunc...)
}

func (m *PaginationStore) List(ctx context.Context, list model.ResourceList, optionsFunc ...ListOptionsFunc) error {
	return m.delegate.List(ctx, list, optionsFunc...)
}

var _ ResourceStore = &PaginationStore{}
