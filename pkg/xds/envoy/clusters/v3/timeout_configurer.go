package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type TimeoutConfigurer struct {
	Protocol core_meta.Protocol
	Conf     *mesh_proto.Timeout_Conf
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = util_proto.Duration(t.Conf.GetConnectTimeoutOrDefault(policies_defaults.DefaultConnectTimeout))
	switch t.Protocol {
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
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
	if t.Protocol == core_meta.ProtocolGRPC {
		if msd := t.Conf.GetGrpc().GetMaxStreamDuration(); msd != nil && msd.AsDuration() != 0 {
			options.MaxStreamDuration = util_proto.Duration(msd.AsDuration())
		}
	}
}
