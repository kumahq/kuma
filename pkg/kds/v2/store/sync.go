package store

import (
	"context"
	std_errors "errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/util"
	client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
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
	// We return now 2 errors:
	// The first error cancels the connection, forcing the client to re-establish it.
	// This can happen due to a database error.
	//
	// The second error is related to non-critical issues, such as `ResourceAlreadyExists`,
	// which shouldn't prevent other resources from being stored.
	// Instead, we return a NACK message with information for the user.
	Sync(ctx context.Context, upstream client_v2.UpstreamResponse, fs ...SyncOptionFunc) (error, error)
}

type SyncOption struct {
	Predicate            func(r core_model.Resource) bool
	SkipConflictResource bool
	Zone                 string
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

func SkipConflictResource() SyncOptionFunc {
	return func(opts *SyncOption) {
		opts.SkipConflictResource = true
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
	transactions  store.Transactions
	metric        prometheus.Histogram
	extensions    context.Context
}

func NewResourceSyncer(
	log logr.Logger,
	resourceStore store.ResourceStore,
	transactions store.Transactions,
	metrics core_metrics.Metrics,
	extensions context.Context,
) (ResourceSyncer, error) {
	metric := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "kds_resources_sync",
		Help: "Time it took to sync resources from the upstream over KDS",
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &syncResourceStore{
		log:           log,
		resourceStore: resourceStore,
		transactions:  transactions,
		metric:        metric,
		extensions:    extensions,
	}, nil
}

type OnUpdate struct {
	r    core_model.Resource
	opts []store.UpdateOptionsFunc
}

func (s *syncResourceStore) Sync(syncCtx context.Context, upstreamResponse client_v2.UpstreamResponse, fs ...SyncOptionFunc) (error, error) {
	now := core.Now()
	defer func() {
		s.metric.Observe(float64(time.Since(now).Milliseconds()) / 1000)
	}()
	opts := NewSyncOptions(fs...)
	ctx := user.Ctx(syncCtx, user.ControlPlane)
	log := s.log.WithValues("type", upstreamResponse.Type)
	log = kuma_log.AddFieldsFromCtx(log, ctx, s.extensions)
	upstream := upstreamResponse.AddedResources
	downstream, err := registry.Global().NewList(upstreamResponse.Type)
	indexedInvalidResources := core_model.IndexKeys(upstreamResponse.InvalidResourcesKey)
	if err != nil {
		return err, nil
	}
	if upstreamResponse.IsInitialRequest {
		if err := s.resourceStore.List(ctx, downstream); err != nil {
			return err, nil
		}
	} else {
		upstreamChangeKeys := append(core_model.ResourceListToResourceKeys(upstream), upstreamResponse.RemovedResourcesKey...)
		if err := s.resourceStore.List(ctx, downstream, store.ListByResourceKeys(upstreamChangeKeys)); err != nil {
			return err, nil
		}
	}
	// if there are no resources in the upstream and downstream then we don't need to do anything
	if len(downstream.GetItems()) == 0 && len(upstream.GetItems()) == 0 {
		return nil, nil
	}

	filterResources := func(resources core_model.ResourceList, predicates []func(core_model.Resource) bool) (core_model.ResourceList, error) {
		return filter(resources, func(r core_model.Resource) bool {
			includeResource := true
			for _, predicate := range predicates {
				includeResource = includeResource && predicate(r)
			}
			return includeResource
		})
	}

	isValidResource := func(r core_model.Resource) bool {
		_, exists := indexedInvalidResources[core_model.MetaToResourceKey(r.GetMeta())]
		return !exists
	}

	predicate := []func(r core_model.Resource) bool{isValidResource}
	if opts.Predicate != nil {
		predicate = append(predicate, opts.Predicate)
	}

	log.V(1).Info("before filtering", "downstream", downstream, "upstream", upstream)

	if downstream, err = filterResources(downstream, predicate); err != nil {
		return err, nil
	}
	if upstream, err = filterResources(upstream, predicate); err != nil {
		return err, nil
	}

	log.V(1).Info("after filtering", "downstream", downstream, "upstream", upstream)

	indexedDownstream := core_model.IndexByKey(downstream.GetItems())
	indexedUpstream := core_model.IndexByKey(upstream.GetItems())

	onDelete := []core_model.Resource{}
	// 1. delete resources which were removed from the upstream
	// on the first request when the control-plane starts we want to sync
	// whole the resources in the store. In this case we do not check removed
	// resources because we want to make stores synced. When we already
	// have resources in the map, we are going to receive only updates
	// so we don't want to remove resources haven't changed.
	if upstreamResponse.IsInitialRequest {
		for _, r := range downstream.GetItems() {
			if indexedUpstream[core_model.MetaToResourceKey(r.GetMeta())] == nil {
				onDelete = append(onDelete, r)
			}
		}
	} else {
		for _, rk := range upstreamResponse.RemovedResourcesKey {
			// check if we are adding and removing the resource at the same time
			if r := indexedUpstream[rk]; r != nil {
				// it isn't remove but update
				continue
			}
			if r := indexedDownstream[rk]; r != nil {
				onDelete = append(onDelete, r)
			}
		}
	}

	// 2. create resources which are not represented in 'downstream' and update the rest of them
	onCreate := []core_model.Resource{}
	onUpdate := []OnUpdate{}
	for _, r := range upstream.GetItems() {
		existing := indexedDownstream[core_model.MetaToResourceKey(r.GetMeta())]
		if existing == nil {
			onCreate = append(onCreate, r)
			continue
		}
		newLabels := r.GetMeta().GetLabels()
		if !core_model.Equal(existing.GetSpec(), r.GetSpec()) ||
			!maps.Equal(existing.GetMeta().GetLabels(), newLabels) ||
			!core_model.Equal(existing.GetStatus(), r.GetStatus()) {
			// we have to use meta of the current Store during update, because some Stores (Kubernetes, Memory)
			// expect to receive ResourceMeta of own type.
			r.SetMeta(existing.GetMeta())
			onUpdate = append(onUpdate, OnUpdate{r: r, opts: []store.UpdateOptionsFunc{store.UpdateWithLabels(newLabels)}})
		}
	}

	zone := system.NewZoneResource()
	if opts.Zone != "" && len(onCreate) > 0 {
		if err := s.resourceStore.Get(ctx, zone, store.GetByKey(opts.Zone, core_model.NoMesh)); err != nil {
			return err, nil
		}
	}
	var nackError error
	err = store.InTx(ctx, s.transactions, func(ctx context.Context) error {
		for _, r := range onCreate {
			rk := core_model.MetaToResourceKey(r.GetMeta())
			log.Info("creating a new resource from upstream", "name", r.GetMeta().GetName(), "mesh", r.GetMeta().GetMesh())

			createOpts := []store.CreateOptionsFunc{
				store.CreateBy(rk),
				store.CreatedAt(core.Now()),
				store.CreateWithLabels(r.GetMeta().GetLabels()),
			}
			if opts.Zone != "" {
				createOpts = append(createOpts, store.CreateWithOwner(zone))
			}

			// some Stores try to cast ResourceMeta to own Store type that's why we have to set meta to nil
			r.SetMeta(nil)

			if err := s.resourceStore.Create(ctx, r, createOpts...); err != nil {
				if opts.SkipConflictResource && store.IsResourceAlreadyExists(err) {
					nackError = std_errors.Join(nackError, err)
				} else {
					return err
				}
			}
		}

		for _, r := range onDelete {
			rk := core_model.MetaToResourceKey(r.GetMeta())
			log.Info("deleting a resource since it's no longer available in the upstream", "name", r.GetMeta().GetName(), "mesh", r.GetMeta().GetMesh())
			if err := s.resourceStore.Delete(ctx, r, store.DeleteBy(rk)); err != nil {
				return err
			}
		}

		for _, upd := range onUpdate {
			log.V(1).Info("updating a resource", "name", upd.r.GetMeta().GetName(), "mesh", upd.r.GetMeta().GetMesh())
			now := time.Now()
			// some stores manage ModificationTime time on they own (Kubernetes), in order to be consistent
			// we set ModificationTime when we add to downstream store. This time is almost the same with ModificationTime
			// from upstream store, because we update downstream only when resource have changed in upstream
			if err := s.resourceStore.Update(ctx, upd.r, append(upd.opts, store.ModifiedAt(now))...); err != nil {
				return err
			}
		}
		return nil
	})
	return err, nackError
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

func ZoneSyncCallback(ctx context.Context, syncer ResourceSyncer, k8sStore bool, kubeFactory resources_k8s.KubeFactory, systemNamespace string) *client_v2.Callbacks {
	return &client_v2.Callbacks{
		OnResourcesReceived: func(upstream client_v2.UpstreamResponse) (error, error) {
			tDesc := upstream.AddedResources.NewItem().Descriptor()
			if k8sStore && !tDesc.SkipKDSHash {
				if err := addNamespaceSuffix(kubeFactory, upstream, systemNamespace); err != nil {
					return err, nil
				}
			}

			syncOptions := []SyncOptionFunc{
				PrefilterBy(func(r core_model.Resource) bool {
					// If it's not provided by zone we should always override as this is when we prefer global over zone
					// this can happen when we've just started federation
					if !tDesc.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
						return true
					}
					return !core_model.IsLocallyOriginated(config_core.Zone, r.GetMeta().GetLabels())
				}),
			}
			// When there is no KDS hash suffix, we can have conflicts in resources so we simply skip them
			if tDesc.SkipKDSHash {
				syncOptions = append(syncOptions, SkipConflictResource())
			}

			return syncer.Sync(ctx, upstream, syncOptions...)
		},
	}
}

func GlobalSyncCallback(
	ctx context.Context,
	syncer ResourceSyncer,
	k8sStore bool,
	kubeFactory resources_k8s.KubeFactory,
	systemNamespace string,
) *client_v2.Callbacks {
	supportsHashSuffixes := kds.ContextHasFeature(ctx, kds.FeatureHashSuffix)

	return &client_v2.Callbacks{
		OnResourcesReceived: func(upstream client_v2.UpstreamResponse) (error, error) {
			if !supportsHashSuffixes {
				// todo: remove in 2 releases after 2.6.x
				upstream.RemovedResourcesKey = util.AddPrefixToResourceKeyNames(upstream.RemovedResourcesKey, upstream.ControlPlaneId)
				upstream.InvalidResourcesKey = util.AddPrefixToResourceKeyNames(upstream.InvalidResourcesKey, upstream.ControlPlaneId)
				util.AddPrefixToNames(upstream.AddedResources.GetItems(), upstream.ControlPlaneId)
			}

			for _, r := range upstream.AddedResources.GetItems() {
				r.SetMeta(util.CloneResourceMeta(r.GetMeta(),
					util.WithLabel(mesh_proto.ZoneTag, upstream.ControlPlaneId),
					util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin)),
				))
			}

			if k8sStore {
				if err := addNamespaceSuffix(kubeFactory, upstream, systemNamespace); err != nil {
					return err, nil
				}
			}

			switch upstream.Type {
			case core_mesh.ZoneIngressType:
				for _, zi := range upstream.AddedResources.(*core_mesh.ZoneIngressResourceList).Items {
					zi.Spec.Zone = upstream.ControlPlaneId
				}
			case core_mesh.ZoneEgressType:
				for _, ze := range upstream.AddedResources.(*core_mesh.ZoneEgressResourceList).Items {
					ze.Spec.Zone = upstream.ControlPlaneId
				}
			}

			return syncer.Sync(ctx, upstream, PrefilterBy(func(r core_model.Resource) bool {
				// Assuming the global CP was updated first, the prefix check is only necessary
				// if the client doesn't have `supportsHashSuffixes`.
				// But maybe some prefixed resources were previously synced, we
				// can filter them and the zone won't sync them again.
				// When we are further from this migration we can remove this
				// check.
				hasOldStylePrefix := strings.HasPrefix(r.GetMeta().GetName(), fmt.Sprintf("%s.", upstream.ControlPlaneId))
				hasZoneLabel := r.GetMeta().GetLabels()[mesh_proto.ZoneTag] == upstream.ControlPlaneId
				return hasOldStylePrefix || hasZoneLabel
			}), Zone(upstream.ControlPlaneId))
		},
	}
}

func addNamespaceSuffix(kubeFactory resources_k8s.KubeFactory, upstream client_v2.UpstreamResponse, ns string) error {
	// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
	// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
	kubeObject, err := kubeFactory.NewObject(upstream.AddedResources.NewItem())
	if err != nil {
		return errors.Wrap(err, "could not convert object")
	}
	if kubeObject.Scope() == k8s_model.ScopeNamespace {
		util.AddSuffixToNames(upstream.AddedResources.GetItems(), ns)
		upstream.RemovedResourcesKey = util.AddSuffixToResourceKeyNames(upstream.RemovedResourcesKey, ns)
		upstream.InvalidResourcesKey = util.AddSuffixToResourceKeyNames(upstream.InvalidResourcesKey, ns)
	}
	return nil
}
