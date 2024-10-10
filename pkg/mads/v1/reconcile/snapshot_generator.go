package reconcile

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads/generator"
	mads_v1_cache "github.com/kumahq/kuma/pkg/mads/v1/cache"
	meshmetrics_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
)

var log = core.Log.WithName("mads").WithName("v1").WithName("reconcile")

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, resourceGenerator generator.ResourceGenerator, meshCache *mesh.Cache) *SnapshotGenerator {
	return &SnapshotGenerator{
		resourceManager:   resourceManager,
		resourceGenerator: resourceGenerator,
		meshCache:         meshCache,
	}
}

type SnapshotGenerator struct {
	resourceManager   core_manager.ReadOnlyResourceManager
	resourceGenerator generator.ResourceGenerator
	meshCache         *mesh.Cache
}

func (s *SnapshotGenerator) GenerateSnapshot(ctx context.Context) (map[string]util_xds_v3.Snapshot, error) {
	meshesWithMeshMetrics, err := s.getMeshesWithMeshMetrics(ctx)
	if err != nil {
		return nil, err
	}

	meshes, err := s.getMeshesWithPrometheusEnabled(ctx)
	if err != nil {
		return nil, err
	}

	if len(meshes) > 0 && len(meshesWithMeshMetrics) > 0 {
		log.Info("it is not supported to use both MeshMetrics policy and 'metrics' under Mesh resource. For now MeshMetrics will take precedence. If migrating please remove the 'metrics' section and apply an equivalent MeshMetrics resource")
	}

	var resources []*core_xds.Resource
	resourcesPerClientId := map[string]util_xds_v3.Snapshot{}
	if len(meshesWithMeshMetrics) == 0 {
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

		resources, err = s.resourceGenerator.Generate(args)
		if err != nil {
			return nil, err
		}
		resourcesPerClientId[meshmetrics_generator.DefaultKumaClientId] = createSnapshot(resources)
	} else {
		for clientId, meshes := range meshesWithMeshMetrics {
			meshMetricConfToDataplanes, err := s.getMatchingDataplanes(ctx, meshes)
			if err != nil {
				return nil, err
			}

			resources, err = meshmetrics_generator.Generate(meshMetricConfToDataplanes, clientId)
			if err != nil {
				return nil, err
			}
			resourcesPerClientId[clientId] = createSnapshot(resources)
		}
	}

	return resourcesPerClientId, nil
}

func (s *SnapshotGenerator) getMeshesWithMeshMetrics(ctx context.Context) (map[string][]string, error) {
	meshMetricsList := v1alpha1.MeshMetricResourceList{}
	if err := s.resourceManager.List(ctx, &meshMetricsList); err != nil {
		return nil, err
	}

	clientToMeshes := map[string][]string{}
	for _, meshMetric := range meshMetricsList.Items {
		if meshMetric.Spec.Default.Backends == nil {
			continue
		}
		meshName := meshMetric.GetMeta().GetMesh()
		for _, backend := range *meshMetric.Spec.Default.Backends { // can backends be nil?
			// match against client ID or fallback to "" when specified by user
			if backend.Type == v1alpha1.PrometheusBackendType {
				client := pointer.DerefOr(backend.Prometheus.ClientId, meshmetrics_generator.DefaultKumaClientId)
				clientToMeshes[client] = append(clientToMeshes[client], meshName)
			}
		}
	}

	return clientToMeshes, nil
}

func (s *SnapshotGenerator) getMatchingDataplanes(ctx context.Context, meshesWithMeshMetrics []string) (map[*v1alpha1.Conf]*core_mesh.DataplaneResource, error) {
	meshMetricConfToDataplanes := map[*v1alpha1.Conf]*core_mesh.DataplaneResource{}
	for _, meshName := range meshesWithMeshMetrics {
		meshContext, err := s.meshCache.GetMeshContext(ctx, meshName)
		if err != nil {
			return nil, errors.Wrap(err, "could not get mesh context")
		}

		for _, dp := range meshContext.DataplanesByName {
			matchedPolicies, err := matchers.MatchedPolicies(v1alpha1.MeshMetricType, dp, meshContext.Resources)
			if err != nil {
				return nil, errors.Wrap(err, "error on matching dpp")
			}
			if len(matchedPolicies.SingleItemRules.Rules) == 1 {
				conf := matchedPolicies.SingleItemRules.Rules[0].Conf.(v1alpha1.Conf)
				meshMetricConfToDataplanes[&conf] = dp
			}
		}
	}

	return meshMetricConfToDataplanes, nil
}

func (s *SnapshotGenerator) getMeshesWithPrometheusEnabled(ctx context.Context) ([]*core_mesh.MeshResource, error) {
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

func (s *SnapshotGenerator) getDataplanes(ctx context.Context, meshes []*core_mesh.MeshResource) ([]*core_mesh.DataplaneResource, error) {
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

func (s *SnapshotGenerator) getMeshGateways(ctx context.Context, meshes []*core_mesh.MeshResource) ([]*core_mesh.MeshGatewayResource, error) {
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

func createSnapshot(resources []*core_xds.Resource) *mads_v1_cache.Snapshot {
	return mads_v1_cache.NewSnapshot(core.NewUUID(), core_xds.ResourceList(resources).ToIndex())
}
