package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

// CreateDownstreamTlsContext creates DownstreamTlsContext for incoming connections
// It verifies that incoming connection has TLS certificate signed by Mesh CA with URI SAN of prefix spiffe://{mesh_name}/
// It secures inbound listener with certificate of "identity_cert" that will be received from the SDS (it contains URI SANs of all inbounds).
func CreateDownstreamTlsContext(downstreamMesh core_xds.CaRequest, mesh core_xds.IdentityCertRequest) (*envoy_tls.DownstreamTlsContext, error) {
	var validationSANMatchers []*envoy_tls.SubjectAltNameMatcher
	meshNames := downstreamMesh.MeshName()
	for _, meshName := range meshNames {
		validationSANMatchers = append(validationSANMatchers, MeshSpiffeIDPrefixMatcher(meshName))
	}

	commonTlsContext := createCommonTlsContext(mesh, downstreamMesh, validationSANMatchers)
	return &envoy_tls.DownstreamTlsContext{
		CommonTlsContext:         commonTlsContext,
		RequireClientCertificate: util_proto.Bool(true),
	}, nil
}

// CreateUpstreamTlsContext creates UpstreamTlsContext for outgoing connections
// It verifies that the upstream server has TLS certificate signed by Mesh CA with URI SAN of spiffe://{mesh_name}/{upstream_service}
// The downstream client exposes for the upstream server cert with multiple URI SANs, which means that if DP has inbound with services "web" and "web-api" and communicates with "backend"
// the upstream server ("backend") will see that DP with TLS certificate of URIs of "web" and "web-api".
// There is no way to correlate incoming request to "web" or "web-api" with outgoing request to "backend" to expose only one URI SAN.
//
// Pass "*" for upstreamService to validate that upstream service is a service that is part of the mesh (but not specific one)
func CreateUpstreamTlsContext(mesh core_xds.IdentityCertRequest, upstreamMesh core_xds.CaRequest, upstreamService string, sni string, verifyIdentities []string) (*envoy_tls.UpstreamTlsContext, error) {
	var validationSANMatchers []*envoy_tls.SubjectAltNameMatcher
	meshNames := upstreamMesh.MeshName()
	for _, meshName := range meshNames {
		if upstreamService == "*" {
			if len(verifyIdentities) == 0 {
				validationSANMatchers = append(validationSANMatchers, MeshSpiffeIDPrefixMatcher(meshName))
			}
			for _, identity := range verifyIdentities {
				stringMatcher := ServiceSpiffeIDMatcher(meshName, identity)
				matcher := &envoy_tls.SubjectAltNameMatcher{
					SanType: envoy_tls.SubjectAltNameMatcher_URI,
					Matcher: stringMatcher,
				}
				validationSANMatchers = append(validationSANMatchers, matcher)
			}
		} else {
			stringMatcher := ServiceSpiffeIDMatcher(meshName, upstreamService)
			matcher := &envoy_tls.SubjectAltNameMatcher{
				SanType: envoy_tls.SubjectAltNameMatcher_URI,
				Matcher: stringMatcher,
			}
			validationSANMatchers = append(validationSANMatchers, matcher)
		}
	}
	commonTlsContext := createCommonTlsContext(mesh, upstreamMesh, validationSANMatchers)
	commonTlsContext.AlpnProtocols = xds_tls.KumaALPNProtocols
	return &envoy_tls.UpstreamTlsContext{
		CommonTlsContext: commonTlsContext,
		Sni:              sni,
	}, nil
}

func createCommonTlsContext(ownMesh core_xds.IdentityCertRequest, targetMeshCa core_xds.CaRequest, matchers []*envoy_tls.SubjectAltNameMatcher) *envoy_tls.CommonTlsContext {
	meshCaSecret := NewSecretConfigSource(targetMeshCa.Name())
	identitySecret := NewSecretConfigSource(ownMesh.Name())

	return &envoy_tls.CommonTlsContext{
		ValidationContextType: &envoy_tls.CommonTlsContext_CombinedValidationContext{
			CombinedValidationContext: &envoy_tls.CommonTlsContext_CombinedCertificateValidationContext{
				DefaultValidationContext: &envoy_tls.CertificateValidationContext{
					MatchTypedSubjectAltNames: matchers,
				},
				ValidationContextSdsSecretConfig: meshCaSecret,
			},
		},
		TlsCertificateSdsSecretConfigs: []*envoy_tls.SdsSecretConfig{
			identitySecret,
		},
	}
}

func NewSecretConfigSource(secretName string) *envoy_tls.SdsSecretConfig {
	return &envoy_tls.SdsSecretConfig{
		Name: secretName,
		SdsConfig: &envoy_core.ConfigSource{
			ResourceApiVersion:    envoy_core.ApiVersion_V3,
			ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{},
		},
	}
}

