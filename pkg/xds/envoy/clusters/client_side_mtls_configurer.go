package clusters

import (
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/envoy"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes"
)

func ClientSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&ClientSideMTLSConfigurer{
			ctx:            ctx,
			metadata:       metadata,
		})
	})
}

type ClientSideMTLSConfigurer struct {
	ctx xds_context.Context
	metadata *core_xds.DataplaneMetadata
}

func (c *ClientSideMTLSConfigurer) Configure(cluster *v2.Cluster) error {
	tlsContext, err := envoy.CreateUpstreamTlsContext(c.ctx, c.metadata)
	if err != nil {
		return err
	}
	if tlsContext != nil {
		pbst, err := ptypes.MarshalAny(tlsContext)
		if err != nil {
			return err
		}
		cluster.TransportSocket = &envoy_core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &envoy_core.TransportSocket_TypedConfig{
				TypedConfig: pbst,
			},
		}
	}
	return nil
}
