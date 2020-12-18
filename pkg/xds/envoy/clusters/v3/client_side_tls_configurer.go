package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type ClientSideTLSConfigurer struct {
	Endpoints []xds.Endpoint
}

var _ ClusterConfigurer = &ClientSideTLSConfigurer{}

func (c *ClientSideTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
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

			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &envoy_cluster.Cluster_TransportSocketMatch{
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
