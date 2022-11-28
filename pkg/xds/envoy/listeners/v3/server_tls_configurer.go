package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	util_tls "github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerTLSConfigurer struct {
	CaPEM        []byte
	ServerPair   util_tls.KeyPair
	MinVersion   string
	MaxVersion   string
	CipherSuites []string
}

var _ FilterChainConfigurer = &ServerTLSConfigurer{}

func (c *ServerTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext := tls.StaticDownstreamTlsContext(&c.ServerPair)

	if len(c.MinVersion) > 0 {
		if minVersion, ok := envoy_tls.TlsParameters_TlsProtocol_value[c.MinVersion]; ok {
			tlsContext.CommonTlsContext.TlsParams.TlsMinimumProtocolVersion = envoy_tls.TlsParameters_TlsProtocol(minVersion)
		}
	}

	if len(c.MaxVersion) > 0 {
		if maxVersion, ok := envoy_tls.TlsParameters_TlsProtocol_value[c.MaxVersion]; ok {
			tlsContext.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion = envoy_tls.TlsParameters_TlsProtocol(maxVersion)
		}
	}

	if len(c.CipherSuites) > 0 {
		tlsContext.CommonTlsContext.TlsParams.CipherSuites = c.CipherSuites
	}

	if len(c.CaPEM) > 0 {
		tlsContext.RequireClientCertificate = util_proto.Bool(true)
		tlsContext.CommonTlsContext.ValidationContextType = &envoy_tls.CommonTlsContext_ValidationContext{
			ValidationContext: &envoy_tls.CertificateValidationContext{
				TrustedCa: &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineBytes{
						InlineBytes: c.CaPEM,
					},
				},
			},
		}
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
