package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type UpstreamTLSContextConfigure struct {
	Config *envoy_tls.UpstreamTlsContext
}

var _ ClusterConfigurer = &UpstreamTLSContextConfigure{}

func (c *UpstreamTLSContextConfigure) Configure(cluster *envoy_cluster.Cluster) error {
	pbst, err := proto.MarshalAnyDeterministic(c.Config)
	if err != nil {
		return err
	}
	transportSocket := &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}
	cluster.TransportSocket = transportSocket
	return nil
}
