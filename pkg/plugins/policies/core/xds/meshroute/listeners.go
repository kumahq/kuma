package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func MakeTCPSplit(
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplit(
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolUnknown: {},
			core_mesh.ProtocolKafka:   {},
			core_mesh.ProtocolTCP:     {},
			core_mesh.ProtocolHTTP:    {},
			core_mesh.ProtocolHTTP2:   {},
			core_mesh.ProtocolGRPC:    {},
		},
		clusterCache,
		servicesAcc,
		refs,
		meshCtx,
	)
}

func MakeHTTPSplit(
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplit(
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolHTTP:  {},
			core_mesh.ProtocolHTTP2: {},
			core_mesh.ProtocolGRPC:  {},
		},
		clusterCache,
		servicesAcc,
		refs,
		meshCtx,
	)
}

type DestinationService struct {
	Outbound    mesh_proto.OutboundInterface
	Protocol    core_mesh.Protocol
	ServiceName string
	BackendRef  common_api.BackendRef
}

func CollectServices(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) []DestinationService {
	var dests []DestinationService
	for _, outbound := range proxy.Dataplane.Spec.GetNetworking().GetOutbounds() {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		if outbound.BackendRef != nil {
			ms, ok := meshCtx.MeshServiceByName[outbound.BackendRef.Name]
			if !ok {
				// we want to ignore service which is not found. Logging might be excessive here.
				// We don't have other mechanism to bubble up warnings yet.
				continue
			}
			port, ok := ms.FindPort(outbound.BackendRef.Port)
			if !ok {
				continue
			}
			protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
			if port.Protocol != "" {
				protocol = port.Protocol
			}
			dests = append(dests, DestinationService{
				Outbound:    oface,
				Protocol:    protocol,
				ServiceName: ms.DestinationName(outbound.BackendRef.Port),
				BackendRef: common_api.BackendRef{
					TargetRef: common_api.TargetRef{
						Kind: common_api.MeshService,
						Name: ms.GetMeta().GetName(),
					},
					Port: &port.Port,
				},
			})
		} else {
			serviceName := outbound.GetService()
			dests = append(dests, DestinationService{
				Outbound:    oface,
				Protocol:    meshCtx.GetServiceProtocol(serviceName),
				ServiceName: serviceName,
				BackendRef: common_api.BackendRef{
					TargetRef: common_api.TargetRef{
						Kind: common_api.MeshService,
						Name: serviceName,
						Tags: outbound.GetTags(),
					},
				},
			})
		}
	}
	return dests
}

func makeSplit(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
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

		var service string
		var protocol core_mesh.Protocol
		if pointer.DerefOr(ref.Weight, 1) == 0 {
			continue
		}
		if ref.Port != nil { // in this case, reference real MeshService instead of kuma.io/service tag
			ms, ok := meshCtx.MeshServiceByName[ref.Name]
			if !ok {
				continue
			}
			port, ok := ms.FindPort(*ref.Port)
			if !ok {
				continue
			}
			service = ms.DestinationName(*ref.Port)
			protocol = port.Protocol // todo(jakubdyszkiewicz): do we need to default to TCP or will this be done by MeshService defaulter?
		} else {
			service = ref.Name
			protocol = meshCtx.GetServiceProtocol(service)
		}
		if _, ok := protocols[protocol]; !ok {
			return nil
		}

		clusterName, _ := tags.Tags(ref.Tags).
			WithTags(mesh_proto.ServiceTag, service).
			DestinationClusterName(nil)

		// The mesh tag is present here if this destination is generated
		// from a cross-mesh MeshGateway listener virtual outbound.
		// It is not part of the service tags.
		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			// The name should be distinct to the service & mesh combination
			clusterName = fmt.Sprintf("%s_%s", clusterName, mesh)
		}

		isExternalService := meshCtx.IsExternalService(service)
		refHash := ref.Hash()

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
