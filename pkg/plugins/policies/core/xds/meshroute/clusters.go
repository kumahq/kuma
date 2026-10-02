package meshroute

import (
	"sort"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	unified_naming "github.com/kumahq/kuma/v2/pkg/core/naming/unified-naming"
	core_resources "github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core/destinationname"
	meshmultizoneservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_sni "github.com/kumahq/kuma/v2/pkg/core/resources/sni"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	bldrs_matcher "github.com/kumahq/kuma/v2/pkg/envoy/builders/matcher"
	bldrs_tls "github.com/kumahq/kuma/v2/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/resolve"
	util_maps "github.com/kumahq/kuma/v2/pkg/util/maps"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
	envoy_tags "github.com/kumahq/kuma/v2/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/system_names"
)

func GenerateClusters(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	services envoy_common.Services,
	systemNamespace string,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	unifiedNaming := unified_naming.Enabled(proxy.Metadata, meshCtx.Resource)

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
					realResourceRef := service.BackendRef().RealResourceBackendRef()
					dest, port, ok := DestinationPortFromRef(meshCtx, realResourceRef)
					if !ok {
						continue
					}
					edsClusterBuilder.Configure(envoy_clusters.EdsCluster())
					if egressSANs := meshCtx.ZoneEgressSANs(); len(egressSANs) > 0 && proxy.WorkloadIdentity != nil {
						// Zone proxies key the SNI by port name, a backendRef may use the number.
						sni := core_sni.FromKRI(kri.WithSectionName(realResourceRef.Resource, port.GetName()))
						upstreamCtx, err := UpstreamTLSContext(proxy, sni, egressSANs)
						if err != nil {
							return nil, err
						}
						edsClusterBuilder.Configure(envoy_clusters.UpstreamTLSContext(upstreamCtx))
					} else {
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMTLSCustomSNI(
							proxy.SecretsTracker,
							unifiedNaming,
							meshCtx.Resource,
							mesh_proto.ZoneEgressServiceName,
							true,
							SniForBackendRef(realResourceRef, dest, port, systemNamespace),
							false,
						))
					}
				case meshCtx.Resource.ZoneEgressEnabled():
					// path for old ExternalService
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.ClientSideMTLS(
							proxy.SecretsTracker,
							unifiedNaming,
							meshCtx.Resource,
							mesh_proto.ZoneEgressServiceName,
							tlsReady,
							clusterTags,
							false,
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
				case core_meta.ProtocolHTTP:
					edsClusterBuilder.Configure(envoy_clusters.Http())
				case core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
					edsClusterBuilder.Configure(envoy_clusters.Http2())
				default:
				}
			} else {
				edsClusterBuilder.
					Configure(envoy_clusters.EdsCluster()).
					Configure(envoy_clusters.Http2())

				if upstreamMeshName := cluster.Mesh(); upstreamMeshName != "" {
					for _, otherMesh := range meshCtx.Resources.Meshes().Items {
						if otherMesh.GetMeta().GetName() == upstreamMeshName {
							edsClusterBuilder.Configure(
								envoy_clusters.CrossMeshClientSideMTLS(
									proxy.SecretsTracker, unifiedNaming, meshCtx.Resource, otherMesh, serviceName, tlsReady, clusterTags,
								),
							)
							break
						}
					}
				} else {
					if realResourceRef := service.BackendRef().RealResourceBackendRef(); realResourceRef != nil {
						dest, port, ok := DestinationPortFromRef(meshCtx, realResourceRef)
						if !ok {
							continue
						}
						tlsReady = true // tls readiness is only relevant for MeshService
						isLocalMeshService := false
						if common_api.TargetRefKind(realResourceRef.Resource.ResourceType) == common_api.MeshService {
							ms := dest.(*meshservice_api.MeshServiceResource)
							// we only check TLS status for local service
							// services that are synced can be accessed only with TLS through ZoneIngress
							isLocalMeshService = ms.IsLocalMeshService()
							tlsReady = !isLocalMeshService || ms.Status.TLS.Status == meshservice_api.TLSReady
							protocol = port.GetProtocol()
						}
						mtls, err := ClientMTLS(proxy, meshCtx, realResourceRef, dest, port, systemNamespace, unifiedNaming, tlsReady)
						if err != nil {
							return nil, err
						}
						edsClusterBuilder.Configure(mtls)
					} else {
						edsClusterBuilder.Configure(envoy_clusters.ClientSideMTLS(proxy.SecretsTracker, unifiedNaming, meshCtx.Resource, serviceName, tlsReady, clusterTags, len(meshCtx.CAsByTrustDomain) > 0))
					}
				}
			}

			edsCluster, err := edsClusterBuilder.Build()
			if err != nil {
				return nil, errors.Wrapf(err, "build CDS for cluster %s failed", clusterName)
			}

			resources = resources.Add(&core_xds.Resource{
				Name:           clusterName,
				Origin:         metadata.OriginOutbound,
				Resource:       edsCluster,
				ResourceOrigin: service.BackendRef().Resource(),
				Protocol:       protocol,
			})
		}
	}

	return resources, nil
}

