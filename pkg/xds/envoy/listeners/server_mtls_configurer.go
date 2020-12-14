package listeners

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

	tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v2"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func ServerSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&ServerSideMTLSConfigurer{
			ctx:      ctx,
			metadata: metadata,
		})
	})
}

type ServerSideMTLSConfigurer struct {
	ctx      xds_context.Context
	metadata *core_xds.DataplaneMetadata
}

func (c *ServerSideMTLSConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	tlsContext, err := tls.CreateDownstreamTlsContext(c.ctx, c.metadata)
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
