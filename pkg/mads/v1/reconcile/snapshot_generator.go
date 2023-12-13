package reconcile

import (
	"context"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/mads/v1/meshmetrics"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/pkg/errors"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/mads/generator"
	mads_v1_cache "github.com/kumahq/kuma/pkg/mads/v1/cache"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

var log = core.Log.WithName("mads").WithName("v1").WithName("reconcile")

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

func (s *snapshotGenerator) GenerateSnapshot(ctx context.Context, node *envoy_core.Node) (util_xds_v3.Snapshot, error) {
	meshMetrics, err := s.getMeshMetricsWithCorrespondingPrometheusBackend(ctx, node.Id)
	if err != nil {
		return nil, err
	}

	meshes, err := s.getMeshesWithPrometheusEnabled(ctx)
	if err != nil {
		return nil, err
	}

	if len(meshes) > 0 && len(meshMetrics) > 0 {
		log.Info("it is not supported to use both MeshMetrics policy and 'metrics' under Mesh resource. If migrating please remove the 'metrics' section and apply an equivalent MeshMetrics resource")
	}

	var resources []*core_xds.Resource
	if len(meshMetrics) == 0 {
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
		dataplanes, err := s.getMatchingDataplanes(ctx, meshMetrics)
		if err != nil {
			return nil, err
		}

		resources, err = meshmetrics.Generate(dataplanes)
		if err != nil {
			return nil, err
		}
	}

	return mads_v1_cache.NewSnapshot("", core_xds.ResourceList(resources).ToIndex()), nil
}

func (s *snapshotGenerator) getMeshMetricsWithCorrespondingPrometheusBackend(ctx context.Context, clientId string) ([]*v1alpha1.MeshMetricResource, error) {
	meshMetricsList := v1alpha1.MeshMetricResourceList{}
	if err := s.resourceManager.List(ctx, &meshMetricsList); err != nil {
		return nil, err
	}

	meshMetrics := make([]*v1alpha1.MeshMetricResource, 0)
	for _, meshMetric := range meshMetricsList.Items {
		for _, backend := range *meshMetric.Spec.Default.Backends { // can backends be nil?
			// match against client ID or fallback to "" when specified by user
			if backend.Type == v1alpha1.PrometheusBackendType && (backend.Prometheus.ClientId == nil || *backend.Prometheus.ClientId == clientId || *backend.Prometheus.ClientId == "") {
				meshMetrics = append(meshMetrics, meshMetric)
			}
		}
	}

	return meshMetrics, nil
}

func (s *snapshotGenerator) getMatchingDataplanes(ctx context.Context, meshMetrics []*v1alpha1.MeshMetricResource) (map[*v1alpha1.MeshMetricResource][]*core_mesh.DataplaneResource, error) {
	meshMetricToDataplanes := map[*v1alpha1.MeshMetricResource][]*core_mesh.DataplaneResource{}

	dataplaneList := &core_mesh.DataplaneResourceList{}
	err := s.resourceManager.List(ctx, dataplaneList)
	if err != nil {
		return nil, errors.Wrap(err, "could not list dpps")
	}


	for _, meshMetric := range meshMetrics {
		var filteredDataplaneList []*core_mesh.DataplaneResource
		for _, dpp := range dataplaneList.Items {
			matched, err := matchers.PolicyMatches(meshMetric, dpp, xds_context.NewResources()) // what are the dependant resources? do I need them?
			if err != nil {
				return nil, errors.Wrap(err, "errored out on matching dpp")
			}
			if matched {
				filteredDataplaneList = append(filteredDataplaneList, dpp)
			}
		}
		meshMetricToDataplanes[meshMetric] = filteredDataplaneList
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
