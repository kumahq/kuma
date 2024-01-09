package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func MakeTCPSplit(
	clusterCache map[common_api.TargetRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplit(
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolUnknown: {},
			core_mesh.ProtocolTCP:     {},
			core_mesh.ProtocolHTTP:    {},
			core_mesh.ProtocolHTTP2:   {},
		},
		clusterCache,
		servicesAcc,
		refs,
		meshCtx,
	)
}

func MakeHTTPSplit(
	clusterCache map[common_api.TargetRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplit(
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolHTTP:  {},
			core_mesh.ProtocolHTTP2: {},
		},
		clusterCache,
		servicesAcc,
		refs,
		meshCtx,
	)
}

func makeSplit(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.TargetRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	var split []envoy_common.Split

	for _, ref := range refs {
		switch ref.Kind {
		case common_api.MeshService, common_api.MeshServiceSubset:
		default:
			continue
		}

		service := ref.Name
		if pointer.DerefOr(ref.Weight, 1) == 0 {
			continue
		}

		protocol := meshCtx.GetServiceProtocol(service)
		if _, ok := protocols[protocol]; !ok {
			return nil
		}

		clusterName, _ := tags.Tags(ref.Tags).
			WithTags(mesh_proto.ServiceTag, ref.Name).
			DestinationClusterName(nil)

		// The mesh tag is present here if this destination is generated
		// from a cross-mesh MeshGateway listener virtual outbound.
		// It is not part of the service tags.
		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			// The name should be distinct to the service & mesh combination
			clusterName = fmt.Sprintf("%s_%s", clusterName, mesh)
		}

		isExternalService := meshCtx.IsExternalService(service)
		refHash := ref.TargetRef.Hash()

		if existingClusterName, ok := clusterCache[refHash]; ok {
			// cluster already exists, so adding only split
			split = append(split, plugins_xds.NewSplitBuilder().
				WithClusterName(existingClusterName).
				WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
				WithExternalService(isExternalService).
				Build())
			continue
		}

		clusterCache[refHash] = clusterName

		split = append(split, plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
			WithExternalService(isExternalService).
			Build())

		clusterBuilder := plugins_xds.NewClusterBuilder().
			WithService(service).
			WithName(clusterName).
			WithTags(envoy_tags.Tags(ref.Tags).
				WithTags(mesh_proto.ServiceTag, ref.Name).
				WithoutTags(mesh_proto.MeshTag)).
			WithExternalService(isExternalService)

		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			clusterBuilder.WithMesh(mesh)
		}

		servicesAcc.Add(clusterBuilder.Build())
	}

	return split
}
