package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/envoy"
	"github.com/Kong/kuma/pkg/xds/envoy/tls"
)

func ClientSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata, clientService string, tags []envoy.Tags) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&clientSideMTLSConfigurer{
			ctx:           ctx,
			metadata:      metadata,
			clientService: clientService,
			tags:          tags,
		})
	})
}

// UnknownDestinationClientSideMTLS configures cluster with mTLS for a mesh but without extensive destination verification (only Mesh is verified)
func UnknownDestinationClientSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&clientSideMTLSConfigurer{
			ctx:           ctx,
			metadata:      metadata,
			clientService: "*",
			tags:          nil,
		})
	})
}

type clientSideMTLSConfigurer struct {
	ctx           xds_context.Context
	metadata      *core_xds.DataplaneMetadata
	clientService string
	tags          []envoy.Tags
}

func (c *clientSideMTLSConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if !c.ctx.Mesh.Resource.MTLSEnabled() {
		return nil
	}
	// there might be a situation when there are multiple sam tags passed here for example two outbound listeners with the same tags, therefore we need to distinguish between them.
	distinctTags := envoy.DistinctTags(c.tags)
	switch {
	case len(distinctTags) == 0:
		transportSocket, err := c.createTransportSocket("")
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	case len(distinctTags) == 1:
		transportSocket, err := c.createTransportSocket(tls.SNIFromTags(c.tags[0]))
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	default:
		for _, tags := range distinctTags {
			sni := tls.SNIFromTags(tags)
			transportSocket, err := c.createTransportSocket(sni)
			if err != nil {
				return err
			}
			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_api.Cluster_TransportSocketMatch{
				Name: sni,
				Match: &pstruct.Struct{
					Fields: envoy.MetadataFields(tags),
				},
				TransportSocket: transportSocket,
			})
		}
	}
	return nil
}

func (c *clientSideMTLSConfigurer) createTransportSocket(sni string) (*envoy_core.TransportSocket, error) {
	tlsContext, err := envoy.CreateUpstreamTlsContext(c.ctx, c.metadata, c.clientService, sni)
	if err != nil {
		return nil, err
	}
	if tlsContext == nil {
		return nil, nil
	}
	pbst, err := proto.MarshalAnyDeterministic(tlsContext)
	if err != nil {
		return nil, err
	}
	transportSocket := &envoy_core.TransportSocket{
		Name: envoy_wellknown.TransportSocketTls,
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return transportSocket, nil
}
