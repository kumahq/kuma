package reconcile

import (
	"context"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	config_store "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/kds"
	kds_cache "github.com/kumahq/kuma/v3/pkg/kds/cache"
	"github.com/kumahq/kuma/v3/pkg/kds/util"
)

type (
	// ResourceFilter filters resources based on the context and the resource itself
	// return true if the resource should be included in the request sent to KDS
	ResourceFilter func(ctx context.Context, clusterID string, features kds.Features, r model.Resource) bool
	ResourceMapper func(features kds.Features, r model.Resource) (model.Resource, error)
)

func NoopResourceMapper(_ kds.Features, r model.Resource) (model.Resource, error) {
	return r, nil
}

func Any(context.Context, string, kds.Features, model.Resource) bool {
	return true
}

// TypeIs filters resources by type. This should be avoided and used only in exceptional cases.
// it is usually preferable to use a filter on a property of the ResourceTypeDescriptor
// are these are meant to define generic behaviors and not case by case exceptions.
func TypeIs(rtype model.ResourceType) func(model.Resource) bool {
	return func(r model.Resource) bool {
		return r.Descriptor().Name == rtype
	}
}

func IsKubernetes(storeType config_store.StoreType) func(model.Resource) bool {
	return func(_ model.Resource) bool {
		return storeType == config_store.KubernetesStore
	}
}

func HasStatus(r model.Resource) bool {
	return r.Descriptor().HasStatus
}

func NameHasPrefix(prefix string) func(model.Resource) bool {
	return func(r model.Resource) bool {
		return strings.HasPrefix(r.GetMeta().GetName(), prefix)
	}
}

func And(fs ...func(model.Resource) bool) func(model.Resource) bool {
	return func(r model.Resource) bool {
		for _, f := range fs {
			if !f(r) {
				return false
			}
		}
		return true
	}
}

func If(condition func(model.Resource) bool, m ResourceMapper) ResourceMapper {
	return func(features kds.Features, r model.Resource) (model.Resource, error) {
		if condition(r) {
			return m(features, r)
		}
		return r, nil
	}
}

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, filter ResourceFilter, mapper ResourceMapper) SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager: resourceManager,
		resourceFilter:  filter,
		resourceMapper:  mapper,
		mapped:          newMappedResourcesCache(),
	}
}

type snapshotGenerator struct {
	resourceManager core_manager.ReadOnlyResourceManager
	resourceFilter  ResourceFilter
	resourceMapper  ResourceMapper
	mapped          *mappedResourcesCache
}

func (s *snapshotGenerator) GenerateSnapshot(
	ctx context.Context,
	node *envoy_core.Node,
	builder kds_cache.SnapshotBuilder,
	resTypes map[model.ResourceType]struct{},
) (envoy_cache.ResourceSnapshot, error) {
	for typ := range resTypes {
		resources, err := s.getResources(ctx, typ, node)
		if err != nil {
			return nil, err
		}
		builder = builder.With(typ, resources)
	}

	return builder.Build(""), nil
}

func (s *snapshotGenerator) getResources(ctx context.Context, typ model.ResourceType, node *envoy_core.Node) ([]envoy_types.Resource, error) {
	rlist, err := registry.Global().NewList(typ)
	if err != nil {
		return nil, err
	}
	if err := s.resourceManager.List(ctx, rlist); err != nil {
		return nil, err
	}

	features := getFeatures(node)
	entry := s.mapped.entryFor(typ, features, rlist)

	resources := make([]envoy_types.Resource, 0, len(rlist.GetItems()))
	for _, r := range rlist.GetItems() {
		if !s.resourceFilter(ctx, node.GetId(), features, r) {
			continue
		}
		key := model.MetaToResourceKey(r.GetMeta())
		res, ok := entry.get(key)
		if !ok {
			mapped, err := s.resourceMapper(features, r)
			if err != nil {
				return nil, err
			}
			res, err = util.ToEnvoyResource(mapped)
			if err != nil {
				return nil, err
			}
			entry.put(key, res)
		}
		resources = append(resources, res)
	}

	return resources, nil
}

func getFeatures(node *envoy_core.Node) kds.Features {
	features := kds.Features{}
	for _, value := range node.GetMetadata().GetFields()[kds.MetadataFeatures].GetListValue().GetValues() {
		features[value.GetStringValue()] = true
	}
	return features
}
