package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/tls"
)

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
			if sni != "" {
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
