package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ServerSideMTLSConfigurer struct {
	Ctx xds_context.Context
}

var _ FilterChainConfigurer = &ServerSideMTLSConfigurer{}

func (c *ServerSideMTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext, err := tls.CreateDownstreamTlsContext(c.Ctx)
	if err != nil {
		return err
	}
	if tlsContext != nil {
		pbst, err := proto.MarshalAnyDeterministic(tlsContext)
		if err != nil {
			return err
		}
		filterChain.TransportSocket = &envoy_core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &envoy_core.TransportSocket_TypedConfig{
				TypedConfig: pbst,
			},
		}
	}
	return nil
}
