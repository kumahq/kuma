package store

import (
	"context"
	"strconv"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func NewPaginationStore(delegate ResourceStore) ResourceStore {
	return &PaginationStore{
		delegate: delegate,
	}
}

type PaginationStore struct {
	delegate ResourceStore
}

func (p *PaginationStore) Create(ctx context.Context, resource model.Resource, optionsFunc ...CreateOptionsFunc) error {
	return p.delegate.Create(ctx, resource, optionsFunc...)
}

func (p *PaginationStore) Update(ctx context.Context, resource model.Resource, optionsFunc ...UpdateOptionsFunc) error {
	return p.delegate.Update(ctx, resource, optionsFunc...)
}

func (p *PaginationStore) Delete(ctx context.Context, resource model.Resource, optionsFunc ...DeleteOptionsFunc) error {
	return p.delegate.Delete(ctx, resource, optionsFunc...)
}

func (p *PaginationStore) Get(ctx context.Context, resource model.Resource, optionsFunc ...GetOptionsFunc) error {
	return p.delegate.Get(ctx, resource, optionsFunc...)
}

func (p *PaginationStore) List(ctx context.Context, list model.ResourceList, optionsFunc ...ListOptionsFunc) error {
	fullList, err := registry.Global().NewList(list.GetItemType())
	if err != nil {
		return err
	}

	err = p.delegate.List(ctx, fullList, optionsFunc...)
	if err != nil {
		return err
	}

	allItems := fullList.GetItems()
	lenAllItems := len(allItems)

	opts := NewListOptions(optionsFunc...)

	offset := 0
	pageSize := lenAllItems
	paginateResults := opts.PageSize != 0
	if paginateResults {
		pageSize = opts.PageSize
		if opts.PageOffset != "" {
			o, err := strconv.Atoi(opts.PageOffset)
			if err != nil {
				return ErrorInvalidOffset
			}
			offset = o
		}
	}

	for i := offset; i < offset+pageSize && i < lenAllItems; i++ {
		for i < lenAllItems && !opts.Filter(allItems[i]) {
			i++
		}
		if i == lenAllItems {
			// no filtered items to add
			break
		}
		_ = list.AddItem(allItems[i])
	}

	if paginateResults {
		nextOffset := ""
		if offset+pageSize < lenAllItems { // set new offset only if we did not reach the end of the collection
			nextOffset = strconv.Itoa(offset + opts.PageSize)
		}
		list.GetPagination().SetNextOffset(nextOffset)
	}

	list.GetPagination().SetTotal(uint32(lenAllItems))

	return nil
}

var _ ResourceStore = &PaginationStore{}
