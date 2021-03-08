package clusters

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
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

func (t *TimeoutConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = ptypes.DurationProto(t.Conf.GetConnectTimeoutOrDefault(defaultConnectTimeout))
	switch t.Protocol {
	case mesh_core.ProtocolHTTP, mesh_core.ProtocolHTTP2:
		// nolint:staticcheck // keep deprecated options to be compatible with Envoy 1.16.x in Kuma 1.0.x
		cluster.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{
			IdleTimeout: ptypes.DurationProto(t.Conf.GetHttp().GetIdleTimeout().AsDuration()),
		}

		// options := &envoy_upstream_http.HttpProtocolOptions{
		// 	CommonHttpProtocolOptions: &envoy_core.HttpProtocolOptions{
		// 		IdleTimeout: ptypes.DurationProto(t.Conf.GetHttp().GetIdleTimeout().AsDuration()),
		// 	},
		// }
		// pbst, err := proto.MarshalAnyDeterministic(options)
		// if err != nil {
		// 	return err
		// }
		// cluster.TypedExtensionProtocolOptions = map[string]*any.Any{
		// 	"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": pbst,
		// }
	case mesh_core.ProtocolGRPC:
		if maxStreamDuration := t.Conf.GetGrpc().GetMaxStreamDuration().AsDuration(); maxStreamDuration != 0 {
			// nolint:staticcheck // keep deprecated options to be compatible with Envoy 1.16.x in Kuma 1.0.x
			cluster.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{
				MaxStreamDuration: ptypes.DurationProto(maxStreamDuration),
			}

			// options := &envoy_upstream_http.HttpProtocolOptions{
			// 	CommonHttpProtocolOptions: &envoy_core.HttpProtocolOptions{
			// 		MaxStreamDuration: ptypes.DurationProto(maxStreamDuration),
			// 	},
			// }
			// pbst, err := proto.MarshalAnyDeterministic(options)
			// if err != nil {
			// 	return err
			// }
			// cluster.TypedExtensionProtocolOptions = map[string]*any.Any{
			// 	"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": pbst,
			// }
		}
	}
	return nil
}
