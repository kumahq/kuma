package reconcile

import (
	"context"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
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
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var log = core.Log.WithName("mads").WithName("v1").WithName("reconcile")

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, resourceGenerator generator.ResourceGenerator, meshCache *mesh.Cache) util_xds_v3.SnapshotGenerator {
	return &snapshotGenerator{
		resourceManager:   resourceManager,
		resourceGenerator: resourceGenerator,
		meshCache: meshCache,
	}
}

type snapshotGenerator struct {
	resourceManager   core_manager.ReadOnlyResourceManager
	resourceGenerator generator.ResourceGenerator
	meshCache         *mesh.Cache
}

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, node *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	meshesWithMeshMetrics, err := s.getMeshesWithMeshMetrics(ctx, node.Id)
	if err != nil {
		return nil, err
	}

	meshes, err := s.getMeshesWithPrometheusEnabled(ctx)
	if err != nil {
		return nil, err
	}

	if len(meshes) > 0 && len(meshesWithMeshMetrics) > 0 {
		log.Info("it is not supported to use both MeshMetrics policy and 'metrics' under Mesh resource. If migrating please remove the 'metrics' section and apply an equivalent MeshMetrics resource")
	}

	var resources []*core_xds.Resource
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
	} else {
		dataplanes, err := s.getMatchingDataplanes(ctx, meshesWithMeshMetrics)
		if err != nil {
			return nil, err
		}

		resources, err = meshmetrics_generator.Generate(dataplanes)
		if err != nil {
			return nil, err
		}
	}

	return mads_v1_cache.NewSnapshot("", core_xds.ResourceList(resources).ToIndex()), nil
}

func (s *snapshotGenerator) getMeshesWithMeshMetrics(ctx context.Context, clientId string) ([]string, error) {
	meshMetricsList := v1alpha1.MeshMetricResourceList{}
	if err := s.resourceManager.List(ctx, &meshMetricsList); err != nil {
		return nil, err
	}

	var meshes []string
	for _, meshMetric := range meshMetricsList.Items {
		for _, backend := range *meshMetric.Spec.Default.Backends { // can backends be nil?
			// match against client ID or fallback to "" when specified by user
			if backend.Type == v1alpha1.PrometheusBackendType && (backend.Prometheus.ClientId == nil || *backend.Prometheus.ClientId == clientId || *backend.Prometheus.ClientId == "") {
				meshes = append(meshes, meshMetric.GetMeta().GetMesh())
			}
		}
	}

	return meshes, nil
}

func (s *snapshotGenerator) getMatchingDataplanes(ctx context.Context, meshesWithMeshMetrics []string) (map[*v1alpha1.MeshMetricResource][]*core_mesh.DataplaneResource, error) {
	//aggregatedMeshCtxs.MeshContextsByName["some"].Resources
	dataplaneList := &core_mesh.DataplaneResourceList{}
	err := s.resourceManager.List(ctx, dataplaneList)
	if err != nil {
		return nil, errors.Wrap(err, "could not list dpps")
	}

	meshMetricToDataplanes := map[*v1alpha1.MeshMetricResource][]*core_mesh.DataplaneResource{}
	for _, meshName := range meshesWithMeshMetrics {
		meshContext, err := s.meshCache.GetMeshContext(ctx, meshName)
		if err != nil {
			return nil, errors.Wrap(err, "could not get mesh context")
		}

		for _, dp := range dataplaneList.Items {
			matchedPolicies, err := matchers.MatchedPolicies(v1alpha1.MeshMetricType, dp, meshContext.Resources)
			if err != nil {
				return nil, errors.Wrap(err, "error on matching dpp")
			}
			matchedPolicies.SingleItemRules.Rules[0].Conf.()

		}
	}


	return meshMetricToDataplanes, nil
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
