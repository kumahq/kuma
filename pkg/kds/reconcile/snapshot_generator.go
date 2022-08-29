package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/cache"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

type ResourceFilter func(ctx context.Context, clusterID string, features kds.Features, r model.Resource) bool
type ResourceMapper func(r model.Resource) (model.Resource, error)

func NoopResourceMapper(r model.Resource) (model.Resource, error) {
	return r, nil
}

func Any(context.Context, string, kds.Features, model.Resource) bool {
	return true
}

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, types []model.ResourceType, filter ResourceFilter, mapper ResourceMapper) SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager: resourceManager,
		resourceTypes:   types,
		resourceFilter:  filter,
		resourceMapper:  mapper,
	}
}

type snapshotGenerator struct {
	resourceManager core_manager.ReadOnlyResourceManager
	resourceTypes   []model.ResourceType
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

func (s *snapshotGenerator) getResources(ctx context.Context, typ model.ResourceType, node *envoy_core.Node) ([]envoy_types.Resource, error) {
	rlist, err := registry.Global().NewList(typ)
	if err != nil {
		return nil, err
	}
	if err := s.resourceManager.List(ctx, rlist); err != nil {
		return nil, err
	}

	resources, err := s.mapper(s.filter(ctx, rlist, node))
	if err != nil {
		return nil, err
	}

	return util.ToEnvoyResources(resources)
}

func (s *snapshotGenerator) filter(ctx context.Context, rs model.ResourceList, node *envoy_core.Node) model.ResourceList {
	features := kds.Features{}
	for _, value := range node.GetMetadata().GetFields()[kds.MetadataFeatures].GetListValue().GetValues() {
		features[value.GetStringValue()] = true
	}

	rv, _ := registry.Global().NewList(rs.GetItemType())
	for _, r := range rs.GetItems() {
		if s.resourceFilter(ctx, node.GetId(), features, r) {
			_ = rv.AddItem(r)
		}
	}
	return rv
}

func (s *snapshotGenerator) mapper(rs model.ResourceList) (model.ResourceList, error) {
	rv, _ := registry.Global().NewList(rs.GetItemType())

	for _, r := range rs.GetItems() {
		resource, err := s.resourceMapper(r)
		if err != nil {
			return nil, err
		}

		if err := rv.AddItem(resource); err != nil {
			return nil, err
		}
	}

	return rv, nil
}
