package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/kumahq/kuma/v3/pkg/util/proto"
)

type UpstreamTLSContextConfigure struct {
	Config *envoy_tls.UpstreamTlsContext
}

var _ ClusterConfigurer = &UpstreamTLSContextConfigure{}

func (c *UpstreamTLSContextConfigure) Configure(cluster *envoy_cluster.Cluster) error {
	transportSocket, err := createTLSTransportSocket(c.Config)
	if err != nil {
		return err
	}
	cluster.TransportSocket = transportSocket
	return nil
}

func createTLSTransportSocket(config *envoy_tls.UpstreamTlsContext) (*envoy_core.TransportSocket, error) {
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return nil, err
	}
	return &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}, nil
}
