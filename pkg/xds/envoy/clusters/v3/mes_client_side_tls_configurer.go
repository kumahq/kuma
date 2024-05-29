package clusters

import (
	"github.com/asaskevich/govalidator"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
)

type MesClientSideTLSConfigurer struct {
	Tls *v1alpha1.Tls
}

var _ ClusterConfigurer = &MesClientSideTLSConfigurer{}

func (c *MesClientSideTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if c.Tls.Enabled {
		tlsContext := &envoy_tls.UpstreamTlsContext{
			AllowRenegotiation: c.Tls.AllowRenegotiation,
		}

		var matchNames []*envoy_tls.SubjectAltNameMatcher
		if c.Tls.Verification.SubjectAltNames != nil {
			for _, san := range *c.Tls.Verification.SubjectAltNames {
				for _, typ := range []envoy_tls.SubjectAltNameMatcher_SanType{
					envoy_tls.SubjectAltNameMatcher_DNS,
					envoy_tls.SubjectAltNameMatcher_IP_ADDRESS,
					envoy_tls.SubjectAltNameMatcher_URI,
				} {
					if san.Type == v1alpha1.SANMatchExact {
						matchNames = append(matchNames, &envoy_tls.SubjectAltNameMatcher{
							SanType: typ,
							Matcher: &envoy_type_matcher.StringMatcher{
								MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
									Exact: san.Value,
								},
							},
						})
					} else {
						matchNames = append(matchNames, &envoy_tls.SubjectAltNameMatcher{
							SanType: typ,
							Matcher: &envoy_type_matcher.StringMatcher{
								MatchPattern: &envoy_type_matcher.StringMatcher_Prefix{
									Prefix: san.Value,
								},
							},
						})
					}
				}
			}

			tlsContext.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion

			tlsContext.CommonTlsContext.ValidationContextType = &envoy_tls.CommonTlsContext_ValidationContext{
				ValidationContext: &envoy_tls.CertificateValidationContext{
					TrustedCa:                 dataSourceFromBytes(ca),
					MatchTypedSubjectAltNames: matchNames,
				},
			}
		}


	}


	for _, ep := range c.Endpoints {
		if ep.ExternalService != nil && ep.ExternalService.TLSEnabled {
			sni := ep.ExternalService.ServerName
			if ep.ExternalService.ServerName == "" && govalidator.IsDNSName(ep.Target) {
				// SNI can only be a hostname, not IP
				sni = ep.Target
			}

			tlsContext, err := envoy_tls.UpstreamTlsContextOutsideMesh(
				ep.ExternalService.CaCert,
				ep.ExternalService.ClientCert,
				ep.ExternalService.ClientKey,
				ep.ExternalService.AllowRenegotiation,
				ep.ExternalService.SkipHostnameVerification,
				ep.Target,
				sni,
			)
			if err != nil {
				return err
			}

			pbst, err := proto.MarshalAnyDeterministic(tlsContext)
			if err != nil {
				return err
			}

			transportSocket := &envoy_core.TransportSocket{
				Name: "envoy.transport_sockets.tls",
				ConfigType: &envoy_core.TransportSocket_TypedConfig{
					TypedConfig: pbst,
				},
			}

			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_cluster.Cluster_TransportSocketMatch{
				Name: ep.Target,
				Match: &structpb.Struct{
					Fields: envoy_metadata.MetadataFields(tags.Tags(ep.Tags).WithoutTags(mesh_proto.ServiceTag)),
				},
				TransportSocket: transportSocket,
			})
		}
	}

	return nil
}
