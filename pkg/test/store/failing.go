package store

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type FailingStore struct {
	Err       error
	CreateErr error
}

var _ core_store.ResourceStore = &FailingStore{}

func (f *FailingStore) Create(context.Context, model.Resource, ...core_store.CreateOptionsFunc) error {
	if f.CreateErr != nil {
		return f.CreateErr
	}
	return f.Err
}

func (f *FailingStore) Update(context.Context, model.Resource, ...core_store.UpdateOptionsFunc) error {
	return f.Err
}

func (f *FailingStore) Delete(context.Context, model.Resource, ...core_store.DeleteOptionsFunc) error {
	return f.Err
}

func (f *FailingStore) Get(context.Context, model.Resource, ...core_store.GetOptionsFunc) error {
	return f.Err
}

func (f *FailingStore) List(context.Context, model.ResourceList, ...core_store.ListOptionsFunc) error {
	return f.Err
}
