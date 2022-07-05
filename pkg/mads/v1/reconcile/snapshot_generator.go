package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads/generator"
	mads_v1_cache "github.com/kumahq/kuma/pkg/mads/v1/cache"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, resourceGenerator generator.ResourceGenerator) util_xds_v3.SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager:   resourceManager,
		resourceGenerator: resourceGenerator,
	}
}

type snapshotGenerator struct {
	resourceManager   core_manager.ReadOnlyResourceManager
	resourceGenerator generator.ResourceGenerator
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, _ *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	meshes, err := s.getMeshesWithPrometheusEnabled(ctx)
	if err != nil {
		return nil, err
	}

	dataplanes, err := s.getDataplanes(ctx, meshes)
	if err != nil {
		return nil, err
	}

	meshGateways, err := s.getMeshGateways(ctx, meshes)
	if err != nil {
		return nil, err
	}

	args := generator.Args{
		Meshes:       meshes,
		Dataplanes:   dataplanes,
		MeshGateways: meshGateways,
	}

	resources, err := s.resourceGenerator.Generate(args)
	if err != nil {
		return nil, err
	}

	return mads_v1_cache.NewSnapshot("", core_xds.ResourceList(resources).ToIndex()), nil
}

func (s *snapshotGenerator) getMeshesWithPrometheusEnabled(ctx context.Context) ([]*core_mesh.MeshResource, error) {
	meshList := &core_mesh.MeshResourceList{}
	if err := s.resourceManager.List(ctx, meshList); err != nil {
		return nil, err
	}

	meshes := make([]*core_mesh.MeshResource, 0)
	for _, mesh := range meshList.Items {
		if mesh.HasPrometheusMetricsEnabled() {
			meshes = append(meshes, mesh)
		}
	}
	return meshes, nil
}

func (s *snapshotGenerator) getDataplanes(ctx context.Context, meshes []*core_mesh.MeshResource) ([]*core_mesh.DataplaneResource, error) {
	dataplanes := make([]*core_mesh.DataplaneResource, 0)
	for _, mesh := range meshes {
		dataplaneList := &core_mesh.DataplaneResourceList{}
		if err := s.resourceManager.List(ctx, dataplaneList, core_store.ListByMesh(mesh.Meta.GetName())); err != nil {
			return nil, err
		}
		dataplanes = append(dataplanes, dataplaneList.Items...)
	}
	return dataplanes, nil
}

func (s *snapshotGenerator) getMeshGateways(ctx context.Context, meshes []*core_mesh.MeshResource) ([]*core_mesh.MeshGatewayResource, error) {
	meshGateways := make([]*core_mesh.MeshGatewayResource, 0)
	for _, mesh := range meshes {
		meshGatewayList := &core_mesh.MeshGatewayResourceList{}
		if err := s.resourceManager.List(ctx, meshGatewayList, core_store.ListByMesh(mesh.Meta.GetName())); err != nil {
			return nil, err
		}
		meshGateways = append(meshGateways, meshGatewayList.Items...)
	}
	return meshGateways, nil
}
