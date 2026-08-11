package meshroute

import (
	"sort"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	meshmultizoneservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
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
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/system_names"
)

func GenerateClusters(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	services envoy_common.Services,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]
		protocol := meshCtx.GetServiceProtocol(serviceName)

		for _, cluster := range service.Clusters() {
			clusterName := cluster.Name()
			edsClusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName)
			if meshCtx.IsExternalService(serviceName) {
				realResourceRef := service.BackendRef().RealResourceBackendRef()
				_, port, ok := DestinationPortFromRef(meshCtx, realResourceRef)
				if !ok {
					continue
				}
				// A MeshExternalService is only reachable through a zone egress,
				// and the egress listener is generated only when the proxy has a
				// WorkloadIdentity (see ZoneProxyListenerGenerator). Without one
				// there is nothing to terminate the connection, so skip the cluster.
				if proxy.WorkloadIdentity == nil {
					continue
				}
				// The destination advertises its SNI from the resolved port
				// name, so normalize a numeric backend-ref section (named port
				// targeted by number) to the port name before deriving the KRI SNI.
				kriID := kri.WithSectionName(realResourceRef.Resource, port.GetName())
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

				if realResourceRef := service.BackendRef().RealResourceBackendRef(); realResourceRef != nil {
					dest, port, ok := DestinationPortFromRef(meshCtx, realResourceRef)
					if !ok {
						continue
					}
					tlsReady := true // tls readiness is only relevant for MeshService
					if common_api.TargetRefKind(realResourceRef.Resource.ResourceType) == common_api.MeshService {
						ms := dest.(*meshservice_api.MeshServiceResource)
						// we only check TLS status for local service
						// services that are synced can be accessed only with TLS through the zone proxy
						tlsReady = !ms.IsLocalMeshService() || ms.Status.TLS.Status == meshservice_api.TLSReady
						protocol = port.GetProtocol()
					}
					// Every zone is reachable through a mesh-scoped zone proxy, which
					// matches the KRI SNI, so a proxy with WorkloadIdentity always uses it.
					if proxy.WorkloadIdentity != nil {
						// The destination advertises its SNI from the resolved
						// port name, so normalize a numeric backend-ref section
						// (named port targeted by number) to the port name
						// before deriving the KRI SNI.
						kriID := kri.WithSectionName(realResourceRef.Resource, port.GetName())
						if errs := core_sni.ValidateKRI(kriID); len(errs) > 0 {
							continue
						}
						// A proxy receives its own identity independently of
						// the destination's, so requiring mTLS before the
						// destination reports it can serve TLS drops every
						// request sent in between. Ready is sticky for as long
						// as the mesh keeps any MeshIdentity, and only resets
						// once they are gone - which is exactly when the
						// destination stops terminating TLS and plaintext
						// becomes correct again.
						if tlsReady {
							sans := Identities(realResourceRef, meshCtx)
							upstreamCtx, err := UpstreamTLSContext(proxy, core_sni.FromKRI(kriID), sans)
							if err != nil {
								return nil, err
							}
							edsClusterBuilder.Configure(envoy_clusters.UpstreamTLSContext(upstreamCtx))
						}
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

func Identities(
	backendRef *resolve.RealResourceBackendRef,
	meshCtx xds_context.MeshContext,
) []string {
	var result []string
	switch common_api.TargetRefKind(backendRef.Resource.ResourceType) {
	case common_api.MeshService:
		ms := meshCtx.GetServiceByKRI(backendRef.Resource)
		if ms == nil {
			return result
		}
		for _, identity := range pointer.Deref(ms.(*meshservice_api.MeshServiceResource).Spec.Identities) {
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
				if identity.Type != meshservice_api.MeshServiceIdentitySpiffeIDType {
					continue
				}
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
