package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func MakeTCPSplit(
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
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
		proxy,
	)
}

func MakeHTTPSplit(
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
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
		proxy,
	)
}

type DestinationService struct {
	Outbound    *xds_types.Outbound
	Protocol    core_mesh.Protocol
	ServiceName string
}

func (ds *DestinationService) DefaultBackendRef() *resolve.ResolvedBackendRef {
	if r, ok := ds.Outbound.AssociatedServiceResource(); ok {
		return resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
			Resource: &r,
			Weight:   100,
		})
	} else {
		return resolve.NewResolvedBackendRef(&resolve.LegacyBackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: pointer.To(ds.Outbound.LegacyOutbound.GetService()),
				Tags: pointer.To(ds.Outbound.LegacyOutbound.GetTags()),
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
	ms := meshCtx.GetMeshServiceByKRI(pointer.Deref(outbound.Resource))
	if ms == nil {
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
	mes := meshCtx.GetMeshExternalServiceByKRI(pointer.Deref(outbound.Resource))
	if mes == nil {
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
	svc := meshCtx.GetMeshMultiZoneServiceByKRI(pointer.Deref(outbound.Resource))
	if svc == nil {
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
	ref *resolve.RealResourceBackendRef,
	kriStats bool,
) (string, string, core_mesh.Protocol, uint32, bool) {
	switch common_api.TargetRefKind(ref.Resource.ResourceType) {
	case common_api.MeshExternalService:
		mes := meshCtx.GetMeshExternalServiceByKRI(pointer.Deref(ref.Resource))
		if mes == nil {
			return "", "", "", 0, false
		}
		port := uint32(mes.Spec.Match.Port)
		service := kri.From(mes, "").String()
		statName := ""
		if !kriStats {
			statName = mes.DestinationName(port)
		}
		protocol := mes.Spec.Match.Protocol
		return service,statName, protocol, port, true
	case common_api.MeshMultiZoneService:
		ms := meshCtx.GetMeshMultiZoneServiceByKRI(pointer.Deref(ref.Resource))
		if ms == nil {
			return "", "","", 0, false
		}
		port, ok := ms.FindPortByName(ref.Resource.SectionName)
		if !ok {
			return "", "","", 0, false
		}
		service := kri.From(ms,ref.Resource.SectionName).String()
		statName := ""
		if !kriStats {
			statName = ms.DestinationName(port.Port)
		}
		protocol := port.AppProtocol
		return service, statName, protocol, port.Port, true
	case common_api.MeshService:
		ms := meshCtx.GetMeshServiceByKRI(pointer.Deref(ref.Resource))
		if ms == nil {
			return "", "","", 0, false
		}
		port, ok := ms.FindPortByName(ref.Resource.SectionName)
		if !ok {
			return "", "","", 0, false
		}
		service := kri.From(ms,ref.Resource.SectionName).String()
		statName := ""
		if !kriStats {
			statName = ms.DestinationName(port.Port)
		}
		protocol := port.AppProtocol // todo(jakubdyszkiewicz): do we need to default to TCP or will this be done by MeshService defaulter?
		return service, statName, protocol, port.Port, true
	default:
		return "", "","", 0, false
	}
}

func makeSplit(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
) []envoy_common.Split {
	var split []envoy_common.Split

	for _, ref := range refs {
		if ref.ReferencesRealResource() {
			if s := handleRealResources(protocols, clusterCache, servicesAcc, ref.RealResourceBackendRef(), meshCtx, proxy); s != nil {
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
	ref *resolve.RealResourceBackendRef,
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
) envoy_common.Split {
	if ref.Weight == 0 {
		return nil
	}
	useKri := proxy.Metadata.HasFeature(xds_types.FeatureKRIStats)

	service, _, protocol, port, ok := GetServiceProtocolPortFromRef(meshCtx, ref, useKri)
	if !ok {
		return nil
	}
	if _, ok := protocols[protocol]; !ok {
		return nil
	}

	var clusterName string
	var statsName string
	var isExternalService bool

	switch common_api.TargetRefKind(ref.Resource.ResourceType) {
	case common_api.MeshExternalService:
		clusterName = pointer.Deref(ref.Resource).String()
		statsName = meshCtx.GetMeshExternalServiceByKRI(pointer.Deref(ref.Resource)).DestinationName(port)
		if proxy.Metadata.Features.HasFeature(xds_types.FeatureKRIStats) {
			statsName = clusterName
		}
		isExternalService = true
	case common_api.MeshMultiZoneService:
		clusterName = pointer.Deref(ref.Resource).String()
		statsName = meshCtx.GetMeshMultiZoneServiceByKRI(pointer.Deref(ref.Resource)).DestinationName(port)
		if proxy.Metadata.Features.HasFeature(xds_types.FeatureKRIStats) {
			statsName = clusterName
		}
	case common_api.MeshService:
		clusterName = pointer.Deref(ref.Resource).String()
		statsName = meshCtx.GetMeshServiceByKRI(pointer.Deref(ref.Resource)).DestinationName(port)
		if proxy.Metadata.Features.HasFeature(xds_types.FeatureKRIStats) {
			statsName = clusterName
		}
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
		WithStatName(statsName).
		WithTags(envoy_tags.Tags{}.WithTags(mesh_proto.ServiceTag, service)). // todo(lobkovilya): do we need tags for real resource cluster?
		WithExternalService(isExternalService)

	servicesAcc.AddBackendRef(resolve.NewResolvedBackendRef(ref), clusterBuilder.Build())

	return splitTo(clusterName)
}

func handleLegacyBackendRef(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	ref *resolve.LegacyBackendRef,
	meshCtx xds_context.MeshContext,
) envoy_common.Split {
	if ref.Weight != nil && *ref.Weight == 0 {
		return nil
	}

	service := pointer.Deref(ref.Name)
	protocol := meshCtx.GetServiceProtocol(service)
	if _, ok := protocols[protocol]; !ok {
		return nil
	}
	clusterName, _ := envoy_tags.Tags(pointer.Deref(ref.Tags)).
		WithTags(mesh_proto.ServiceTag, service).
		DestinationClusterName(nil)

	// The mesh tag is present here if this destination is generated
	// from a cross-mesh MeshGateway listener virtual outbound.
	// It is not part of the service tags.
	if mesh, ok := pointer.Deref(ref.Tags)[mesh_proto.MeshTag]; ok {
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
		WithTags(envoy_tags.Tags(pointer.Deref(ref.Tags)).
			WithTags(mesh_proto.ServiceTag, pointer.Deref(ref.Name)).
			WithoutTags(mesh_proto.MeshTag)).
		WithExternalService(isExternalService)

	if mesh, ok := pointer.Deref(ref.Tags)[mesh_proto.MeshTag]; ok {
		clusterBuilder.WithMesh(mesh)
	}

	servicesAcc.AddBackendRef(resolve.NewResolvedBackendRef(ref), clusterBuilder.Build())
	return split
}
