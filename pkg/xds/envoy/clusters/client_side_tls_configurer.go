package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

func ClientSideTLS(sni string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&clientSideTLSConfigurer{
			sni: sni,
		})
	})
}

type clientSideTLSConfigurer struct {
	sni string
}

func (c *clientSideTLSConfigurer) Configure(cluster *envoy_api.Cluster) error {
	tlsContext, err := envoy.CreateUpstreamTlsContextNoMetadata(c.sni)
	if err != nil {
		return err
	}
	pbst, err := proto.MarshalAnyDeterministic(tlsContext)
	if err != nil {
		return err
	}
	transportSocket := &envoy_core.TransportSocket{
		Name: envoy_wellknown.TransportSocketTls,
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	if err != nil {
		return err
	}
	cluster.TransportSocket = transportSocket

	return nil
}
