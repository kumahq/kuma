package reconcile

import (
	"context"

	"github.com/Kong/kuma/pkg/core/resources/model"

	"github.com/Kong/kuma/pkg/kds/util"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/kds/cache"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

type ResourceFilter func(clusterID string, r model.Resource) bool

func Any(clusterID string, r model.Resource) bool {
	return true
}

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, types []model.ResourceType, filter ResourceFilter) SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager: resourceManager,
		resourceTypes:   types,
		resourceFilter:  filter,
	}
}

type snapshotGenerator struct {
	resourceManager core_manager.ReadOnlyResourceManager
	resourceTypes   []model.ResourceType
	resourceFilter  ResourceFilter
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, node *envoy_core.Node) (util_xds.Snapshot, error) {
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

func (s *snapshotGenerator) getResources(context context.Context, typ model.ResourceType, node *envoy_core.Node) ([]envoy_types.Resource, error) {
	rlist, err := registry.Global().NewList(typ)
	if err != nil {
		return nil, err
	}
	if err := s.resourceManager.List(context, rlist); err != nil {
		return nil, err
	}
	return util.ToEnvoyResources(s.filter(rlist, node))
}

func (s *snapshotGenerator) filter(rs model.ResourceList, node *envoy_core.Node) model.ResourceList {
	rv, _ := registry.Global().NewList(rs.GetItemType())
	for _, r := range rs.GetItems() {
		if s.resourceFilter(node.GetId(), r) {
			_ = rv.AddItem(r)
		}
	}
	return rv
}
