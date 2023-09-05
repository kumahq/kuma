package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideStaticTLSConfigurer struct {
	CertPath string
	KeyPath  string
}

var _ FilterChainConfigurer = &ServerSideStaticTLSConfigurer{}

func (c *ServerSideStaticTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext := envoy_tls.StaticDownstreamTlsContextWithPath(c.CertPath, c.KeyPath)

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
