package reconcile

import (
	"context"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	mads_v1 "github.com/kumahq/kuma/v3/pkg/mads/v1"
	meshmetrics_generator "github.com/kumahq/kuma/v3/pkg/mads/v1/generator"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_xds_v3 "github.com/kumahq/kuma/v3/pkg/util/xds/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/mesh"
)

func NewSnapshotGenerator(resourceManager core_manager.ReadOnlyResourceManager, meshCache *mesh.Cache, inboundTagsDisabled bool) *SnapshotGenerator {
	return &SnapshotGenerator{
		resourceManager:     resourceManager,
		meshCache:           meshCache,
		inboundTagsDisabled: inboundTagsDisabled,
	}
}

type SnapshotGenerator struct {
	resourceManager     core_manager.ReadOnlyResourceManager
	meshCache           *mesh.Cache
	inboundTagsDisabled bool
}

func (s *SnapshotGenerator) GenerateSnapshot(ctx context.Context) (map[string]envoy_cache.ResourceSnapshot, error) {
	meshesWithMeshMetrics, err := s.getMeshesWithMeshMetrics(ctx)
	if err != nil {
		return nil, err
	}

	resourcesPerClientId := map[string]envoy_cache.ResourceSnapshot{}
	if len(meshesWithMeshMetrics) == 0 {
		// keep an empty snapshot for the default client so MADS subscribers
		// receive a valid (empty) response when no MeshMetric is configured
		resourcesPerClientId[meshmetrics_generator.DefaultKumaClientId] = createSnapshot(nil)
	} else {
		for clientId, meshes := range meshesWithMeshMetrics {
			meshMetricConfToDataplanes, err := s.getMatchingDataplanes(ctx, meshes)
			if err != nil {
				return nil, err
			}

			resources, err := meshmetrics_generator.Generate(meshMetricConfToDataplanes, clientId, s.inboundTagsDisabled)
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

func createSnapshot(resources []*core_xds.Resource) envoy_cache.ResourceSnapshot {
	return util_xds_v3.NewSingleTypeSnapshot(core.NewUUID(), mads_v1.MonitoringAssignmentType, core_xds.ResourceList(resources).Payloads())
}
