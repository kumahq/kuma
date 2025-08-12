package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type DownstreamTlsContextConfigurer struct {
	Config *envoy_tls.DownstreamTlsContext
}

var _ FilterChainConfigurer = &DownstreamTlsContextConfigurer{}

func (c *DownstreamTlsContextConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	pbst, err := util_proto.MarshalAnyDeterministic(c.Config)
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
