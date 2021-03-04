package reconcile

import (
	"context"
	"github.com/kumahq/kuma/pkg/mads"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	mads_v1_cache "github.com/kumahq/kuma/pkg/mads/v1/cache"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, resourceGenerator mads.ResourceGenerator) SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager:   resourceManager,
		resourceGenerator: resourceGenerator,
	}
}

type snapshotGenerator struct {
	resourceManager   core_manager.ReadOnlyResourceManager
	resourceGenerator mads.ResourceGenerator
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, _ *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	meshes, err := s.getMeshes(ctx)
	if err != nil {
		return nil, err
	}

	dataplanes, err := s.getDataplanes(ctx, meshes)
	if err != nil {
		return nil, err
	}

	args := mads.Args{
		Meshes:     meshes,
		Dataplanes: dataplanes,
	}

	resources, err := s.resourceGenerator.Generate(args)
	if err != nil {
		return nil, err
	}

	return mads_v1_cache.NewSnapshot("", core_xds.ResourceList(resources).ToIndex()), nil
}

func (s *snapshotGenerator) getMeshes(ctx context.Context) ([]*mesh_core.MeshResource, error) {
	meshList := &mesh_core.MeshResourceList{}
	if err := s.resourceManager.List(ctx, meshList); err != nil {
		return nil, err
	}

	meshes := make([]*mesh_core.MeshResource, 0)
	for _, mesh := range meshList.Items {
		if mesh.HasPrometheusMetricsEnabled() {
			meshes = append(meshes, mesh)
		}
	}
	return meshes, nil
}

func (s *snapshotGenerator) getDataplanes(ctx context.Context, meshes []*mesh_core.MeshResource) ([]*mesh_core.DataplaneResource, error) {
	dataplanes := make([]*mesh_core.DataplaneResource, 0)
	for _, mesh := range meshes {
		dataplaneList := &mesh_core.DataplaneResourceList{}
		if err := s.resourceManager.List(ctx, dataplaneList, core_store.ListByMesh(mesh.Meta.GetName())); err != nil {
			return nil, err
		}
		dataplanes = append(dataplanes, dataplaneList.Items...)
	}
	return dataplanes, nil
}
