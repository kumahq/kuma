package meshroute

import (
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_sni "github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_matcher "github.com/kumahq/kuma/v3/pkg/envoy/builders/matcher"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/system_names"
)

// destination is everything a cluster needs to know about what it talks to: the
// protocol advertised by the destination port, the identities this proxy has to
// accept when it originates mTLS, and the upstream HTTP options to apply.
type destination struct {
	protocol core_meta.Protocol
	sans     []string
	// upstreamHTTP is nil when the connection carries opaque bytes.
	upstreamHTTP envoy_clusters.ClusterBuilderOpt
	// plaintext keeps the cluster without a transport socket, for a destination
	// that does not terminate TLS yet.
	plaintext bool
}

// upstreamHTTPOptions returns the upstream options matching the protocol the
// destination speaks, or nil when the connection carries opaque bytes.
func upstreamHTTPOptions(protocol core_meta.Protocol) envoy_clusters.ClusterBuilderOpt {
	switch protocol {
	case core_meta.ProtocolHTTP:
		return envoy_clusters.Http()
	case core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
		return envoy_clusters.Http2()
	default:
		return nil
	}
}

// GenerateClusters builds one EDS cluster per outbound destination. Every
// destination is reachable through a mesh-scoped zone proxy, so a cluster
// originates mTLS towards the KRI SNI of the destination port once this proxy
// has an identity to present and the destination can terminate TLS.
func GenerateClusters(
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
	services envoy_common.Services,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	hasIdentity := proxy.WorkloadIdentity != nil

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]

		backendRef := service.BackendRef().RealResourceBackendRef()
		if backendRef == nil {
			continue
		}
		destResource, port, ok := DestinationPortFromRef(meshCtx, backendRef)
		if !ok {
			continue
		}

		var dest destination
		switch backendRef.Resource.ResourceType {
		case meshservice_api.MeshServiceType:
			ms := destResource.(*meshservice_api.MeshServiceResource)
			dest = destination{
				protocol: port.GetProtocol(),
				sans:     meshServiceIdentities(meshCtx, backendRef.Resource),
				// The hop between the two proxies is HTTP/2 no matter which
				// protocol the application speaks.
				upstreamHTTP: envoy_clusters.Http2(),
				plaintext:    !ms.TerminatesTLS(),
			}
		case meshmultizoneservice_api.MeshMultiZoneServiceType:
			dest = destination{
				protocol:     port.GetProtocol(),
				sans:         meshMultiZoneServiceIdentities(meshCtx, backendRef.Resource),
				upstreamHTTP: envoy_clusters.Http2(),
			}
		case meshexternalservice_api.MeshExternalServiceType:
			// An external service is only reachable through a zone egress: the
			// egress terminates this connection and matches the KRI SNI, so the
			// egress identity is what this proxy has to trust. The egress
			// forwards the original protocol, so the upstream keeps speaking it.
			egressSANs := meshCtx.ZoneEgressSANs()
			if len(egressSANs) == 0 {
				continue
			}
			dest = destination{
				protocol:     port.GetProtocol(),
				sans:         egressSANs,
				upstreamHTTP: upstreamHTTPOptions(port.GetProtocol()),
			}
		default:
			continue
		}

		kriID := kri.WithSectionName(backendRef.Resource, port.GetName())
		if errs := core_sni.ValidateKRI(kriID); len(errs) > 0 {
			continue
		}

		var transportSocket envoy_clusters.ClusterBuilderOpt
		if hasIdentity && !dest.plaintext {
			upstreamCtx, err := UpstreamTLSContext(proxy, core_sni.FromKRI(kriID), dest.sans)
			if err != nil {
				return nil, err
			}
			transportSocket = envoy_clusters.UpstreamTLSContext(upstreamCtx)
		}

		for _, cluster := range service.Clusters() {
			clusterName := cluster.Name()
			edsCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName).
				Configure(envoy_clusters.EdsCluster()).
				ConfigureIf(dest.upstreamHTTP != nil, dest.upstreamHTTP).
				ConfigureIf(transportSocket != nil, transportSocket).
				Build()
			if err != nil {
				return nil, errors.Wrapf(err, "build CDS for cluster %s failed", clusterName)
			}

			resources = resources.Add(&core_xds.Resource{
				Name:           clusterName,
				Origin:         metadata.OriginOutbound,
				Resource:       edsCluster,
				ResourceOrigin: service.BackendRef().Resource(),
				Protocol:       dest.protocol,
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

// meshServiceIdentities returns the SPIFFE IDs advertised by a MeshService.
func meshServiceIdentities(meshCtx xds_context.MeshContext, id kri.Identifier) []string {
	ms, ok := meshCtx.GetServiceByKRI(id).(*meshservice_api.MeshServiceResource)
	if !ok {
		return nil
	}
	var identities []string
	for _, identity := range pointer.Deref(ms.Spec.Identities) {
		if identity.Type == meshservice_api.MeshServiceIdentitySpiffeIDType {
			identities = append(identities, identity.Value)
		}
	}
	return identities
}

// meshMultiZoneServiceIdentities returns the SPIFFE IDs advertised by every
// MeshService the MeshMultiZoneService matched, across all zones.
func meshMultiZoneServiceIdentities(meshCtx xds_context.MeshContext, id kri.Identifier) []string {
	mzms, ok := meshCtx.GetServiceByKRI(id).(*meshmultizoneservice_api.MeshMultiZoneServiceResource)
	if !ok {
		return nil
	}
	identities := map[string]struct{}{}
	for _, matched := range mzms.Status.MeshServices {
		msID := kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Name:         matched.Name,
			Namespace:    matched.Namespace,
			Zone:         matched.Zone,
			Mesh:         matched.Mesh,
		}
		for _, identity := range meshServiceIdentities(meshCtx, msID) {
			identities[identity] = struct{}{}
		}
	}
	return util_maps.SortedKeys(identities)
}
