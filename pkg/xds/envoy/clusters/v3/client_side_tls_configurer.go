package clusters

import (
	"github.com/asaskevich/govalidator"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

type ClientSideTLSConfigurer struct {
	Endpoints []xds.Endpoint
}

var _ ClusterConfigurer = &ClientSideTLSConfigurer{}

func (c *ClientSideTLSConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	for _, ep := range c.Endpoints {
		if ep.ExternalService != nil && ep.ExternalService.TLSEnabled {
			sni := ep.ExternalService.ServerName
			if ep.ExternalService.ServerName == "" && govalidator.IsDNSName(ep.Target) {
				// SNI can only be a hostname, not IP
				sni = ep.Target
			}

			tlsContext, err := envoy_tls.UpstreamTlsContextOutsideMesh(
				ep.ExternalService.CaCert,
				ep.ExternalService.ClientCert,
				ep.ExternalService.ClientKey,
				ep.ExternalService.AllowRenegotiation,
				ep.Target,
				sni,
			)
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
				Match: &structpb.Struct{
					Fields: envoy_metadata.MetadataFields(ep.Tags),
				},
				TransportSocket: transportSocket,
			})
		}
	}

	return nil
}
