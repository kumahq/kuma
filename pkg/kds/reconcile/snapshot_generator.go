package reconcile

import (
	"context"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/cache"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

type (
	ResourceFilter func(ctx context.Context, clusterID string, features kds.Features, r core_model.Resource) bool
	ResourceMapper func(features kds.Features, r core_model.Resource) (core_model.Resource, error)
)

func NoopResourceMapper(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
	return r, nil
}

func Any(context.Context, string, kds.Features, core_model.Resource) bool {
	return true
}

func TypeIs(rtype core_model.ResourceType) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return r.Descriptor().Name == rtype
	}
}

func IsKubernetes(storeType config_store.StoreType) func(core_model.Resource) bool {
	return func(_ core_model.Resource) bool {
		return storeType == config_store.KubernetesStore
	}
}

func ScopeIs(s core_model.ResourceScope) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return r.Descriptor().Scope == s
	}
}

func NameHasPrefix(prefix string) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return strings.HasPrefix(r.GetMeta().GetName(), prefix)
	}
}

func Not(f func(core_model.Resource) bool) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return !f(r)
	}
}

func And(fs ...func(core_model.Resource) bool) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		for _, f := range fs {
			if !f(r) {
				return false
			}
		}
		return true
	}
}

func If(condition func(core_model.Resource) bool, m ResourceMapper) ResourceMapper {
	return func(features kds.Features, r core_model.Resource) (core_model.Resource, error) {
		if condition(r) {
			return m(features, r)
		}
		return r, nil
	}
}

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, types []core_model.ResourceType, filter ResourceFilter, mapper ResourceMapper) SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager: resourceManager,
		resourceTypes:   types,
		resourceFilter:  filter,
		resourceMapper:  mapper,
	}
}

type snapshotGenerator struct {
	resourceManager core_manager.ReadOnlyResourceManager
	resourceTypes   []core_model.ResourceType
	resourceFilter  ResourceFilter
	resourceMapper  ResourceMapper
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, node *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	builder := cache.NewSnapshotBuilder()
	for _, typ := range s.resourceTypes {
		resources, err := s.getResources(ctx, typ, node)
		if err != nil {
			return nil, err
		}
		builder = builder.With(string(typ), resources)
	}

	return builder.Build(""), nil
}

func (s *snapshotGenerator) getResources(ctx context.Context, typ core_model.ResourceType, node *envoy_core.Node) ([]envoy_types.Resource, error) {
	rlist, err := registry.Global().NewList(typ)
	if err != nil {
		return nil, err
	}
	if err := s.resourceManager.List(ctx, rlist); err != nil {
		return nil, err
	}

	resources, err := s.mapper(s.filter(ctx, rlist, node), node)
	if err != nil {
		return nil, err
	}

	return util.ToEnvoyResources(resources)
}

func (s *snapshotGenerator) filter(ctx context.Context, rs core_model.ResourceList, node *envoy_core.Node) core_model.ResourceList {
	features := getFeatures(node)

	rv := registry.Global().MustNewList(rs.GetItemType())
	for _, r := range rs.GetItems() {
		if s.resourceFilter(ctx, node.GetId(), features, r) {
			_ = rv.AddItem(r)
		}
	}
	return rv
}

func (s *snapshotGenerator) mapper(rs core_model.ResourceList, node *envoy_core.Node) (core_model.ResourceList, error) {
	features := getFeatures(node)

	rv := registry.Global().MustNewList(rs.GetItemType())
	for _, r := range rs.GetItems() {
		resource, err := s.resourceMapper(features, r)
		if err != nil {
			return nil, err
		}

		if err := rv.AddItem(resource); err != nil {
			return nil, err
		}
	}

	return rv, nil
}

func getFeatures(node *envoy_core.Node) kds.Features {
	features := kds.Features{}
	for _, value := range node.GetMetadata().GetFields()[kds.MetadataFeatures].GetListValue().GetValues() {
		features[value.GetStringValue()] = true
	}
	return features
}
