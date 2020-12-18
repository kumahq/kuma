package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v2"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v2"
)

type ClientSideMTLSConfigurer struct {
	Ctx           xds_context.Context
	Metadata      *core_xds.DataplaneMetadata
	ClientService string
	Tags          []envoy.Tags
}

var _ ClusterConfigurer = &ClientSideMTLSConfigurer{}

func (c *ClientSideMTLSConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if !c.Ctx.Mesh.Resource.MTLSEnabled() {
		return nil
	}
	mesh := c.Ctx.Mesh.Resource.GetMeta().GetName()
	// there might be a situation when there are multiple sam tags passed here for example two outbound listeners with the same tags, therefore we need to distinguish between them.
	distinctTags := envoy.DistinctTags(c.Tags)
	switch {
	case len(distinctTags) == 0:
		transportSocket, err := c.createTransportSocket("")
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	case len(distinctTags) == 1:
		transportSocket, err := c.createTransportSocket(tls.SNIFromTags(c.Tags[0].WithTags("mesh", mesh)))
		if err != nil {
			return err
		}
		cluster.TransportSocket = transportSocket
	default:
		for _, tags := range distinctTags {
			sni := tls.SNIFromTags(tags.WithTags("mesh", mesh))
			transportSocket, err := c.createTransportSocket(sni)
			if err != nil {
				return err
			}
			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_api.Cluster_TransportSocketMatch{
				Name: sni,
				Match: &pstruct.Struct{
					Fields: envoy_metadata.MetadataFields(tags),
				},
				TransportSocket: transportSocket,
			})
		}
	}
	return nil
}

func (c *ClientSideMTLSConfigurer) createTransportSocket(sni string) (*envoy_core.TransportSocket, error) {
	tlsContext, err := envoy_tls.CreateUpstreamTlsContext(c.Ctx, c.Metadata, c.ClientService, sni)
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
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	return transportSocket, nil
}