func UpstreamTLSContext(proxy *core_xds.Proxy, sni string, sans []string) (*envoy_tls.UpstreamTlsContext, error) {
	sanMatchers := make([]*bldrs_common.Builder[envoy_tls.SubjectAltNameMatcher], 0, len(sans))
	for _, san := range sans {
		conf := bldrs_tls.NewSubjectAltNameMatcher().Configure(bldrs_tls.URI(bldrs_matcher.NewStringMatcher().Configure(bldrs_matcher.ExactMatcher(san))))
		sanMatchers = append(sanMatchers, conf)
	}
	var validationSds bldrs_common.Configurer[envoy_tls.CommonTlsContext_CombinedCertificateValidationContext]
	if proxy.WorkloadIdentity.ExternalValidationSourceConfigurer != nil {
		validationSds = bldrs_tls.ValidationContextSdsSecretConfig(
			bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
				proxy.WorkloadIdentity.ExternalValidationSourceConfigurer(),
			),
		)
	} else {
		validationSds = bldrs_tls.ValidationContextSdsSecretConfig(
			bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
				bldrs_tls.SdsSecretConfigSource(
					system_names.SystemResourceNameCABundle,
					bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
				),
			),
		)
	}
	commonTlsContext := bldrs_tls.NewCommonTlsContext().
		Configure(bldrs_tls.CombinedCertificateValidationContext(
			bldrs_tls.NewCombinedCertificateValidationContext().
				Configure(validationSds).
				Configure(bldrs_tls.DefaultValidationContext(
					bldrs_tls.NewDefaultValidationContext().Configure(bldrs_tls.SANs(sanMatchers)),
				)),
		)).
		Configure(bldrs_tls.TlsCertificateSdsSecretConfigs([]*bldrs_common.Builder[envoy_tls.SdsSecretConfig]{
			bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
				proxy.WorkloadIdentity.IdentitySourceConfigurer(),
			),
		})).
		Configure(bldrs_tls.KumaAlpnProtocol())
	return bldrs_tls.NewUpstreamTLSContext().
		Configure(bldrs_tls.SNI(sni)).
		Configure(bldrs_tls.UpstreamCommonTlsContext(commonTlsContext)).
		Build()
}

