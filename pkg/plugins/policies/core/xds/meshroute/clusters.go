package meshroute

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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
					if msName := service.MeshServiceName(); len(msName) > 0 {
						ms := meshCtx.MeshServiceByName[msName]
						name := ms.Meta.GetName()
						// we need to use original name and namespace for services that were synced from another cluster
						if dn := ms.GetMeta().GetLabels()[mesh_proto.DisplayName]; dn != "" {
							name = dn
							if ns := ms.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]; ns != "" {
								// when we sync resources from universal to kube, when we retrieve it has KubeNamespaceTag as label value
								if systemNamespace == "" || systemNamespace != ns {
									name += "." + ns
								}
							}
						}
						var serviceTagIdentities []string
						for _, identity := range ms.Spec.Identities {
							if identity.Type == v1alpha1.MeshServiceIdentityServiceTagType {
								serviceTagIdentities = append(serviceTagIdentities, identity.Value)
							}
						}
						sni := tls.SNIForResource(
							name,
							ms.Meta.GetMesh(),
							ms.Descriptor().Name,
							service.MeshServicePort(),
							nil, // todo(jakubdyszkiewicz) splits not yet supported
						)
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMultiIdentitiesMTLS(
							proxy.SecretsTracker,
							meshCtx.Resource, tlsReady, sni, serviceTagIdentities))
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
				Name:     clusterName,
				Origin:   generator.OriginOutbound,
				Resource: edsCluster,
			})
		}
	}

	return resources, nil
}

func isMeshExternalService(endpoints []core_xds.Endpoint) bool {
	if len(endpoints) > 0 {
		return endpoints[0].IsMeshExternalService()
	}
	return false
}
