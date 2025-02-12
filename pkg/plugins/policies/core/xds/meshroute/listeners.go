package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
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
	Outbound    *xds_types.Outbound
	Protocol    core_mesh.Protocol
	ServiceName string
}

func (ds *DestinationService) DefaultBackendRef() *core_model.ResolvedBackendRef {
	if ds.Outbound.Resource != nil {
		return core_model.NewResolvedBackendRef(&core_model.RealResourceBackendRef{
			Resource: ds.Outbound.Resource,
			Weight:   100,
		})
	} else {
		return core_model.NewResolvedBackendRef(&core_model.LegacyBackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: ds.Outbound.LegacyOutbound.GetService(),
				Tags: ds.Outbound.LegacyOutbound.GetTags(),
			},
			Weight: pointer.To(uint(100)),
		})
	}
}

func CollectServices(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) []DestinationService {
	var dests []DestinationService
	for _, outbound := range proxy.Outbounds {
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
		Outbound:    outbound,
		Protocol:    protocol,
		ServiceName: ms.DestinationName(port.Port),
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
	protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
	if mes.Spec.Match.Protocol != "" {
		protocol = mes.Spec.Match.Protocol
	}
	return &DestinationService{
		Outbound:    outbound,
		Protocol:    protocol,
		ServiceName: mes.DestinationName(uint32(mes.Spec.Match.Port)),
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
		Outbound:    outbound,
		Protocol:    protocol,
		ServiceName: svc.DestinationName(port.Port),
	}
}

func collectServiceTagService(
	outbound *xds_types.Outbound,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	serviceName := outbound.LegacyOutbound.GetService()
	return &DestinationService{
		Outbound:    outbound,
		Protocol:    meshCtx.GetServiceProtocol(serviceName),
		ServiceName: serviceName,
	}
}

func GetServiceProtocolPortFromRef(
	meshCtx xds_context.MeshContext,
	ref *core_model.RealResourceBackendRef,
) (string, core_mesh.Protocol, uint32, bool) {
	switch common_api.TargetRefKind(ref.Resource.ResourceType) {
	case common_api.MeshExternalService:
		mes, ok := meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		if !ok {
			return "", "", 0, false
		}
		port := uint32(mes.Spec.Match.Port)
		service := mes.DestinationName(port)
		protocol := mes.Spec.Match.Protocol
		return service, protocol, port, true
	case common_api.MeshMultiZoneService:
		ms, ok := meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		if !ok {
			return "", "", 0, false
		}
		port, ok := ms.FindPortByName(ref.Resource.SectionName)
		if !ok {
			return "", "", 0, false
		}
		service := ms.DestinationName(port.Port)
		protocol := port.AppProtocol
		return service, protocol, port.Port, true
	case common_api.MeshService:
		ms, ok := meshCtx.MeshServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		if !ok {
			return "", "", 0, false
		}
		port, ok := ms.FindPortByName(ref.Resource.SectionName)
		if !ok {
			return "", "", 0, false
		}
		service := ms.DestinationName(port.Port)
		protocol := port.AppProtocol // todo(jakubdyszkiewicz): do we need to default to TCP or will this be done by MeshService defaulter?
		return service, protocol, port.Port, true
	default:
		return "", "", 0, false
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
		if ref.ReferencesRealResource() {
			if s := handleRealResources(protocols, clusterCache, servicesAcc, ref.RealResourceBackendRef(), meshCtx); s != nil {
				split = append(split, s)
			}
		} else {
			if s := handleLegacyBackendRef(protocols, clusterCache, servicesAcc, ref.LegacyBackendRef(), meshCtx); s != nil {
				split = append(split, s)
			}
		}
	}

	return split
}

func handleRealResources(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	ref *core_model.RealResourceBackendRef,
	meshCtx xds_context.MeshContext,
) envoy_common.Split {
	if ref.Weight == 0 {
		return nil
	}

	service, protocol, port, ok := GetServiceProtocolPortFromRef(meshCtx, ref)
	if !ok {
		return nil
	}
	if _, ok := protocols[protocol]; !ok {
		return nil
	}

	var clusterName string
	var isExternalService bool

	switch common_api.TargetRefKind(ref.Resource.ResourceType) {
	case common_api.MeshExternalService:
		clusterName = meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier].DestinationName(port)
		isExternalService = true
	case common_api.MeshMultiZoneService:
		clusterName = meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier].DestinationName(port)
	case common_api.MeshService:
		clusterName = meshCtx.MeshServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier].DestinationName(port)
	}

	// todo(lobkovilya): instead of computing hash we should use ResourceIdentifier as a key in clusterCache (or maybe we don't need clusterCache)
	refHash := common_api.BackendRefHash(ref.Resource.String())

	splitTo := func(clusterName string) envoy_common.Split {
		return plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithWeight(uint32(ref.Weight)).
			WithExternalService(isExternalService).
			Build()
	}

	if existingClusterName, ok := clusterCache[refHash]; ok {
		// cluster already exists, so adding only split
		return splitTo(existingClusterName)
	}

	clusterCache[refHash] = clusterName

	clusterBuilder := plugins_xds.NewClusterBuilder().
		WithService(service).
		WithName(clusterName).
		WithTags(envoy_tags.Tags{}.WithTags(mesh_proto.ServiceTag, service)). // todo(lobkovilya): do we need tags for real resource cluster?
		WithExternalService(isExternalService)

	servicesAcc.AddBackendRef(core_model.NewResolvedBackendRef(ref), clusterBuilder.Build())

	return splitTo(clusterName)
}

func handleLegacyBackendRef(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	ref *core_model.LegacyBackendRef,
	meshCtx xds_context.MeshContext,
) envoy_common.Split {
	if ref.Weight != nil && *ref.Weight == 0 {
		return nil
	}

	service := ref.Name
	protocol := meshCtx.GetServiceProtocol(service)
	if _, ok := protocols[protocol]; !ok {
		return nil
	}
	clusterName, _ := envoy_tags.Tags(ref.Tags).
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
	refHash := common_api.BackendRef(*ref).Hash()

	if existingClusterName, ok := clusterCache[refHash]; ok {
		// cluster already exists, so adding only split
		return plugins_xds.NewSplitBuilder().
			WithClusterName(existingClusterName).
			WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
			WithExternalService(isExternalService).
			Build()
	}

	clusterCache[refHash] = clusterName

	split := plugins_xds.NewSplitBuilder().
		WithClusterName(clusterName).
		WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
		WithExternalService(isExternalService).
		Build()

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

	servicesAcc.AddBackendRef(core_model.NewResolvedBackendRef(ref), clusterBuilder.Build())
	return split
}