func UpstreamTlsContextOutsideMesh(systemCaPath string, ca, cert, key []byte, allowRenegotiation, skipHostnameVerification, fallbackToSystemCa bool, hostname, sni string, sans []core_xds.SAN, minTlsVersion, maxTlsVersion *envoy_tls.TlsParameters_TlsProtocol) (*envoy_tls.UpstreamTlsContext, error) {
	tlsContext := &envoy_tls.UpstreamTlsContext{
		AllowRenegotiation: allowRenegotiation,
		Sni:                sni,
	}
	if cert != nil && key != nil {
		tlsContext.CommonTlsContext = &envoy_tls.CommonTlsContext{
			TlsCertificates: []*envoy_tls.TlsCertificate{
				{
					CertificateChain: dataSourceFromBytes(cert),
					PrivateKey:       dataSourceFromBytes(key),
				},
			},
		}
	}

	if ca != nil || (fallbackToSystemCa && systemCaPath != "") {
		if tlsContext.CommonTlsContext == nil {
			tlsContext.CommonTlsContext = &envoy_tls.CommonTlsContext{}
		}
		var matchNames []*envoy_tls.SubjectAltNameMatcher
		if !skipHostnameVerification {
			subjectAltNameMatch := hostname
			if len(sni) > 0 {
				subjectAltNameMatch = sni
			}
			for _, typ := range []envoy_tls.SubjectAltNameMatcher_SanType{
				envoy_tls.SubjectAltNameMatcher_DNS,
				envoy_tls.SubjectAltNameMatcher_IP_ADDRESS,
			} {
				matchNames = append(matchNames, &envoy_tls.SubjectAltNameMatcher{
					SanType: typ,
					Matcher: &envoy_type_matcher.StringMatcher{
						MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
							Exact: subjectAltNameMatch,
						},
					},
				})
			}
			for _, typ := range []envoy_tls.SubjectAltNameMatcher_SanType{
				envoy_tls.SubjectAltNameMatcher_DNS,
				envoy_tls.SubjectAltNameMatcher_IP_ADDRESS,
			} {
				for _, san := range sans {
					matcher := &envoy_tls.SubjectAltNameMatcher{
						SanType: typ,
					}
					switch san.MatchType {
					case core_xds.SANMatchExact:
						matcher.Matcher = &envoy_type_matcher.StringMatcher{
							MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
								Exact: san.Value,
							},
						}
					case core_xds.SANMatchPrefix:
						matcher.Matcher = &envoy_type_matcher.StringMatcher{
							MatchPattern: &envoy_type_matcher.StringMatcher_Prefix{
								Prefix: san.Value,
							},
						}
					}
					matchNames = append(matchNames, matcher)
				}
			}
		}

		var trustedCa *envoy_core.DataSource
		if ca == nil {
			trustedCa = &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_Filename{
					Filename: systemCaPath,
				},
			}
		} else {
			trustedCa = dataSourceFromBytes(ca)
		}

		tlsContext.CommonTlsContext.ValidationContextType = &envoy_tls.CommonTlsContext_ValidationContext{
			ValidationContext: &envoy_tls.CertificateValidationContext{
				TrustedCa:                 trustedCa,
				MatchTypedSubjectAltNames: matchNames,
			},
		}

		if minTlsVersion != nil {
			tlsContext.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{
				TlsMinimumProtocolVersion: *minTlsVersion,
			}
		}
		if maxTlsVersion != nil {
			if tlsContext.CommonTlsContext.TlsParams == nil {
				tlsContext.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{
					TlsMaximumProtocolVersion: *maxTlsVersion,
				}
			} else {
				tlsContext.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion = *maxTlsVersion
			}
		}
	}
	return tlsContext, nil
}

func dataSourceFromBytes(bytes []byte) *envoy_core.DataSource {
	return &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_InlineBytes{
			InlineBytes: bytes,
		},
	}
}

func MeshSpiffeIDPrefixMatcher(mesh string) *envoy_tls.SubjectAltNameMatcher {
	stringMatcher := &envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Prefix{
			Prefix: xds_tls.MeshSpiffeIDPrefix(mesh),
		},
	}

	return &envoy_tls.SubjectAltNameMatcher{
		SanType: envoy_tls.SubjectAltNameMatcher_URI,
		Matcher: stringMatcher,
	}
}

func ServiceSpiffeIDMatcher(mesh string, service string) *envoy_type_matcher.StringMatcher {
	return &envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
			Exact: xds_tls.ServiceSpiffeID(mesh, service),
		},
	}
}

func KumaIDMatcher(tagName, tagValue string) *envoy_type_matcher.StringMatcher {
	return &envoy_type_matcher.StringMatcher{
		MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
			Exact: xds_tls.KumaID(tagName, tagValue),
		},
	}
}

func StaticDownstreamTlsContextWithPath(certPath, keyPath string) *envoy_tls.DownstreamTlsContext {
	cert := &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_Filename{
			Filename: certPath,
		},
	}
	key := &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_Filename{
			Filename: keyPath,
		},
	}
	return staticDownstreamTlsContext(cert, key)
}

func StaticDownstreamTlsContextWithValue(keyPair *tls.KeyPair) *envoy_tls.DownstreamTlsContext {
	cert := &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_InlineBytes{
			InlineBytes: keyPair.CertPEM,
		},
	}
	key := &envoy_core.DataSource{
		Specifier: &envoy_core.DataSource_InlineBytes{
			InlineBytes: keyPair.KeyPEM,
		},
	}
	return staticDownstreamTlsContext(cert, key)
}

func staticDownstreamTlsContext(cert *envoy_core.DataSource, key *envoy_core.DataSource) *envoy_tls.DownstreamTlsContext {
	return &envoy_tls.DownstreamTlsContext{
		CommonTlsContext: &envoy_tls.CommonTlsContext{
			TlsCertificates: []*envoy_tls.TlsCertificate{
				{
					CertificateChain: cert,
					PrivateKey:       key,
				},
			},
		},
	}
}
