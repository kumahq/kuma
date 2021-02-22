package clusters

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
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
		options := &envoy_upstream_http.HttpProtocolOptions{
			CommonHttpProtocolOptions: &envoy_core.HttpProtocolOptions{
				IdleTimeout: ptypes.DurationProto(t.Conf.GetHttp().GetIdleTimeout().AsDuration()),
			},
		}
		pbst, err := proto.MarshalAnyDeterministic(options)
		if err != nil {
			return err
		}
		cluster.TypedExtensionProtocolOptions = map[string]*any.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": pbst,
		}
	case mesh_core.ProtocolGRPC:
		if maxStreamDuration := t.Conf.GetGrpc().GetMaxStreamDuration().AsDuration(); maxStreamDuration != 0 {
			options := &envoy_upstream_http.HttpProtocolOptions{
				CommonHttpProtocolOptions: &envoy_core.HttpProtocolOptions{
					MaxStreamDuration: ptypes.DurationProto(maxStreamDuration),
				},
			}
			pbst, err := proto.MarshalAnyDeterministic(options)
			if err != nil {
				return err
			}
			cluster.TypedExtensionProtocolOptions = map[string]*any.Any{
				"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": pbst,
			}
		}
	}
	return nil
}
