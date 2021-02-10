package clusters

import (
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

const defaultConnectTimeout = 10 * time.Second

type TimeoutConfigurer struct {
	Protocol mesh_core.Protocol
	Conf     *mesh_proto.Timeout_Conf
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_api.Cluster) error {
	cluster.ConnectTimeout = ptypes.DurationProto(t.Conf.GetConnectTimeoutOrDefault(defaultConnectTimeout))
	switch t.Protocol {
	case mesh_core.ProtocolHTTP, mesh_core.ProtocolHTTP2:
		cluster.CommonHttpProtocolOptions = &envoy_api_v2_core.HttpProtocolOptions{
			IdleTimeout: ptypes.DurationProto(t.Conf.GetHTTPIdleTimeout()),
		}
	case mesh_core.ProtocolGRPC:
		if maxStreamDuration := t.Conf.GetGRPCMaxStreamDuration(); maxStreamDuration != nil {
			cluster.CommonHttpProtocolOptions = &envoy_api_v2_core.HttpProtocolOptions{
				MaxStreamDuration: ptypes.DurationProto(*maxStreamDuration),
			}
		}
	}
	return nil
}
