package store

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

// ResourceSyncer allows to synchronize resources in Store
type ResourceSyncer interface {
	// Sync method takes 'upstream' as a basis and synchronize underlying store.
	// It deletes all resources that absent in 'upstream', creates new resources that
	// are not represented in store yet and updates the rest.
	// Using 'PrefilterBy' option Sync allows to select scope of resources that will be
	// affected by Sync
	//
	// Sync takes into account only 'Name' and 'Mesh' when it comes to upstream's Meta.
	// 'Version', 'CreationTime' and 'ModificationTime' are managed by downstream store.
	Sync(upstream model.ResourceList, fs ...SyncOptionFunc) error
}

type SyncOption struct {
	Predicate func(r model.Resource) bool
	Zone      string
}

type SyncOptionFunc func(*SyncOption)

func NewSyncOptions(fs ...SyncOptionFunc) *SyncOption {
	opts := &SyncOption{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func Zone(name string) SyncOptionFunc {
	return func(opts *SyncOption) {
		opts.Zone = name
	}
}

func PrefilterBy(predicate func(r model.Resource) bool) SyncOptionFunc {
	return func(opts *SyncOption) {
		opts.Predicate = predicate
	}
}

type syncResourceStore struct {
	log           logr.Logger
	resourceStore store.ResourceStore
}

func NewResourceSyncer(log logr.Logger, resourceStore store.ResourceStore) ResourceSyncer {
	return &syncResourceStore{
		log:           log,
		resourceStore: resourceStore,
	}
}

func (s *syncResourceStore) Sync(upstream model.ResourceList, fs ...SyncOptionFunc) error {
	opts := NewSyncOptions(fs...)
	ctx := context.Background()
	log := s.log.WithValues("type", upstream.GetItemType())
	downstream, err := registry.Global().NewList(upstream.GetItemType())
	if err != nil {
		return err
	}
	if err := s.resourceStore.List(ctx, downstream); err != nil {
		return err
	}
	log.V(1).Info("before filtering", "downstream", downstream, "upstream", upstream)

	if opts.Predicate != nil {
		if filtered, err := filter(downstream, opts.Predicate); err != nil {
			return err
		} else {
			downstream = filtered
		}
		if filtered, err := filter(upstream, opts.Predicate); err != nil {
			return err
		} else {
			upstream = filtered
		}
	}
	log.V(1).Info("after filtering", "downstream", downstream, "upstream", upstream)

	indexedUpstream := newIndexed(upstream)
	indexedDownstream := newIndexed(downstream)

	// 1. delete resources from store which are not represented in 'upstream'
	onDelete := []model.Resource{}
	for _, r := range downstream.GetItems() {
		if indexedUpstream.get(model.MetaToResourceKey(r.GetMeta())) == nil {
			onDelete = append(onDelete, r)
		}
	}

	// 2. create resources which are not represented in 'downstream' and update the rest of them
	onCreate := []model.Resource{}
	onUpdate := []model.Resource{}
	for _, r := range upstream.GetItems() {
		existing := indexedDownstream.get(model.MetaToResourceKey(r.GetMeta()))
		if existing == nil {
			onCreate = append(onCreate, r)
			continue
		}
		if !proto.Equal(existing.GetSpec(), r.GetSpec()) {
			// we have to use meta of the current Store during update, because some Stores (Kubernetes, Memory)
			// expect to receive ResourceMeta of own type.
			r.SetMeta(existing.GetMeta())
			onUpdate = append(onUpdate, r)
		}
	}

	for _, r := range onDelete {
		rk := model.MetaToResourceKey(r.GetMeta())
		log.Info("deleting a resource since it's no longer available in the upstream", "name", r.GetMeta().GetName(), "mesh", r.GetMeta().GetMesh())
		if err := s.resourceStore.Delete(ctx, r, store.DeleteBy(rk)); err != nil {
			return err
		}
	}

	zone := system.NewZoneResource()
	if opts.Zone != "" && len(onCreate) > 0 {
		if err := s.resourceStore.Get(ctx, zone, store.GetByKey(opts.Zone, model.NoMesh)); err != nil {
			return err
		}
	}

	for _, r := range onCreate {
		rk := model.MetaToResourceKey(r.GetMeta())
		log.Info("creating a new resource from upstream", "name", r.GetMeta().GetName(), "mesh", r.GetMeta().GetMesh())
		creationTime := r.GetMeta().GetCreationTime()
		// some Stores try to cast ResourceMeta to own Store type that's why we have to set meta to nil
		r.SetMeta(nil)

		createOpts := []store.CreateOptionsFunc{
			store.CreateBy(rk),
			store.CreatedAt(creationTime),
		}
		if opts.Zone != "" {
			createOpts = append(createOpts, store.CreateWithOwner(zone))
		}
		if err := s.resourceStore.Create(ctx, r, createOpts...); err != nil {
			return err
		}
	}

	for _, r := range onUpdate {
		log.Info("updating a resource", "name", r.GetMeta().GetName(), "mesh", r.GetMeta().GetMesh())
		now := time.Now()
		// some stores manage ModificationTime time on they own (Kubernetes), in order to be consistent
		// we set ModificationTime when we add to downstream store. This time is almost the same with ModificationTime
		// from upstream store, because we update downstream only when resource have changed in upstream
		if err := s.resourceStore.Update(ctx, r, store.ModifiedAt(now)); err != nil {
			return err
		}
	}

	return nil
}

func filter(rs model.ResourceList, predicate func(r model.Resource) bool) (model.ResourceList, error) {
	rv, err := registry.Global().NewList(rs.GetItemType())
	if err != nil {
		return nil, err
	}
	for _, r := range rs.GetItems() {
		if predicate(r) {
			if err := rv.AddItem(r); err != nil {
				return nil, err
			}
		}
	}
	return rv, nil
}

type indexed struct {
	indexByResourceKey map[model.ResourceKey]model.Resource
}

func (i *indexed) get(rk model.ResourceKey) model.Resource {
	return i.indexByResourceKey[rk]
}

func newIndexed(rs model.ResourceList) *indexed {
	idxByRk := map[model.ResourceKey]model.Resource{}
	for _, r := range rs.GetItems() {
		idxByRk[model.MetaToResourceKey(r.GetMeta())] = r
	}
	return &indexed{indexByResourceKey: idxByRk}
}
