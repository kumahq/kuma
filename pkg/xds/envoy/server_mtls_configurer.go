package envoy

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
)

func ServerSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
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

func (c *ServerSideMTLSConfigurer) Configure(l *v2.Listener) error {
	for i := range l.FilterChains {
		l.FilterChains[i].TlsContext = CreateDownstreamTlsContext(c.ctx, c.metadata)
	}

	return nil
}
