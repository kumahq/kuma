package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
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
	refs []core_model.ResolvedBackendRef,
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
	refs []core_model.ResolvedBackendRef,
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
	OutboundInterface mesh_proto.OutboundInterface
	Resource          *core_model.TypedResourceIdentifier
	Tags              envoy_tags.Tags
	Protocol          core_mesh.Protocol
	ServiceName       string
	BackendRef        common_api.BackendRef
}

func CollectServices(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) []DestinationService {

	var dests []DestinationService
	for _, outbound := range proxy.Outbounds {
		core.Log.Info("CollectServices", "outbound", outbound)
		var destinationService *DestinationService
		if outbound.LegacyOutbound != nil {
			destinationService = collectServiceTagService(outbound, meshCtx)
		} else {
			switch outbound.Resource.ResourceType {
			case core_model.ResourceType(common_api.MeshService):
				destinationService = collectMeshService(outbound, meshCtx)
			case core_model.ResourceType(common_api.MeshExternalService):
				destinationService = collectMeshExternalService(outbound, meshCtx)
			case core_model.ResourceType(common_api.MeshMultiZoneService):
				destinationService = collectMeshMultiZoneService(outbound, meshCtx)
			}
		}
		if destinationService != nil {
			dests = append(dests, *destinationService)
		}
	}
	core.Log.Info("CollectServices", "outbound", dests)
	return dests
}

func collectMeshService(
	outbound *xds_types.Outbound,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	ms, msOk := meshCtx.MeshServiceByIdentifier[pointer.Deref(outbound.Resource).ResourceIdentifier]
	if !msOk {
		// we want to ignore service which is not found. Logging might be excessive here.
		// We don't have other mechanism to bubble up warnings yet.
		return nil
	}
	port, ok := ms.FindPortByName(outbound.Resource.SectionName)
	if !ok {
		return nil
	}
	protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
	if port.AppProtocol != "" {
		protocol = port.AppProtocol
	}
	return &DestinationService{
		OutboundInterface: mesh_proto.OutboundInterface{
			DataplaneIP:   outbound.GetAddress(),
			DataplanePort: outbound.GetPort(),
		},
		Tags:        outbound.LegacyOutbound.GetTags(),
		Resource:    outbound.Resource,
		Protocol:    protocol,
		ServiceName: ms.DestinationName(port.Port),
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: ms.GetMeta().GetName(),
			},
			Port: &port.Port,
		},
	}
}

func collectMeshExternalService(
	outbound *xds_types.Outbound,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	mes, mesOk := meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(outbound.Resource).ResourceIdentifier]
	if !mesOk {
		return nil
	}
	return &DestinationService{
		OutboundInterface: mesh_proto.OutboundInterface{
			DataplaneIP:   outbound.GetAddress(),
			DataplanePort: outbound.GetPort(),
		},
		Tags:        outbound.LegacyOutbound.GetTags(),
		Resource:    outbound.Resource,
		Protocol:    mes.Spec.Match.Protocol,
		ServiceName: mes.DestinationName(uint32(mes.Spec.Match.Port)),
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshExternalService,
				Name: mes.GetMeta().GetName(),
			},
			Port: pointer.To(uint32(mes.Spec.Match.Port)),
		},
	}
}

func collectMeshMultiZoneService(
	outbound *xds_types.Outbound,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	svc, mesOk := meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(outbound.Resource).ResourceIdentifier]
	if !mesOk {
		return nil
	}
	port, ok := svc.FindPortByName(outbound.Resource.SectionName)
	if !ok {
		return nil
	}
	protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
	if port.AppProtocol != "" {
		protocol = port.AppProtocol
	}
	return &DestinationService{
		OutboundInterface: mesh_proto.OutboundInterface{
			DataplaneIP:   outbound.GetAddress(),
			DataplanePort: outbound.GetPort(),
		},
		Tags:        outbound.LegacyOutbound.GetTags(),
		Resource:    outbound.Resource,
		Protocol:    protocol,
		ServiceName: svc.DestinationName(port.Port),
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshMultiZoneService,
				Name: svc.GetMeta().GetName(),
			},
			Port: &port.Port,
		},
	}
}

func collectServiceTagService(
	outbound *xds_types.Outbound,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	serviceName := outbound.LegacyOutbound.GetService()
	return &DestinationService{
		OutboundInterface: mesh_proto.OutboundInterface{
			DataplaneIP:   outbound.GetAddress(),
			DataplanePort: outbound.GetPort(),
		},
		Tags:        outbound.LegacyOutbound.GetTags(),
		Resource:    outbound.Resource,
		Protocol:    meshCtx.GetServiceProtocol(serviceName),
		ServiceName: serviceName,
		BackendRef: common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: serviceName,
				Tags: outbound.LegacyOutbound.GetTags(),
			},
		},
	}
}

