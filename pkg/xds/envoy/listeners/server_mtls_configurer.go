package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/envoy"
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
	tlsContext, err := envoy.CreateDownstreamTlsContext(c.ctx, c.metadata)
	if err != nil {
		return err
	}
	filterChain.TlsContext = tlsContext
	return nil
}
