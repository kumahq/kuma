package clusters

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/types/known/durationpb"

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
	cluster.ConnectTimeout = durationpb.New(t.Conf.GetConnectTimeoutOrDefault(defaultConnectTimeout))
	switch t.Protocol {
	case mesh_core.ProtocolHTTP, mesh_core.ProtocolHTTP2:
		err := UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}
			options.CommonHttpProtocolOptions.IdleTimeout = durationpb.New(t.Conf.GetHttp().GetIdleTimeout().AsDuration())
		})
		if err != nil {
			return err
		}
	case mesh_core.ProtocolGRPC:
		if maxStreamDuration := t.Conf.GetGrpc().GetMaxStreamDuration().AsDuration(); maxStreamDuration != 0 {
			err := UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
				if options.CommonHttpProtocolOptions == nil {
					options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
				}
				options.CommonHttpProtocolOptions.MaxStreamDuration = durationpb.New(maxStreamDuration)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
