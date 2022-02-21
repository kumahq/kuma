package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideMTLSWithCPConfigurer struct {
	Ctx xds_context.Context
}

var _ FilterChainConfigurer = &ServerSideMTLSWithCPConfigurer{}

func (c *ServerSideMTLSWithCPConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext, err := tls.CreateDownstreamTlsContext(c.Ctx.Mesh.Resource)
	if err != nil {
		return err
	}
	if tlsContext == nil { // if mTLS is not enabled, fallback on self-signed certs
		tlsContext = tls.StaticDownstreamTlsContext(c.Ctx.ControlPlane.AdminProxyKeyPair)
	}

	// require certs that were presented when DP connects to the CP
	tlsContext.RequireClientCertificate = util_proto.Bool(true)
	tlsContext.CommonTlsContext.ValidationContextType = &envoy_extensions_transport_sockets_tls_v3.CommonTlsContext_ValidationContextSdsSecretConfig{
		ValidationContextSdsSecretConfig: &envoy_extensions_transport_sockets_tls_v3.SdsSecretConfig{
			Name: xds_tls.CpValidationCtx,
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
