package meshroute

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

func MakeTCPSplit(
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	return makeSplits(
		map[core_meta.Protocol]struct{}{
			core_meta.ProtocolUnknown: {},
			core_meta.ProtocolTCP:     {},
			core_meta.ProtocolHTTP:    {},
			core_meta.ProtocolHTTP2:   {},
			core_meta.ProtocolGRPC:    {},
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
		map[core_meta.Protocol]struct{}{
			core_meta.ProtocolHTTP:  {},
			core_meta.ProtocolHTTP2: {},
			core_meta.ProtocolGRPC:  {},
		},
		clusterCache,
		servicesAcc,
		refs,
		meshCtx,
	)
}

type DestinationService struct {
	Outbound            *xds_types.Outbound
	Protocol            core_meta.Protocol
	DestinationResource string
}

// OutboundListenerTags returns the outbound listener's io.kuma.tags: the
// destination KRI under kuma.io/unified-name.
func (ds *DestinationService) OutboundListenerTags() map[string]string {
	if ds.Outbound == nil {
		return nil
	}
	if id, ok := ds.Outbound.AssociatedServiceResource(); ok {
		return map[string]string{mesh_proto.UnifiedNameTag: id.String()}
	}
	return nil
}

func (ds *DestinationService) DefaultBackendRef() *resolve.ResolvedBackendRef {
	return resolve.NewResolvedBackendRef(&resolve.RealResourceBackendRef{
		Resource: ds.Outbound.Resource,
		Weight:   100,
	})
}

// CollectServices builds a slice of DestinationService from proxy.Outbounds
//
// Every outbound is resolved by matching its KRI identifier and port name.
//
// Skips outbounds that are incomplete or invalid:
// - nil entries
// - outbounds with missing resource reference
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

		var svc core.Destination
		var port core.Port
		var protocol core_meta.Protocol
		var ok bool

		// ignore outbound when no service matches the KRI identifier
		// TODO: Add a clear way to pass warnings up when needed. Right now
		//  we skip logging to avoid too much noise, and there’s no system
		//  for handling warnings yet
		if svc = meshCtx.GetServiceByKRI(outbound.Resource); svc == nil {
			continue
		}

		if outbound.Resource.SectionName == "" {
			ports := svc.GetPorts()
			if len(ports) != 1 {
				continue
			}
			port = ports[0]
		} else {
			// skip outbounds when no port matches SectionName
			if port, ok = svc.FindPortByName(outbound.Resource.SectionName); !ok {
				continue
			}
		}

		// determine protocol, default to TCP if unspecified
		if protocol = port.GetProtocol(); protocol == "" {
			protocol = core_meta.ProtocolTCP
		}

		result = append(
			result,
			DestinationService{
				Outbound:            outbound,
				Protocol:            protocol,
				DestinationResource: outbound.Resource.String(),
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

	if dest = meshCtx.GetServiceByKRI(ref.Resource); dest == nil {
		return nil, nil, false
	}

	if ref.Resource.SectionName == "" {
		ports := dest.GetPorts()
		if len(ports) != 1 {
			return nil, nil, false
		}
		return dest, ports[0], true
	}

	if port, ok = dest.FindPortByName(ref.Resource.SectionName); !ok {
		return nil, nil, false
	}

	return dest, port, true
}

func makeSplits(
	protocols map[core_meta.Protocol]struct{},
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []resolve.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) []envoy_common.Split {
	var result []envoy_common.Split

	for _, ref := range refs {
		// A ref that resolves to no real resource gets no cluster from
		// GenerateClusters, so splitting towards it would leave the route
		// pointing at a cluster that is never emitted.
		if !ref.ReferencesRealResource() {
			continue
		}
		if split := handleRealResources(protocols, clusterCache, servicesAcc, ref.RealResourceBackendRef(), meshCtx); split != nil {
			result = append(result, split)
		}
	}

	return result
}

func handleRealResources(
	protocols map[core_meta.Protocol]struct{},
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

	if common_api.TargetRefKind(ref.Resource.ResourceType) == common_api.MeshService && ref.Resource.SectionName == "" {
		ref.Resource = kri.WithSectionName(ref.Resource, port.GetName())
	}

	service := destinationname.MustResolve(dest, port)

	clusterName := ref.Resource.String()

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
		WithExternalService(isExternalService)

	servicesAcc.AddBackendRef(resolve.NewResolvedBackendRef(ref), clusterBuilder.Build())

	return splitTo(clusterName)
}
