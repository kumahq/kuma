package meshroute

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
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
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.ClientSideMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource,
							mesh_proto.ZoneEgressServiceName,
							tlsReady,
							clusterTags,
						))
				} else {
					endpoints := meshCtx.ExternalServicesEndpointMap[serviceName]
					isIPv6 := proxy.Dataplane.IsIPv6()

					edsClusterBuilder.
						Configure(envoy_clusters.ProvidedCustomEndpointCluster(isIPv6, isMeshExternalService(endpoints), endpoints...))
					if isMeshExternalService(endpoints) {
						edsClusterBuilder.WithName(envoy_names.GetMeshExternalServiceName(serviceName))
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
					if service.BackendRef().ReferencesRealObject() {
						if service.BackendRef().Kind == common_api.MeshService {
							if ms := meshCtx.MeshServiceByName[service.BackendRef().Name]; ms != nil {
								tlsReady = ms.Status.TLS.Status == v1alpha1.TLSReady
							}
						}
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMultiIdentitiesMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource,
							tlsReady,
							sniForBackendRef(service.BackendRef(), meshCtx, systemNamespace),
							serviceTagIdentities(service.BackendRef(), meshCtx),
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

func createResourceOrigin(ref common_api.BackendRef, meshCtx xds_context.MeshContext) *core_rules.UniqueResourceIdentifier {
	switch {
	case ref.Kind == common_api.MeshService && ref.ReferencesRealObject():
		ms := meshCtx.MeshServiceByName[ref.Name]
		port, ok := ms.FindPort(pointer.Deref(ref.Port))
		if ok {
			return pointer.To(core_rules.UniqueKey(ms, port.Name))
		}
		return pointer.To(core_rules.UniqueKey(ms, ""))
	case ref.Kind == common_api.MeshExternalService:
		mes := meshCtx.MeshExternalServiceByName[ref.Name]
		return pointer.To(core_rules.UniqueKey(mes, ""))
	case ref.Kind == common_api.MeshMultiZoneService:
		mzs := meshCtx.MeshMultiZoneServiceByName[ref.Name]
		port, ok := mzs.FindPort(pointer.Deref(ref.Port))
		if ok {
			return pointer.To(core_rules.UniqueKey(mzs, port.Name))
		}
		return pointer.To(core_rules.UniqueKey(mzs, ""))
	}
	return nil
}

func sniForBackendRef(
	backendRef common_api.BackendRef,
	meshCtx xds_context.MeshContext,
	systemNamespace string,
) string {
	var resource core_model.Resource
	var name string
	switch backendRef.Kind {
	case common_api.MeshService:
		ms := meshCtx.MeshServiceByName[backendRef.Name]
		resource = ms
		name = ms.SNIName(systemNamespace)
	case common_api.MeshExternalService:
		mes := meshCtx.MeshExternalServiceByName[backendRef.Name]
		resource = mes
		name = core_model.GetDisplayName(resource.GetMeta())
	case common_api.MeshMultiZoneService:
		resource = meshCtx.MeshMultiZoneServiceByName[backendRef.Name]
		name = core_model.GetDisplayName(resource.GetMeta())
	}
	return tls.SNIForResource(
		name,
		resource.GetMeta().GetMesh(),
		resource.Descriptor().Name,
		pointer.Deref(backendRef.Port),
		nil,
	)
}

func serviceTagIdentities(
	backendRef common_api.BackendRef,
	meshCtx xds_context.MeshContext,
) []string {
	var result []string
	switch backendRef.Kind {
	case common_api.MeshService:
		ms := meshCtx.MeshServiceByName[backendRef.Name]
		for _, identity := range ms.Spec.Identities {
			if identity.Type == v1alpha1.MeshServiceIdentityServiceTagType {
				result = append(result, identity.Value)
			}
		}
	case common_api.MeshMultiZoneService:
		svc := meshCtx.MeshMultiZoneServiceByName[backendRef.Name]
		identities := map[string]struct{}{}
		for _, matchedMs := range svc.Status.MeshServices {
			ms, ok := meshCtx.MeshServiceByName[matchedMs.Name]
			if !ok {
				continue
			}
			for _, identity := range ms.Spec.Identities {
				if identity.Type == v1alpha1.MeshServiceIdentityServiceTagType {
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
