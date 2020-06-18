package generator

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/kds"
	"github.com/Kong/kuma/pkg/kds/cache"
	mads_reconcile "github.com/Kong/kuma/pkg/mads/reconcile"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager) mads_reconcile.SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager: resourceManager,
	}
}

type snapshotGenerator struct {
	resourceManager core_manager.ReadOnlyResourceManager
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, _ *envoy_core.Node) (util_xds.Snapshot, error) {
	builder := cache.NewSnapshotBuilder()
	for _, typ := range kds.SupportedTypes {
		resources, err := s.getResources(ctx, typ)
		if err != nil {
			return nil, err
		}
		builder = builder.With(kds.TypeURL(typ), resources)
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
	return convert(rlist)
}

func convert(rlist model.ResourceList) ([]envoy_types.Resource, error) {
	rv := make([]envoy_types.Resource, 0, len(rlist.GetItems()))
	for _, r := range rlist.GetItems() {
		pbany, err := ptypes.MarshalAny(r.GetSpec())
		if err != nil {
			return nil, err
		}
		rv = append(rv, &mesh_proto.KumaResource{
			Meta: &mesh_proto.KumaResource_Meta{
				Name: r.GetMeta().GetName(),
				Mesh: r.GetMeta().GetMesh(),
			},
			Spec: pbany,
		})
	}
	return rv, nil
}
