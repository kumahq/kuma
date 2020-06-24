package store

import (
	"context"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

// SyncResourceStore extends ResourceStore with Sync method
type SyncResourceStore interface {
	store.ResourceStore
	// Sync method takes 'upstream' as a basis and synchronize underlying store.
	// It deletes all resources that absent in 'upstream', creates new resources that
	// are not represented in store yet and updates the rest.
	Sync(upstream model.ResourceList, fs ...SyncOptionFunc) error
}

type SyncOption struct {
	Predicate func(r model.Resource) bool
}

type SyncOptionFunc func(*SyncOption)

func NewSyncOptions(fs ...SyncOptionFunc) *SyncOption {
	opts := &SyncOption{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func PrefilterBy(predicate func(r model.Resource) bool) SyncOptionFunc {
	return func(opts *SyncOption) {
		opts.Predicate = predicate
	}
}

type syncResourceStore struct {
	store.ResourceStore
}

func NewSyncResourceStore(resourceStore store.ResourceStore) SyncResourceStore {
	return &syncResourceStore{
		ResourceStore: resourceStore,
	}
}

func (s *syncResourceStore) Sync(upstream model.ResourceList, fs ...SyncOptionFunc) error {
	opts := NewSyncOptions(fs...)
	ctx := context.Background()

	downstream, err := registry.Global().NewList(upstream.GetItemType())
	if err != nil {
		return err
	}
	if err := s.ResourceStore.List(ctx, downstream); err != nil {
		return err
	}

	if opts.Predicate != nil {
		filtered, err := registry.Global().NewList(upstream.GetItemType())
		if err != nil {
			return err
		}
		for _, r := range downstream.GetItems() {
			if opts.Predicate(r) {
				if err := filtered.AddItem(r); err != nil {
					return err
				}
			}
		}
		downstream = filtered
	}

	// 1. delete resources from store which are not represented in 'upstream'
	onDelete := []model.Resource{}
	for _, r := range downstream.GetItems() {
		if !set(upstream.GetItems()).contains(model.MetaToResourceKey(r.GetMeta())) {
			onDelete = append(onDelete, r)
		}
	}

	// 2. create resources which are not represented in 'downstream' and update the rest of them
	onCreate := []model.Resource{}
	onUpdate := []model.Resource{}
	for _, r := range upstream.GetItems() {
		if existing := set(downstream.GetItems()).get(model.MetaToResourceKey(r.GetMeta())); existing != nil {
			// we have to use meta of the current store during update
			r.SetMeta(existing.GetMeta())
			onUpdate = append(onUpdate, r)
		} else {
			onCreate = append(onCreate, r)
		}
	}

	for _, r := range onDelete {
		if err := s.ResourceStore.Delete(ctx, r, store.DeleteBy(model.MetaToResourceKey(r.GetMeta()))); err != nil {
			return err
		}
	}

	for _, r := range onCreate {
		rk := model.MetaToResourceKey(r.GetMeta())
		r.SetMeta(nil)
		if err := s.ResourceStore.Create(ctx, r, store.CreateBy(rk)); err != nil {
			return err
		}
	}

	for _, r := range onUpdate {
		if err := s.ResourceStore.Update(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

type set []model.Resource

func (s set) contains(rk model.ResourceKey) bool {
	for _, r := range s {
		if rk == model.MetaToResourceKey(r.GetMeta()) {
			return true
		}
	}
	return false
}

func (s set) get(rk model.ResourceKey) model.Resource {
	for _, r := range s {
		if rk == model.MetaToResourceKey(r.GetMeta()) {
			return r
		}
	}
	return nil
}
