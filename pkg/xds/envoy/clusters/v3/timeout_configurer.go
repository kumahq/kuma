package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	policies_defaults "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/defaults"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

type TimeoutConfigurer struct {
	Protocol core_meta.Protocol
	Conf     *envoy_common.Timeouts
}

var _ ClusterConfigurer = &TimeoutConfigurer{}

func (t *TimeoutConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = util_proto.Duration(t.Conf.ConnectOrDefault(policies_defaults.DefaultConnectTimeout))
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
	options.IdleTimeout = util_proto.Duration(t.Conf.GetHttpIdle())
}

func (t *TimeoutConfigurer) setMaxStreamDuration(options *envoy_core.HttpProtocolOptions) {
	if msd := t.Conf.GetHttpMaxStreamDuration(); msd != 0 {
		options.MaxStreamDuration = util_proto.Duration(msd)
	}
}
