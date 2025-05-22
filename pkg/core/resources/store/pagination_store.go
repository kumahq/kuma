package store

import (
	"context"
	"sort"
	"strconv"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

// The Pagination Store is handling only the pagination functionality in the List.
// This is an in-memory operation and offloads this from the persistent stores (k8s, postgres etc.)
// Two reasons why this is needed:
// * There is no filtering + pagination on the native K8S database
// * On Postgres, we keep the object in a column as a string. We would have to use JSON column type and convert it to native SQL queries.
//
// The in-memory filtering has been tested with 10,000 Dataplanes and proved to be fast enough, although not that efficient.
func NewPaginationStore(delegate ResourceStore) ResourceStore {
	return &paginationStore{
		delegate: delegate,
	}
}

type paginationStore struct {
	delegate ResourceStore
}

func (p *paginationStore) Create(ctx context.Context, resource model.Resource, optionsFunc ...CreateOptionsFunc) error {
	return p.delegate.Create(ctx, resource, optionsFunc...)
}

func (p *paginationStore) Update(ctx context.Context, resource model.Resource, optionsFunc ...UpdateOptionsFunc) error {
	return p.delegate.Update(ctx, resource, optionsFunc...)
}

func (p *paginationStore) Delete(ctx context.Context, resource model.Resource, optionsFunc ...DeleteOptionsFunc) error {
	return p.delegate.Delete(ctx, resource, optionsFunc...)
}

func (p *paginationStore) Get(ctx context.Context, resource model.Resource, optionsFunc ...GetOptionsFunc) error {
	return p.delegate.Get(ctx, resource, optionsFunc...)
}

func (p *paginationStore) List(ctx context.Context, list model.ResourceList, optionsFunc ...ListOptionsFunc) error {
	opts := NewListOptions(optionsFunc...)

	// At least one of the following options is required to trigger the paginationStore to do work.
	// Otherwise, it delegates the request and returns early.
	if opts.FilterFunc == nil && opts.PageSize == 0 && opts.PageOffset == "" && !opts.Ordered && len(opts.ResourceKeys) == 0 {
		return p.delegate.List(ctx, list, optionsFunc...)
	}

	fullList, err := registry.Global().NewList(list.GetItemType())
	if err != nil {
		return err
	}

	err = p.delegate.List(ctx, fullList, optionsFunc...)
	if err != nil {
		return err
	}

	filteredList, err := registry.Global().NewList(list.GetItemType())
	if err != nil {
		return err
	}

	for _, item := range fullList.GetItems() {
		_, exists := opts.ResourceKeys[model.MetaToResourceKey(item.GetMeta())]
		if len(opts.ResourceKeys) > 0 && !exists {
			continue
		}
		if !opts.Filter(item) {
			continue
		}
		_ = filteredList.AddItem(item)
	}

	filteredItems := filteredList.GetItems()
	lenFilteredItems := len(filteredItems)
	sort.Sort(model.ByMeta(filteredItems))

	offset := 0
	pageSize := lenFilteredItems
	paginationEnabled := opts.PageSize != 0
	if paginationEnabled {
		pageSize = opts.PageSize
		if opts.PageOffset != "" {
			o, err := strconv.Atoi(opts.PageOffset)
			if err != nil {
				return ErrInvalidOffset
			}
			offset = o
		}
	}

	for i := offset; i < offset+pageSize && i < lenFilteredItems; i++ {
		_ = list.AddItem(filteredItems[i])
	}

	if paginationEnabled {
		nextOffset := ""
		if offset+pageSize < lenFilteredItems { // set new offset only if we did not reach the end of the collection
			nextOffset = strconv.Itoa(offset + opts.PageSize)
		}
		list.GetPagination().SetNextOffset(nextOffset)
	}

	list.GetPagination().SetTotal(uint32(lenFilteredItems))

	return nil
}

var _ ResourceStore = &paginationStore{}