func GetServiceAndProtocolFromRef(
	meshCtx xds_context.MeshContext,
	ref core_model.ResolvedBackendRef,
) (string, core_mesh.Protocol, bool) {
	switch {
	case ref.LegacyBackendRef.Kind == common_api.MeshExternalService:
		mes, ok := meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		if !ok {
			return "", "", false
		}
		port := pointer.Deref(ref.LegacyBackendRef.Port)
		service := mes.DestinationName(port)
		protocol := meshCtx.GetServiceProtocol(service)
		return service, protocol, true
	case ref.LegacyBackendRef.Kind == common_api.MeshMultiZoneService:
		ms, ok := meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		if !ok {
			return "", "", false
		}
		port, ok := ms.FindPort(*ref.LegacyBackendRef.Port)
		if !ok {
			return "", "", false
		}
		service := ms.DestinationName(*ref.LegacyBackendRef.Port)
		protocol := port.AppProtocol
		return service, protocol, true
	case ref.LegacyBackendRef.Kind == common_api.MeshService && ref.LegacyBackendRef.ReferencesRealObject():
		ms, ok := meshCtx.MeshServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		if !ok {
			return "", "", false
		}
		port, ok := ms.FindPort(*ref.LegacyBackendRef.Port)
		if !ok {
			return "", "", false
		}
		service := ms.DestinationName(*ref.LegacyBackendRef.Port)
		protocol := port.AppProtocol // todo(jakubdyszkiewicz): do we need to default to TCP or will this be done by MeshService defaulter?
		return service, protocol, true
	default:
		service := ref.LegacyBackendRef.Name
		protocol := meshCtx.GetServiceProtocol(service)
		return service, protocol, true
	}
}

func makeSplit(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []core_model.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	var split []envoy_common.Split

	for _, ref := range refs {
		switch ref.LegacyBackendRef.Kind {
		case common_api.MeshService, common_api.MeshExternalService, common_api.MeshServiceSubset, common_api.MeshMultiZoneService:
		default:
			continue
		}

		var service string
		var protocol core_mesh.Protocol
		if pointer.DerefOr(ref.LegacyBackendRef.Weight, 1) == 0 {
			continue
		}
		service, protocol, ok := GetServiceAndProtocolFromRef(meshCtx, ref)
		if !ok {
			continue
		}
		if _, ok := protocols[protocol]; !ok {
			return nil
		}

		var clusterName string
		switch ref.LegacyBackendRef.Kind {
		case common_api.MeshExternalService:
			clusterName = envoy_names.GetMeshExternalServiceName(ref.LegacyBackendRef.Name) // todo shouldn't this be in destination name?
		case common_api.MeshMultiZoneService:
			clusterName = meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier].DestinationName(*ref.LegacyBackendRef.Port)
		case common_api.MeshService:
			if ref.LegacyBackendRef.Port != nil {
				clusterName = meshCtx.MeshServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier].DestinationName(*ref.LegacyBackendRef.Port)
				break
			}
			fallthrough
		default:
			clusterName, _ = envoy_tags.Tags(ref.LegacyBackendRef.Tags).
				WithTags(mesh_proto.ServiceTag, service).
				DestinationClusterName(nil)
		}

		// The mesh tag is present here if this destination is generated
		// from a cross-mesh MeshGateway listener virtual outbound.
		// It is not part of the service tags.
		if mesh, ok := ref.LegacyBackendRef.Tags[mesh_proto.MeshTag]; ok {
			// The name should be distinct to the service & mesh combination
			clusterName = fmt.Sprintf("%s_%s", clusterName, mesh)
		}

		isExternalService := meshCtx.IsExternalService(service)
		refHash := ref.LegacyBackendRef.Hash()

		if existingClusterName, ok := clusterCache[refHash]; ok {
			// cluster already exists, so adding only split
			split = append(split, plugins_xds.NewSplitBuilder().
				WithClusterName(existingClusterName).
				WithWeight(uint32(pointer.DerefOr(ref.LegacyBackendRef.Weight, 1))).
				WithExternalService(isExternalService).
				Build())
			continue
		}

		clusterCache[refHash] = clusterName

		split = append(split, plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithWeight(uint32(pointer.DerefOr(ref.LegacyBackendRef.Weight, 1))).
			WithExternalService(isExternalService).
			Build())

		clusterBuilder := plugins_xds.NewClusterBuilder().
			WithService(service).
			WithName(clusterName).
			WithTags(envoy_tags.Tags(ref.LegacyBackendRef.Tags).
				WithTags(mesh_proto.ServiceTag, ref.LegacyBackendRef.Name).
				WithoutTags(mesh_proto.MeshTag)).
			WithExternalService(isExternalService)

		if mesh, ok := ref.LegacyBackendRef.Tags[mesh_proto.MeshTag]; ok {
			clusterBuilder.WithMesh(mesh)
		}

		servicesAcc.AddBackendRef(ref, clusterBuilder.Build())
	}

	return split
}
