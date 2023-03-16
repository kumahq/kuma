package store

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/kds/util"
	zone_client "github.com/kumahq/kuma/pkg/kds/v2/zone/client"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

// ResourceSyncer allows to synchronize resources in Store
type ResourceSyncer interface {
	// Sync method takes 'upstream' as a basis and synchronize underlying store.
	// It deletes resources that were removed in the upstream, creates new resources that
	// are not represented in store yet and updates the rest.
	// Using 'PrefilterBy' option Sync allows to select scope of resources that will be
	// affected by Sync
	//
	// Sync takes into account only 'Name' and 'Mesh' when it comes to upstream's Meta.
	// 'Version', 'CreationTime' and 'ModificationTime' are managed by downstream store.
	Sync(upstream zone_client.UpstreamResponse, fs ...SyncOptionFunc) error
}

type SyncOption struct {
	Predicate func(r core_model.Resource) bool
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

func PrefilterBy(predicate func(r core_model.Resource) bool) SyncOptionFunc {
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

func (s *syncResourceStore) Sync(upstreamResponse zone_client.UpstreamResponse, fs ...SyncOptionFunc) error {
	opts := NewSyncOptions(fs...)
	ctx := user.Ctx(context.TODO(), user.ControlPlane)
	log := s.log.WithValues("type", upstreamResponse.Type)
	upstream := upstreamResponse.AddedResources
	downstream, err := registry.Global().NewList(upstreamResponse.Type)
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

	indexedDownstream := newIndexed(downstream)
	indexedUpstream := newIndexed(upstream)

	onDelete := []core_model.Resource{}
	// 1. delete resources which were removed from the upstream
	// on the first request when the control-plane starts we want to sync
	// whole the resources in the store. In this case we do not check removed
	// resources because we want to make stores synced. When we already
	// have resources in the map, we are going to receive only updates
	// so we don't want to remove resources haven't changed.
	if upstreamResponse.IsInitialRequest {
		for _, r := range downstream.GetItems() {
			if indexedUpstream.get(core_model.MetaToResourceKey(r.GetMeta())) == nil {
				onDelete = append(onDelete, r)
			}
		}
	} else {
		removed := getRemovedResourceKeys(upstreamResponse.RemovedResourceNames)
		for _, rk := range removed {
			// check if we are adding and removing the resource at the same time
			if r := indexedUpstream.get(rk); r != nil {
				// it isn't remove but update
				continue
			}
			if r := indexedDownstream.get(rk); r != nil {
				onDelete = append(onDelete, r)
			}
		}
	}

	// 2. create resources which are not represented in 'downstream' and update the rest of them
	onCreate := []core_model.Resource{}
	onUpdate := []core_model.Resource{}
	for _, r := range upstream.GetItems() {
		existing := indexedDownstream.get(core_model.MetaToResourceKey(r.GetMeta()))
		if existing == nil {
			onCreate = append(onCreate, r)
			continue
		}
		if !core_model.Equal(existing.GetSpec(), r.GetSpec()) {
			// we have to use meta of the current Store during update, because some Stores (Kubernetes, Memory)
			// expect to receive ResourceMeta of own type.
			r.SetMeta(existing.GetMeta())
			onUpdate = append(onUpdate, r)
		}
	}

	for _, r := range onDelete {
		rk := core_model.MetaToResourceKey(r.GetMeta())
		log.Info("deleting a resource since it's no longer available in the upstream", "name", r.GetMeta().GetName(), "mesh", r.GetMeta().GetMesh())
		if err := s.resourceStore.Delete(ctx, r, store.DeleteBy(rk)); err != nil {
			return err
		}
	}

	zone := system.NewZoneResource()
	if opts.Zone != "" && len(onCreate) > 0 {
		if err := s.resourceStore.Get(ctx, zone, store.GetByKey(opts.Zone, core_model.NoMesh)); err != nil {
			return err
		}
	}

	for _, r := range onCreate {
		rk := core_model.MetaToResourceKey(r.GetMeta())
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

func filter(rs core_model.ResourceList, predicate func(r core_model.Resource) bool) (core_model.ResourceList, error) {
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
	indexByResourceKey map[core_model.ResourceKey]core_model.Resource
}

func (i *indexed) get(rk core_model.ResourceKey) core_model.Resource {
	return i.indexByResourceKey[rk]
}

func newIndexed(rs core_model.ResourceList) *indexed {
	idxByRk := map[core_model.ResourceKey]core_model.Resource{}
	for _, r := range rs.GetItems() {
		idxByRk[core_model.MetaToResourceKey(r.GetMeta())] = r
	}
	return &indexed{indexByResourceKey: idxByRk}
}

func Callbacks(
	configToSync map[string]bool,
	syncer ResourceSyncer,
	k8sStore bool,
	localZone string,
	kubeFactory resources_k8s.KubeFactory,
	systemNamespace string,
) *zone_client.Callbacks {
	return &zone_client.Callbacks{
		OnResourcesReceived: func(upstream zone_client.UpstreamResponse) error {
			if k8sStore && upstream.Type != system.ConfigType && upstream.Type != system.SecretType && upstream.Type != system.GlobalSecretType {
				// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
				// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
				// System resources are not in the kubeFactory therefore we need explicit ifs for them
				kubeObject, err := kubeFactory.NewObject(upstream.AddedResources.NewItem())
				if err != nil {
					return errors.Wrap(err, "could not convert object")
				}
				if kubeObject.Scope() == k8s_model.ScopeNamespace {
					util.AddSuffixToNames(upstream.AddedResources.GetItems(), systemNamespace)
				}
			}
			if upstream.Type == mesh.ZoneIngressType {
				return syncer.Sync(upstream, PrefilterBy(func(r core_model.Resource) bool {
					return r.(*mesh.ZoneIngressResource).IsRemoteIngress(localZone)
				}))
			}
			if upstream.Type == system.ConfigType {
				return syncer.Sync(upstream, PrefilterBy(func(r core_model.Resource) bool {
					return configToSync[r.GetMeta().GetName()]
				}))
			}
			if upstream.Type == system.GlobalSecretType {
				return syncer.Sync(upstream, PrefilterBy(func(r core_model.Resource) bool {
					return util.ResourceNameHasAtLeastOneOfPrefixes(
						r.GetMeta().GetName(),
						zoneingress.ZoneIngressSigningKeyPrefix,
						zone_tokens.SigningPublicKeyPrefix,
					)
				}))
			}
			return syncer.Sync(upstream)
		},
	}
}

func getRemovedResourceKeys(removedResourceNames []string) []core_model.ResourceKey {
	removed := []core_model.ResourceKey{}
	for _, resourceName := range removedResourceNames {
		index := strings.LastIndex(resourceName, ".")
		var rk core_model.ResourceKey
		if index != -1 {
			rk = core_model.ResourceKey{
				Mesh: resourceName[index+1:],
				Name: resourceName[:index],
			}
		}
		removed = append(removed, rk)
	}
	return removed
}