package reconcile

import (
	"context"

	"github.com/Kong/kuma/pkg/kds/util"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/kds/cache"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, resourceTypes []model.ResourceType) SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager: resourceManager,
		resourceTypes:   resourceTypes,
	}
}

type snapshotGenerator struct {
	resourceManager core_manager.ReadOnlyResourceManager
	resourceTypes   []model.ResourceType
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, _ *envoy_core.Node) (util_xds.Snapshot, error) {
	builder := cache.NewSnapshotBuilder()
	for _, typ := range s.resourceTypes {
		resources, err := s.getResources(ctx, typ)
		if err != nil {
			return nil, err
		}
		builder = builder.With(string(typ), resources)
	}

	return builder.Build(""), nil
}

func (s *snapshotGenerator) getResources(context context.Context, typ model.ResourceType) ([]envoy_types.Resource, error) {
	rlist, err := registry.Global().NewList(typ)
	if err != nil {
		return nil, err
	}
	if err := s.resourceManager.List(context, rlist); err != nil {
		return nil, err
	}
	return util.ToEnvoyResources(rlist)
}