// ClientMTLS configures the client side of mTLS to a MeshService or MeshMultiZoneService.
// The SNI follows the zone proxy in front of the destination: a mesh-scoped zone proxy matches
// only the KRI SNI, so a proxy on legacy mTLS needs it too whenever such a proxy serves the
// destination, while a legacy ZoneIngress matches only the hash-based SNI.
func ClientMTLS(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	backendRef *resolve.RealResourceBackendRef,
	dest core_resources.Destination,
	port core_resources.Port,
	systemNamespace string,
	unifiedNaming bool,
	tlsReady bool,
) (envoy_clusters.ClusterBuilderOpt, error) {
	hashSNI := SniForBackendRef(backendRef, dest, port, systemNamespace)
	sni := hashSNI
	var legacyZones []string
	if kriSNI, zones := useKRISNI(proxy, meshCtx, backendRef, dest, port); kriSNI {
		// Zone proxies key the SNI by port name, a backendRef may use the number.
		sni = core_sni.FromKRI(kri.WithSectionName(backendRef.Resource, port.GetName()))
		legacyZones = zones
	}
	// ClientSideMultiIdentitiesMTLS validate MTLS enabled on the mesh
	if proxy.WorkloadIdentity == nil {
		var zoneSNIs map[string]string
		if len(legacyZones) > 0 {
			zoneSNIs = make(map[string]string, len(legacyZones))
			for _, lz := range legacyZones {
				zoneSNIs[lz] = hashSNI
			}
		}
		return envoy_clusters.ClientSideMultiIdentitiesMTLS(
			proxy.SecretsTracker,
			unifiedNaming,
			meshCtx.Resource,
			tlsReady,
			sni,
			zoneSNIs,
			Identities(backendRef, meshCtx, false),
			len(meshCtx.CAsByTrustDomain) > 0,
		), nil
	}
	sans := Identities(backendRef, meshCtx, true)
	upstreamCtx, err := UpstreamTLSContext(proxy, sni, sans)
	if err != nil {
		return nil, err
	}
	var zoneMatches map[string]*envoy_tls.UpstreamTlsContext
	if len(legacyZones) > 0 {
		legacyCtx, err := UpstreamTLSContext(proxy, hashSNI, sans)
		if err != nil {
			return nil, err
		}
		zoneMatches = make(map[string]*envoy_tls.UpstreamTlsContext, len(legacyZones))
		for _, lz := range legacyZones {
			zoneMatches[lz] = legacyCtx
		}
	}
	return envoy_clusters.UpstreamTLSContextWithZoneMatches(upstreamCtx, zoneMatches), nil
}

// useKRISNI reports whether the client sends the KRI SNI and, if so, the remote zones that
// still need the hash-based SNI because only a legacy ZoneIngress serves them.
func useKRISNI(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	backendRef *resolve.RealResourceBackendRef,
	dest core_resources.Destination,
	port core_resources.Port,
) (bool, []string) {
	if common_api.TargetRefKind(backendRef.Resource.ResourceType) == common_api.MeshMultiZoneService {
		// MeshMultiZoneService has no zone but aggregates MeshServices across zones into a single
		// cluster, so a single cluster-wide SNI can't satisfy a mix of zone proxies. Keep the KRI SNI
		// as the default unless every endpoint is reachable only through a legacy ZoneIngress.
		endpoints := meshCtx.EndpointMap[destinationname.ResolveLegacyFromDestination(dest, port)]
		legacyZones, hasDefaultSNIEndpoint, viaMeshScopedProxy := classifyMZMSEndpointZones(endpoints, meshCtx.ZonesWithMeshScopedProxy, proxy.Zone)
		if len(legacyZones) > 0 && !hasDefaultSNIEndpoint {
			return false, nil
		}
		return proxy.WorkloadIdentity != nil || viaMeshScopedProxy, legacyZones
	}
	zone := backendRef.Resource.Zone
	// Local MeshService traffic stays sidecar-to-sidecar and never traverses a zone proxy.
	isLocalMeshService := false
	if ms, ok := dest.(*meshservice_api.MeshServiceResource); ok {
		isLocalMeshService = ms.IsLocalMeshService()
	}
	viaMeshScopedProxy := !isLocalMeshService && meshCtx.ZonesWithMeshScopedProxy[zone]
	if zone != "" && !isLocalMeshService && !viaMeshScopedProxy {
		return false, nil
	}
	return proxy.WorkloadIdentity != nil || viaMeshScopedProxy, nil
}

