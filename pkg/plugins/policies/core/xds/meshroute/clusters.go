package meshroute

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
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
				if meshCtx.Resource.ZoneEgressEnabled() {
					endpoints := meshCtx.EndpointMap[serviceName]
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster())
					if isMeshExternalService(endpoints) {
						edsClusterBuilder.
							Configure(envoy_clusters.ClientSideMTLSCustomSNI(
								proxy.SecretsTracker,
								meshCtx.Resource,
								mesh_proto.ZoneEgressServiceName,
								true,
								SniForBackendRef(service.BackendRef(), meshCtx, systemNamespace),
							))
					} else {
						edsClusterBuilder.
							Configure(envoy_clusters.ClientSideMTLS(
								proxy.SecretsTracker,
								meshCtx.Resource,
								mesh_proto.ZoneEgressServiceName,
								tlsReady,
								clusterTags,
							))
					}
				} else {
					endpoints := meshCtx.ExternalServicesEndpointMap[serviceName]
					isIPv6 := proxy.Dataplane.IsIPv6()

					edsClusterBuilder.
						Configure(envoy_clusters.ProvidedCustomEndpointCluster(isIPv6, isMeshExternalService(endpoints), endpoints...))
					if isMeshExternalService(endpoints) {
						edsClusterBuilder.WithName(serviceName)
						edsClusterBuilder.Configure(
							envoy_clusters.MeshExternalServiceClientSideTLS(endpoints, proxy.Metadata.SystemCaPath, true),
						)
					} else {
						edsClusterBuilder.
							Configure(envoy_clusters.ClientSideTLS(endpoints))
					}
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
					if service.BackendRef().LegacyBackendRef.ReferencesRealObject() {
						if service.BackendRef().LegacyBackendRef.Kind == common_api.MeshService {
							if ms := meshCtx.MeshServiceByIdentifier[pointer.Deref(service.BackendRef().Resource).ResourceIdentifier]; ms != nil {
								tlsReady = ms.Status.TLS.Status == meshservice_api.TLSReady
							}
						}
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMultiIdentitiesMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource,
							tlsReady,
							SniForBackendRef(service.BackendRef(), meshCtx, systemNamespace),
							ServiceTagIdentities(service.BackendRef(), meshCtx),
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
				ResourceOrigin: createResourceOrigin(service.BackendRef(), meshCtx),
				Protocol:       protocol,
			})
		}
	}

	return resources, nil
}

func createResourceOrigin(
	ref core_model.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) *core_model.TypedResourceIdentifier {
	switch {
	case ref.LegacyBackendRef.Kind == common_api.MeshService && ref.LegacyBackendRef.ReferencesRealObject():
		ms := meshCtx.MeshServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		port, ok := ms.FindPort(pointer.Deref(ref.LegacyBackendRef.Port))
		if ok {
			return pointer.To(core_rules.UniqueKey(ms, port.Name))
		}
		return pointer.To(core_rules.UniqueKey(ms, ""))
	case ref.LegacyBackendRef.Kind == common_api.MeshExternalService:
		mes := meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		return pointer.To(core_rules.UniqueKey(mes, ""))
	case ref.LegacyBackendRef.Kind == common_api.MeshMultiZoneService:
		mzs := meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(ref.Resource).ResourceIdentifier]
		port, ok := mzs.FindPort(pointer.Deref(ref.LegacyBackendRef.Port))
		if ok {
			return pointer.To(core_rules.UniqueKey(mzs, port.Name))
		}
		return pointer.To(core_rules.UniqueKey(mzs, ""))
	}
	return nil
}

func SniForBackendRef(
	backendRef core_model.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
	systemNamespace string,
) string {
	var resource core_model.Resource
	var name string
	switch backendRef.LegacyBackendRef.Kind {
	case common_api.MeshService:
		ms := meshCtx.MeshServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		resource = ms
		name = ms.SNIName(systemNamespace)
	case common_api.MeshExternalService:
		mes := meshCtx.MeshExternalServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		resource = mes
		name = core_model.GetDisplayName(resource.GetMeta())
	case common_api.MeshMultiZoneService:
		resource = meshCtx.MeshMultiZoneServiceByIdentifier[pointer.Deref(backendRef.Resource).ResourceIdentifier]
		name = core_model.GetDisplayName(resource.GetMeta())
	}
	return tls.SNIForResource(
		name,
		resource.GetMeta().GetMesh(),
		resource.Descriptor().Name,
		pointer.Deref(backendRef.LegacyBackendRef.Port),
		nil,
	)
}

func ServiceTagIdentities(
	backendRef core_model.ResolvedBackendRef,
	meshCtx xds_context.MeshContext,
) []string {
	var result []string
	switch backendRef.LegacyBackendRef.Kind {
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
