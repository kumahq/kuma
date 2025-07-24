package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
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
	return makeSplits(
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
	return makeSplits(
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
	Outbound            *xds_types.Outbound
	Protocol            core_mesh.Protocol
	KumaServiceTagValue string
}

// MaybeResolveKRIWithFallback returns the identifier for this DestinationService.
// If provided condition is met and the Outbound has an associated real resource,
// the identifier is derived from that resource (KRI). Otherwise, the given fallback
// is returned
func (ds *DestinationService) MaybeResolveKRIWithFallback(condition bool, fallback string) string {
	if condition && ds.Outbound != nil {
		if id, ok := ds.Outbound.AssociatedServiceResource(); ok {
			return id.String()
		}
	}
	return fallback
}

// ResolveKRIWithFallback returns the identifier for this DestinationService.
// If the Outbound has an associated real resource, the identifier is derived
// from it (KRI). Otherwise, the given fallback is returned
func (ds *DestinationService) ResolveKRIWithFallback(fallback string) string {
	return ds.MaybeResolveKRIWithFallback(true, fallback)
}

// MaybeResolveKRI returns the identifier for this DestinationService if the
// condition is met and the Outbound has an associated real resource. Otherwise,
// an empty string is returned
func (ds *DestinationService) MaybeResolveKRI(condition bool) string {
	return ds.MaybeResolveKRIWithFallback(condition, "")
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

// CollectServices builds a slice of DestinationService from proxy.Outbounds
//
// It handles two types of outbounds:
// - Legacy outbounds: resolved using service name and protocol from mesh context
// - Real-resource outbounds: resolved by matching KRI identifier and port name
//
// Skips outbounds that are incomplete or invalid:
// - nil entries
// - real-resource outbounds with missing resource reference
// - no service found for the given KRI
// - no port matching the SectionName
//
// When protocol is unset, it defaults to TCP
func CollectServices(proxy *core_xds.Proxy, meshCtx xds_context.MeshContext) []DestinationService {
	var result []DestinationService

	for _, outbound := range proxy.Outbounds {
		if outbound == nil {
			continue
		}

		if lo := outbound.LegacyOutbound; lo != nil {
			result = append(
				result,
				DestinationService{
					Outbound:            outbound,
					Protocol:            meshCtx.GetServiceProtocol(lo.GetService()),
					KumaServiceTagValue: lo.GetService(),
				},
			)

			continue
		}

		var svc core.Destination
		var port core.Port
		var protocol core_mesh.Protocol
		var ok bool

		// ignore outbound when no service matches the KRI identifier
		// TODO: Add a clear way to pass warnings up when needed. Right now
		//  we skip logging to avoid too much noise, and there’s no system
		//  for handling warnings yet
		if svc = meshCtx.GetServiceByKRI(pointer.Deref(outbound.Resource)); svc == nil {
			continue
		}

		// skip outbounds when no port matches SectionName
		if port, ok = svc.FindPortByName(outbound.Resource.SectionName); !ok {
			continue
		}

		// determine protocol, default to TCP if unspecified
		if protocol = port.GetProtocol(); protocol == "" {
			protocol = core_mesh.ProtocolTCP
		}

		result = append(
			result,
			DestinationService{
				Outbound:            outbound,
				Protocol:            protocol,
				KumaServiceTagValue: destinationname.MustResolve(false, svc, port),
			},
		)
	}

	return result
}

func DestinationPortFromRef(
	meshCtx xds_context.MeshContext,
	ref *resolve.RealResourceBackendRef,
) (core.Destination, core.Port, bool) {
	var dest core.Destination
	var port core.Port
	var ok bool

	if dest = meshCtx.GetServiceByKRI(pointer.Deref(ref.Resource)); dest == nil {
		return dest, port, false
	}

	if port, ok = dest.FindPortByName(ref.Resource.SectionName); !ok {
		return dest, port, false
	}

	return dest, port, true
}

func makeSplits(
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	var result []envoy_common.Split

	splitFromRef := func(ref resolve.ResolvedBackendRef) envoy_common.Split {
		if ref.ReferencesRealResource() {
			return handleRealResources(protocols, clusterCache, servicesAcc, ref.RealResourceBackendRef(), meshCtx)
		}

		return handleLegacyBackendRef(protocols, clusterCache, servicesAcc, ref.LegacyBackendRef(), meshCtx)
	}

	for _, ref := range refs {
		result = append(result, splitFromRef(ref))
	}

	// return only non-nil splits
	return util_slices.Filter(
		result,
		func(s envoy_common.Split) bool {
			return s != nil
		},
	)
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

	dest, port, ok := DestinationPortFromRef(meshCtx, ref)
	if !ok {
		return nil
	}

	if _, ok := protocols[port.GetProtocol()]; !ok {
		return nil
	}

	service := destinationname.MustResolve(false, dest, port)
	clusterName := service

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
