package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds/meshroute"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func getClusters(
	routing core_xds.Routing,
	clusterCache map[common_api.TargetRefHash]struct{},
	sc *meshroute_xds.SplitCounter,
	servicesAccumulator envoy_common.ServicesAccumulator,
	backendRefs []common_api.BackendRef,
) []envoy_common.Cluster {
	var clusters []envoy_common.Cluster

	for _, ref := range backendRefs {
		switch ref.Kind {
		case common_api.MeshService, common_api.MeshServiceSubset:
		default:
			continue
		}

		serviceName := ref.Name
		if pointer.DerefOr(ref.Weight, 1) == 0 {
			continue
		}

		clusterName := meshroute_xds.GetClusterName(ref.Name, ref.Tags, sc)
		isExternalService := plugins_xds.HasExternalService(routing, ref.Name)
		refHash := ref.TargetRef.Hash()

		clusterBuilder := plugins_xds.NewClusterBuilder().
			WithService(serviceName).
			WithName(clusterName).
			WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
			WithTags(envoy_tags.Tags(ref.Tags).
				WithTags(mesh_proto.ServiceTag, ref.Name).
				WithoutTags(mesh_proto.MeshTag)).
			WithExternalService(isExternalService)

		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			clusterBuilder.WithMesh(mesh)
		}

		cluster := clusterBuilder.Build()

		clusters = append(clusters, cluster)

		// cluster doesn't exist yet, so we should create a cache entry, and
		// add it to the service accumulator
		if _, ok := clusterCache[refHash]; !ok {
			clusterCache[refHash] = struct{}{}

			servicesAccumulator.Add(cluster)
		}
	}

	return clusters
}
