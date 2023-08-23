package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_admin_tls "github.com/kumahq/kuma/pkg/envoy/admin/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideStaticMTLSConfigurer struct {
	MTLSCerts core_xds.ServerSideMTLSCerts
}

var _ FilterChainConfigurer = &ServerSideStaticMTLSConfigurer{}

func (c *ServerSideStaticMTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext := tls.StaticDownstreamTlsContextWithValue(&c.MTLSCerts.ServerPair)
	tlsContext.RequireClientCertificate = util_proto.Bool(true)

	tlsContext.CommonTlsContext.ValidationContextType = &envoy_tls.CommonTlsContext_ValidationContext{
		ValidationContext: &envoy_tls.CertificateValidationContext{
			TrustedCa: &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_InlineBytes{
					InlineBytes: c.MTLSCerts.CaPEM,
				},
			},
			MatchTypedSubjectAltNames: []*envoy_tls.SubjectAltNameMatcher{{
				SanType: envoy_tls.SubjectAltNameMatcher_DNS,
				Matcher: &envoy_type_matcher.StringMatcher{
					MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
						Exact: envoy_admin_tls.ClientCertSAN,
					},
					IgnoreCase: false,
				},
			}},
		},
	}

	pbst, err := util_proto.MarshalAnyDeterministic(tlsContext)
	if err != nil {
		return err
	}
	filterChain.TransportSocket = &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return nil
}
