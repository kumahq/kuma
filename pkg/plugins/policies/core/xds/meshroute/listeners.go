package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
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
	Outbound      mesh_proto.OutboundInterface
	Protocol      core_mesh.Protocol
	ServiceName   string
	BackendRef    common_api.BackendRef
	OwnerResource *core_rules.UniqueResourceIdentifier
}

func CollectServices(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) []DestinationService {
	var dests []DestinationService
	for _, outbound := range proxy.Outbounds {
		var destinationService *DestinationService
		switch outbound.LegacyOutbound.GetBackendRef().GetKind() {
		case string(common_api.MeshService):
			destinationService = collectMeshService(outbound.LegacyOutbound, proxy, meshCtx)
		case string(common_api.MeshExternalService):
			destinationService = collectMeshExternalService(outbound.LegacyOutbound, proxy, meshCtx)
		case string(common_api.MeshMultiZoneService):
			destinationService = collectMeshMultiZoneService(outbound.LegacyOutbound, proxy, meshCtx)
		default:
			destinationService = collectServiceTagService(outbound.LegacyOutbound, proxy, meshCtx)
		}
		if destinationService != nil {
			dests = append(dests, *destinationService)
		}
	}
	return dests
}

func collectMeshService(
	outbound *mesh_proto.Dataplane_Networking_Outbound,
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	ms, msOk := meshCtx.MeshServiceByName[outbound.BackendRef.Name]
	if !msOk {
		// we want to ignore service which is not found. Logging might be excessive here.
		// We don't have other mechanism to bubble up warnings yet.
		return nil
	}
	port, ok := ms.FindPort(outbound.BackendRef.Port)
	if !ok {
		return nil
	}
	protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
	if port.AppProtocol != "" {
		protocol = port.AppProtocol
	}
	return &DestinationService{
		Outbound:    proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound),
		Protocol:    protocol,
		ServiceName: ms.DestinationName(outbound.BackendRef.Port),
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: ms.GetMeta().GetName(),
			},
			Port: &port.Port,
		},
		OwnerResource: pointer.To(core_rules.UniqueKey(ms, port.Name)),
	}
}

func collectMeshExternalService(
	outbound *mesh_proto.Dataplane_Networking_Outbound,
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	mes, mesOk := meshCtx.MeshExternalServiceByName[outbound.BackendRef.Name]
	if !mesOk {
		return nil
	}
	return &DestinationService{
		Outbound:    proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound),
		Protocol:    core_mesh.Protocol(mes.Spec.Match.Protocol),
		ServiceName: mes.DestinationName(outbound.BackendRef.Port),
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshExternalService,
				Name: mes.GetMeta().GetName(),
			},
			Port: pointer.To(uint32(mes.Spec.Match.Port)),
		},
		OwnerResource: pointer.To(core_rules.UniqueKey(mes, "")),
	}
}

func collectMeshMultiZoneService(
	outbound *mesh_proto.Dataplane_Networking_Outbound,
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	svc, mesOk := meshCtx.MeshMultiZoneServiceByName[outbound.BackendRef.Name]
	if !mesOk {
		return nil
	}
	port, ok := svc.FindPort(outbound.BackendRef.Port)
	if !ok {
		return nil
	}
	protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
	if port.AppProtocol != "" {
		protocol = port.AppProtocol
	}
	return &DestinationService{
		Outbound:    proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound),
		Protocol:    protocol,
		ServiceName: svc.DestinationName(outbound.BackendRef.Port),
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshMultiZoneService,
				Name: svc.GetMeta().GetName(),
			},
			Port: &port.Port,
		},
		OwnerResource: pointer.To(core_rules.UniqueKey(svc, port.Name)),
	}
}

func collectServiceTagService(
	outbound *mesh_proto.Dataplane_Networking_Outbound,
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	serviceName := outbound.GetService()
	return &DestinationService{
		Outbound:    proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound),
		Protocol:    meshCtx.GetServiceProtocol(serviceName),
		ServiceName: serviceName,
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: serviceName,
				Tags: outbound.GetTags(),
			},
		},
	}
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
		case common_api.MeshService, common_api.MeshExternalService, common_api.MeshServiceSubset, common_api.MeshMultiZoneService:
		default:
			continue
		}

		var service string
		var protocol core_mesh.Protocol
		if pointer.DerefOr(ref.Weight, 1) == 0 {
			continue
		}
		switch {
		case ref.Kind == common_api.MeshExternalService:
			mes, ok := meshCtx.MeshExternalServiceByName[ref.Name]
			if !ok {
				continue
			}
			port := pointer.Deref(ref.Port)
			service = mes.DestinationName(port)
			protocol = meshCtx.GetServiceProtocol(service)
		case ref.Kind == common_api.MeshMultiZoneService:
			ms, ok := meshCtx.MeshMultiZoneServiceByName[ref.Name]
			if !ok {
				continue
			}
			port, ok := ms.FindPort(*ref.Port)
			if !ok {
				continue
			}
			service = ms.DestinationName(*ref.Port)
			protocol = port.AppProtocol
		case ref.Kind == common_api.MeshService && ref.ReferencesRealObject():
			ms, ok := meshCtx.MeshServiceByName[ref.Name]
			if !ok {
				continue
			}
			port, ok := ms.FindPort(*ref.Port)
			if !ok {
				continue
			}
			service = ms.DestinationName(*ref.Port)
			protocol = port.AppProtocol // todo(jakubdyszkiewicz): do we need to default to TCP or will this be done by MeshService defaulter?
		default:
			service = ref.Name
			protocol = meshCtx.GetServiceProtocol(service)
		}
		if _, ok := protocols[protocol]; !ok {
			return nil
		}

		var clusterName string
		switch ref.Kind {
		case common_api.MeshExternalService:
			clusterName = envoy_names.GetMeshExternalServiceName(ref.Name) // todo shouldn't this be in destination name?
		case common_api.MeshMultiZoneService:
			clusterName = meshCtx.MeshMultiZoneServiceByName[ref.Name].DestinationName(*ref.Port)
		default:
			clusterName, _ = envoy_tags.Tags(ref.Tags).
				WithTags(mesh_proto.ServiceTag, service).
				DestinationClusterName(nil)
		}

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

		servicesAcc.AddBackendRef(ref, clusterBuilder.Build())
	}

	return split
}
