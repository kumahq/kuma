package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

func ClientSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata, clientService string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&clientSideMTLSConfigurer{
			ctx:           ctx,
			metadata:      metadata,
			clientService: clientService,
		})
	})
}

type clientSideMTLSConfigurer struct {
	ctx           xds_context.Context
	metadata      *core_xds.DataplaneMetadata
	clientService string
}

func (c *clientSideMTLSConfigurer) Configure(cluster *envoy_api.Cluster) error {
	tlsContext, err := envoy.CreateUpstreamTlsContext(c.ctx, c.metadata, c.clientService)
	if err != nil {
		return err
	}
	if tlsContext != nil {
		pbst, err := ptypes.MarshalAny(tlsContext)
		if err != nil {
			return err
		}
		cluster.TransportSocket = &envoy_core.TransportSocket{
			Name: envoy_wellknown.TransportSocketTls,
			ConfigType: &envoy_core.TransportSocket_TypedConfig{
				TypedConfig: pbst,
			},
		}
	}
	return nil
}
