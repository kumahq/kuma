package clusters

import (
	"context"

	"github.com/asaskevich/govalidator"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type MesClientSideTLSConfigurer struct {
	Tls          *v1alpha1.Tls
	Loader       datasource.Loader
	Mesh         string
	SystemCaPath string
}

var _ ClusterConfigurer = &MesClientSideTLSConfigurer{}

func (c *MesClientSideTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if c.Tls != nil && c.Tls.Enabled {
		tlsContext := &envoy_tls.UpstreamTlsContext{
			AllowRenegotiation: c.Tls.AllowRenegotiation,
		}

		var cert, key *envoy_core.DataSource
		if shouldVerifyCa(c.Tls.Verification) {
			var ca *envoy_core.DataSource
			if c.Tls.Verification.CaCert != nil {
				var err error
				ca, err = c.toEnvoyDataSource(c.Tls.Verification.CaCert)
				if err != nil {
					return err
				}
			} else {
				ca = &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_Filename{
						Filename: c.SystemCaPath,
					},
				}
			}
			tlsContext.CommonTlsContext = &envoy_tls.CommonTlsContext{
				ValidationContextType: &envoy_tls.CommonTlsContext_ValidationContext{
					ValidationContext: &envoy_tls.CertificateValidationContext{
						TrustedCa: ca,
					},
				},
			}
		}

		if c.Tls.Verification.ClientCert != nil && c.Tls.Verification.ClientKey != nil {
			var err error
			cert, err = c.toEnvoyDataSource(c.Tls.Verification.ClientCert)
			if err != nil {
				return err
			}
			key, err = c.toEnvoyDataSource(c.Tls.Verification.ClientKey)
			if err != nil {
				return err
			}
		}

		var matchNames []*envoy_tls.SubjectAltNameMatcher
		if c.Tls.Verification.SubjectAltNames != nil {
			for _, san := range *c.Tls.Verification.SubjectAltNames {
				var typ envoy_tls.SubjectAltNameMatcher_SanType
				switch {
				case govalidator.IsIP(san.Value):
					typ = envoy_tls.SubjectAltNameMatcher_IP_ADDRESS
				case govalidator.IsDNSName(san.Value):
					typ = envoy_tls.SubjectAltNameMatcher_DNS
				default:
					typ = envoy_tls.SubjectAltNameMatcher_URI
				}
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

		tlsContext.CommonTlsContext = &envoy_tls.CommonTlsContext{
			TlsCertificates: []*envoy_tls.TlsCertificate{
				{
					CertificateChain: cert,
					PrivateKey:       key,
				},
			},
			ValidationContextType: &envoy_tls.CommonTlsContext_ValidationContext{
				ValidationContext: &envoy_tls.CertificateValidationContext{
					MatchTypedSubjectAltNames: matchNames,
				},
			},
			TlsParams: &envoy_tls.TlsParameters{},
		}

		if c.Tls.Version.Min != nil {
			tlsContext.CommonTlsContext.TlsParams.TlsMinimumProtocolVersion = mapTlsToEnvoyVersion(*c.Tls.Version.Min)
		}
		if c.Tls.Version.Max != nil {
			tlsContext.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion = mapTlsToEnvoyVersion(*c.Tls.Version.Max)
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
			Name:            cluster.Name,
			TransportSocket: transportSocket,
		})
	}

	return nil
}

func shouldVerifyCa(verification *v1alpha1.Verification) bool {
	if verification == nil {
		return true
	}

	if verification.Mode != nil && (*verification.Mode == v1alpha1.TLSVerificationSkipAll || *verification.Mode == v1alpha1.TLSVerificationSkipCA) {
		return false
	}

	return true
}

func (c *MesClientSideTLSConfigurer) toEnvoyDataSource(ds *common_api.DataSource) (*envoy_core.DataSource, error) {
	caCert, err := c.Loader.Load(context.Background(), c.Mesh, ds.ConvertToProto())
	if err != nil {
		return nil, err
	}
	return &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_InlineBytes{
			InlineBytes: caCert,
		},
	}, nil
}

func mapTlsToEnvoyVersion(version v1alpha1.TlsVersion) envoy_tls.TlsParameters_TlsProtocol {
	switch version {
	case v1alpha1.TLSVersion13:
		return envoy_tls.TlsParameters_TLSv1_3
	case v1alpha1.TLSVersion12:
		return envoy_tls.TlsParameters_TLSv1_2
	case v1alpha1.TLSVersion11:
		return envoy_tls.TlsParameters_TLSv1_1
	case v1alpha1.TLSVersion10:
		return envoy_tls.TlsParameters_TLSv1_0
	case v1alpha1.TLSVersionAuto:
		fallthrough
	default:
		return envoy_tls.TlsParameters_TLS_AUTO
	}
}
