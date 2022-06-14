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
	Conf     *mesh_proto.Timeout_Conf
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = util_proto.Duration(t.Conf.GetConnectTimeoutOrDefault(defaultConnectTimeout))
	switch t.Protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		err := UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}

			t.setIdleTimeout(options.CommonHttpProtocolOptions)
			t.setMaxStreamDuration(options.CommonHttpProtocolOptions)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TimeoutConfigurer) setIdleTimeout(options *envoy_core.HttpProtocolOptions) {
	options.IdleTimeout = util_proto.Duration(t.Conf.GetHttp().GetIdleTimeout().AsDuration())
}

func (t *TimeoutConfigurer) setMaxStreamDuration(options *envoy_core.HttpProtocolOptions) {
	if msd := t.Conf.GetHttp().GetMaxStreamDuration(); msd != nil && msd.AsDuration() != 0 {
		options.MaxStreamDuration = msd
		return
	}

	// backwards compatibility
	if t.Protocol == core_mesh.ProtocolGRPC {
		if msd := t.Conf.GetGrpc().GetMaxStreamDuration(); msd != nil && msd.AsDuration() != 0 {
			options.MaxStreamDuration = util_proto.Duration(msd.AsDuration())
		}
	}
}
