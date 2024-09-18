package meshroute

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func GenerateClusters(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	services envoy_common.Services,
	systemNamespace string,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]
		protocol := meshCtx.GetServiceProtocol(serviceName)
		tlsReady := service.TLSReady()

		for _, cluster := range service.Clusters() {
			clusterName := cluster.Name()
			edsClusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName)

			clusterTags := []envoy_tags.Tags{cluster.Tags()}
			if meshCtx.IsExternalService(serviceName) {
				switch {
				case isMeshExternalService(meshCtx.EndpointMap[serviceName]):
					// MeshExternalService is only available through egress
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.ClientSideMTLSCustomSNI(
							proxy.SecretsTracker,
							meshCtx.Resource,
							mesh_proto.ZoneEgressServiceName,
							true,
							SniForBackendRef(service.BackendRef().RealResourceBackendRef(), meshCtx, systemNamespace),
						))
				case meshCtx.Resource.ZoneEgressEnabled():
					// path for old ExternalService
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.ClientSideMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource,
							mesh_proto.ZoneEgressServiceName,
							tlsReady,
							clusterTags,
						))
				default:
					// path for old ExternalService
					endpoints := meshCtx.ExternalServicesEndpointMap[serviceName]
					isIPv6 := proxy.Dataplane.IsIPv6()

					edsClusterBuilder.
						Configure(envoy_clusters.ProvidedCustomEndpointCluster(isIPv6, isMeshExternalService(endpoints), endpoints...)).
						Configure(envoy_clusters.ClientSideTLS(endpoints))
				}

				switch protocol {
				case core_mesh.ProtocolHTTP:
					edsClusterBuilder.Configure(envoy_clusters.Http())
				case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
					edsClusterBuilder.Configure(envoy_clusters.Http2())
				default:
				}
			} else {
				edsClusterBuilder.
					Configure(envoy_clusters.EdsCluster()).
					Configure(envoy_clusters.Http2())

				if upstreamMeshName := cluster.Mesh(); upstreamMeshName != "" {
					for _, otherMesh := range append(meshCtx.Resources.OtherMeshes().Items, meshCtx.Resource) {
						if otherMesh.GetMeta().GetName() == upstreamMeshName {
							edsClusterBuilder.Configure(
								envoy_clusters.CrossMeshClientSideMTLS(
									proxy.SecretsTracker, meshCtx.Resource, otherMesh, serviceName, tlsReady, clusterTags,
								),
							)
							break
						}
					}
				} else {
					if realResourceRef := service.BackendRef().RealResourceBackendRef(); realResourceRef != nil {
						if common_api.TargetRefKind(realResourceRef.Resource.ResourceType) == common_api.MeshService {
							if ms := meshCtx.MeshServiceByIdentifier[pointer.Deref(realResourceRef.Resource).ResourceIdentifier]; ms != nil {
								tlsReady = ms.Status.TLS.Status == meshservice_api.TLSReady
							}
						}
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMultiIdentitiesMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource,
							tlsReady,
							SniForBackendRef(realResourceRef, meshCtx, systemNamespace),
							ServiceTagIdentities(realResourceRef, meshCtx),
						))
					} else {
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource, serviceName, tlsReady, clusterTags))
					}
				}
			}

			edsCluster, err := edsClusterBuilder.Build()
			if err != nil {
				return nil, errors.Wrapf(err, "build CDS for cluster %s failed", clusterName)
			}

			resources = resources.Add(&core_xds.Resource{
				Name:           clusterName,
				Origin:         generator.OriginOutbound,
				Resource:       edsCluster,
				ResourceOrigin: service.BackendRef().ResourceOrNil(),
				Protocol:       protocol,
			})
		}
	}

	return resources, nil
}

func SniForBackendRef(
	backendRef *core_model.RealResourceBackendRef,
	meshCtx xds_context.MeshContext,
	systemNamespace string,
) string {
	var resource core_model.Resource
	var name string
	var port uint32
	switch common_api.TargetRefKind(backendRef.Resource.ResourceType) {
	case common_api.MeshService:
		ms := meshCtx.MeshServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		resource = ms
		name = ms.SNIName(systemNamespace)
		if p, ok := ms.FindPortByName(backendRef.Resource.SectionName); ok {
			port = p.Port
		}
	case common_api.MeshExternalService:
		mes := meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		resource = mes
		name = core_model.GetDisplayName(resource.GetMeta())
		port = uint32(mes.Spec.Match.Port)
	case common_api.MeshMultiZoneService:
		mzms := meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		resource = mzms
		name = core_model.GetDisplayName(resource.GetMeta())
		if p, ok := mzms.FindPortByName(backendRef.Resource.SectionName); ok {
			port = p.Port
		}
	}
	return tls.SNIForResource(name, resource.GetMeta().GetMesh(), resource.Descriptor().Name, port, nil)
}

func ServiceTagIdentities(
	backendRef *core_model.RealResourceBackendRef,
	meshCtx xds_context.MeshContext,
) []string {
	var result []string
	switch common_api.TargetRefKind(backendRef.Resource.ResourceType) {
	case common_api.MeshService:
		ms := meshCtx.MeshServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		for _, identity := range ms.Spec.Identities {
			if identity.Type == meshservice_api.MeshServiceIdentityServiceTagType {
				result = append(result, identity.Value)
			}
		}
	case common_api.MeshMultiZoneService:
		svc := meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		identities := map[string]struct{}{}
		for _, matchedMs := range svc.Status.MeshServices {
			ri := core_model.ResourceIdentifier{
				Name:      matchedMs.Name,
				Namespace: matchedMs.Namespace,
				Zone:      matchedMs.Zone,
				Mesh:      matchedMs.Mesh,
			}
			ms, ok := meshCtx.MeshServiceByIdentifier[ri]
			if !ok {
				continue
			}
			for _, identity := range ms.Spec.Identities {
				if identity.Type == meshservice_api.MeshServiceIdentityServiceTagType {
					identities[identity.Value] = struct{}{}
				}
			}
		}
		result = util_maps.SortedKeys(identities)
	}
	return result
}

func isMeshExternalService(endpoints []core_xds.Endpoint) bool {
	if len(endpoints) > 0 {
		return endpoints[0].IsMeshExternalService()
	}
	return false
}
