package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v2"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v2"
)

type ClientSideTLSConfigurer struct {
	Endpoints []xds.Endpoint
}

var _ ClusterConfigurer = &ClientSideTLSConfigurer{}

func (c *ClientSideTLSConfigurer) Configure(cluster *envoy_api.Cluster) error {
	for _, ep := range c.Endpoints {
		if ep.ExternalService.TLSEnabled {
			tlsContext, err := envoy_tls.UpstreamTlsContextOutsideMesh(
				ep.ExternalService.CaCert,
				ep.ExternalService.ClientCert,
				ep.ExternalService.ClientKey,
				ep.Target)
			if err != nil {
				return err
			}

			pbst, err := proto.MarshalAnyDeterministic(tlsContext)
			if err != nil {
				return err
			}

			transportSocket := &envoy_core.TransportSocket{
				Name: "envoy.transport_sockets.tls",
				ConfigType: &envoy_core.TransportSocket_TypedConfig{
					TypedConfig: pbst,
				},
			}

			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_api.Cluster_TransportSocketMatch{
				Name: ep.Target,
				Match: &pstruct.Struct{
					Fields: envoy_metadata.MetadataFields(ep.Tags),
				},
				TransportSocket: transportSocket,
			})
		}
	}

	return nil
}
