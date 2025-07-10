package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
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
	refs []resolve.ResolvedBackendRef,
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
			serviceName := outbound.LegacyOutbound.GetService()
			destinationService = &DestinationService{
				Outbound:    outbound,
				Protocol:    meshCtx.GetServiceProtocol(serviceName),
				ServiceName: serviceName,
			}
		} else {
			destinationService = createServiceFromRealResource(outbound, meshCtx)
		}
		if destinationService != nil {
			dests = append(dests, *destinationService)
		}
	}
	return dests
}

func createServiceFromRealResource(
	outbound *xds_types.Outbound,
	meshCtx xds_context.MeshContext,
) *DestinationService {
	service := meshCtx.GetServiceByKRI(pointer.Deref(outbound.Resource))
	if service == nil {
		// we want to ignore service which is not found. Logging might be excessive here.
		// We don't have another mechanism to bubble up warnings yet.
		return nil
	}
	port, ok := service.FindPortByName(outbound.Resource.SectionName)
	if !ok {
		return nil
	}
	protocol := core_mesh.Protocol(core_mesh.ProtocolTCP)
	if port.GetProtocol() != "" {
		protocol = port.GetProtocol()
	}
	return &DestinationService{
		Outbound:    outbound,
		Protocol:    protocol,
		ServiceName: service.DestinationName(port.GetValue()),
	}
}

func GetServiceProtocolPortFromRef(
	meshCtx xds_context.MeshContext,
	ref *resolve.RealResourceBackendRef,
) (string, core_mesh.Protocol, int32, bool) {
	dest := meshCtx.GetServiceByKRI(pointer.Deref(ref.Resource))
	if dest == nil {
		return "", "", 0, false
	}
	port, ok := dest.FindPortByName(ref.Resource.SectionName)
	if !ok {
		return "", "", 0, false
	}
	service := dest.DestinationName(port.GetValue())
	protocol := port.GetProtocol()
	return service, protocol, port.GetValue(), true
}

func makeSplit(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
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
	ref *resolve.RealResourceBackendRef,
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

	dest := meshCtx.GetServiceByKRI(pointer.Deref(ref.Resource))
	clusterName := dest.DestinationName(port)
	isExternalService := ref.Resource.ResourceType == meshexternalservice_api.MeshExternalServiceType

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
