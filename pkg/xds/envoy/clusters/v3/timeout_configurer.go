package clusters

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

const defaultConnectTimeout = 10 * time.Second

type TimeoutConfigurer struct {
	Protocol core_mesh.Protocol
	Timeout  *core_mesh.TimeoutResource
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	var conf *mesh_proto.Timeout_Conf

	// Protobuf nil-checking semantics returns defaults for a nil config.
	if t.Timeout != nil {
		conf = t.Timeout.Spec.GetConf()
	}

	cluster.ConnectTimeout = util_proto.Duration(conf.GetConnectTimeoutOrDefault(defaultConnectTimeout))
	switch t.Protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		err := UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}
			options.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(conf.GetHttp().GetIdleTimeout().AsDuration())
		})
		if err != nil {
			return err
		}
	case core_mesh.ProtocolGRPC:
		if maxStreamDuration := conf.GetGrpc().GetMaxStreamDuration().AsDuration(); maxStreamDuration != 0 {
			err := UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
				if options.CommonHttpProtocolOptions == nil {
					options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
				}
				options.CommonHttpProtocolOptions.MaxStreamDuration = util_proto.Duration(maxStreamDuration)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
