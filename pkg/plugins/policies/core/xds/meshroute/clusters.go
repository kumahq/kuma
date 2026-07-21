package meshroute

import (
	"sort"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	unified_naming "github.com/kumahq/kuma/v3/pkg/core/naming/unified-naming"
	core_resources "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	meshmultizoneservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_sni "github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_matcher "github.com/kumahq/kuma/v3/pkg/envoy/builders/matcher"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	envoy_tags "github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/system_names"
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
				if !isMeshExternalService(meshCtx.EndpointMap[serviceName]) {
					continue
				}
				realResourceRef := service.BackendRef().RealResourceBackendRef()
				dest, port, ok := DestinationPortFromRef(meshCtx, realResourceRef)
				if !ok {
					continue
				}
				if proxy.WorkloadIdentity != nil {
					kriID := service.BackendRef().Resource()
					if errs := core_sni.ValidateKRI(kriID); len(errs) > 0 {
						continue
					}
					sni := core_sni.FromKRI(kriID)
					// we only want to route when are mesh-scoped zone egresses
					if len(meshCtx.ZoneEgresses) == 0 {
						continue
					}
					egressSANs := meshCtx.ZoneEgressSANs()
					if len(egressSANs) == 0 {
						continue
					}
					upstreamCtx, err := UpstreamTLSContext(proxy, sni, egressSANs)
					if err != nil {
						return nil, err
					}
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.UpstreamTLSContext(upstreamCtx))
				} else {
					sni := SniForBackendRef(realResourceRef, dest, port, systemNamespace)
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.ClientSideMTLSCustomSNI(
							proxy.SecretsTracker,
							unifiedNaming,
							meshCtx.Resource,
							mesh_proto.ZoneEgressServiceName,
							true,
							sni,
							false,
						))
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
						zone := realResourceRef.Resource.Zone
						isMZMS := common_api.TargetRefKind(realResourceRef.Resource.ResourceType) == common_api.MeshMultiZoneService
						// Local MeshService traffic stays sidecar-to-sidecar and never traverses a zone proxy,
						// so ZonesWithMeshScopedProxy (a remote-zone capability check) doesn't apply.
						// When the consuming proxy has WorkloadIdentity, always use the new KRI-based SNI for local MeshServices.
						// MeshMultiZoneService is always zone=="" (global resource) but aggregates MeshServices
						// across zones into a single cluster. Each remote zone is reachable either through a new
						// mesh-scoped zone proxy (matches the KRI SNI) or only through a legacy ZoneIngress (matches
						// the hash-based SNI), so a single cluster-wide SNI can't satisfy a mix. We keep the KRI SNI
						// as the default and add a per-zone transport socket match with the hash-based SNI for every
						// remote zone that only has a legacy ZoneIngress.
						var useKRISni bool
						var legacyZones []string
						if isMZMS {
							endpoints := meshCtx.EndpointMap[destinationname.ResolveLegacyFromDestination(dest, port)]
							var hasDefaultSNIEndpoint bool
							legacyZones, hasDefaultSNIEndpoint = classifyMZMSEndpointZones(endpoints, meshCtx.ZonesWithMeshScopedProxy)
							// Keep KRI SNI as the default unless every endpoint is reachable only
							// through a legacy ZoneIngress, in which case fall back to the hash-based SNI.
							useKRISni = len(legacyZones) == 0 || hasDefaultSNIEndpoint
						} else {
							useKRISni = zone == "" || isLocalMeshService || meshCtx.ZonesWithMeshScopedProxy[zone]
						}
						kriSNI := useKRISni && proxy.WorkloadIdentity != nil
						var sni string
						if kriSNI {
							if errs := core_sni.ValidateKRI(realResourceRef.Resource); len(errs) > 0 {
								continue
							}
							sni = core_sni.FromKRI(realResourceRef.Resource)
						} else {
							sni = SniForBackendRef(realResourceRef, dest, port, systemNamespace)
						}
						// ClientSideMultiIdentitiesMTLS validate MTLS enabled on the mesh
						if proxy.WorkloadIdentity != nil {
							sans := Identities(realResourceRef, meshCtx, true)
							upstreamCtx, err := UpstreamTLSContext(proxy, sni, sans)
							if err != nil {
								return nil, err
							}
							var zoneMatches map[string]*envoy_tls.UpstreamTlsContext
							if kriSNI && len(legacyZones) > 0 {
								legacyCtx, err := UpstreamTLSContext(proxy, SniForBackendRef(realResourceRef, dest, port, systemNamespace), sans)
								if err != nil {
									return nil, err
								}
								zoneMatches = make(map[string]*envoy_tls.UpstreamTlsContext, len(legacyZones))
								for _, lz := range legacyZones {
									zoneMatches[lz] = legacyCtx
								}
							}
							edsClusterBuilder.Configure(envoy_clusters.UpstreamTLSContextWithZoneMatches(upstreamCtx, zoneMatches))
						} else {
							edsClusterBuilder.Configure(envoy_clusters.ClientSideMultiIdentitiesMTLS(
								proxy.SecretsTracker,
								unifiedNaming,
								meshCtx.Resource,
								tlsReady,
								sni,
								Identities(realResourceRef, meshCtx, false),
								len(meshCtx.CAsByTrustDomain) > 0,
							))
						}
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
// new-style mesh-scoped zone proxy (MeshZoneAddress).
func classifyMZMSEndpointZones(endpoints []core_xds.Endpoint, zonesWithProxy map[string]bool) ([]string, bool) {
	seen := map[string]struct{}{}
	hasDefaultSNIEndpoint := false
	for _, ep := range endpoints {
		if ep.Locality == nil || ep.Locality.Zone == "" || zonesWithProxy[ep.Locality.Zone] {
			hasDefaultSNIEndpoint = true
			continue
		}
		seen[ep.Locality.Zone] = struct{}{}
	}
	return util_maps.SortedKeys(seen), hasDefaultSNIEndpoint
}