func SniForBackendRef(
	backendRef *resolve.RealResourceBackendRef,
	dest core_resources.Destination,
	port core_resources.Port,
	systemNamespace string,
) string {
	name := core_model.GetDisplayName(dest.GetMeta())
	if backendRef.Resource.ResourceType == meshservice_api.MeshServiceType {
		name = dest.(*meshservice_api.MeshServiceResource).SNIName(systemNamespace)
	}

	return tls.SNIForResource(name, dest.GetMeta().GetMesh(), dest.Descriptor().Name, port.GetValue(), nil)
}

func Identities(
	backendRef *resolve.RealResourceBackendRef,
	meshCtx xds_context.MeshContext,
	includeSpiffeID bool,
) []string {
	var result []string
	serviceTagTransformer := func(serviceTag string) string {
		return serviceTag
	}
	// we don't use function which transform service tag to the spiffe id on cluster configuration
	// instead we want to set it here. It's not required for SpiffeID type, only ServiceTag
	if includeSpiffeID {
		serviceTagTransformer = func(serviceTag string) string {
			return tls.ServiceSpiffeID(meshCtx.Resource.Meta.GetName(), serviceTag)
		}
	}
	switch common_api.TargetRefKind(backendRef.Resource.ResourceType) {
	case common_api.MeshService:
		ms := meshCtx.GetServiceByKRI(backendRef.Resource)
		if ms == nil {
			return result
		}
		for _, identity := range pointer.Deref(ms.(*meshservice_api.MeshServiceResource).Spec.Identities) {
			if identity.Type == meshservice_api.MeshServiceIdentityServiceTagType {
				result = append(result, serviceTagTransformer(identity.Value))
			}
			if identity.Type == meshservice_api.MeshServiceIdentitySpiffeIDType {
				result = append(result, identity.Value)
			}
		}
	case common_api.MeshMultiZoneService:
		svc := meshCtx.GetServiceByKRI(backendRef.Resource)
		if svc == nil {
			return result
		}
		identities := map[string]struct{}{}
		for _, matchedMs := range svc.(*meshmultizoneservice_api.MeshMultiZoneServiceResource).Status.MeshServices {
			ri := kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Name:         matchedMs.Name,
				Namespace:    matchedMs.Namespace,
				Zone:         matchedMs.Zone,
				Mesh:         matchedMs.Mesh,
			}
			ms := meshCtx.GetServiceByKRI(ri)
			if ms == nil {
				continue
			}
			for _, identity := range pointer.Deref(ms.(*meshservice_api.MeshServiceResource).Spec.Identities) {
				identities[identity.Value] = struct{}{}
			}
		}
		result = util_maps.SortedKeys(identities)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func isMeshExternalService(endpoints []core_xds.Endpoint) bool {
	if len(endpoints) > 0 {
		return endpoints[0].IsMeshExternalService()
	}
	return false
}

// classifyMZMSEndpointZones partitions a MeshMultiZoneService cluster's
// endpoints by the SNI format their zone expects. It returns the sorted,
// deduplicated set of remote zones reachable only through a legacy ZoneIngress
// (Locality.Zone set and absent from zonesWithProxy, matching the hash-based
// SNI), and whether any endpoint expects the default KRI-based SNI: endpoints
// without locality (local zone, sidecar-to-sidecar) or in a zone served by a
// new-style mesh-scoped zone proxy (MeshZoneAddress); and whether any remote
// endpoint is served by such a proxy.
func classifyMZMSEndpointZones(endpoints []core_xds.Endpoint, zonesWithProxy map[string]bool, localZone string) ([]string, bool, bool) {
	seen := map[string]struct{}{}
	hasDefaultSNIEndpoint := false
	viaMeshScopedProxy := false
	for _, ep := range endpoints {
		if ep.Locality == nil || ep.Locality.Zone == "" || zonesWithProxy[ep.Locality.Zone] {
			hasDefaultSNIEndpoint = true
			viaMeshScopedProxy = viaMeshScopedProxy || !ep.IsReachableFromZone(localZone)
			continue
		}
		seen[ep.Locality.Zone] = struct{}{}
	}
	return util_maps.SortedKeys(seen), hasDefaultSNIEndpoint, viaMeshScopedProxy
}
