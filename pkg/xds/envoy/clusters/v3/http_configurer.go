package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
)

type HttpConfigurer struct{}

var _ ClusterConfigurer = &HttpConfigurer{}

func (p *HttpConfigurer) Configure(c *envoy_cluster.Cluster) error {
	return UpdateCommonHttpProtocolOptions(c, func(options *envoy_upstream_http.HttpProtocolOptions) {
		if options.UpstreamProtocolOptions == nil {
			options.UpstreamProtocolOptions = &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
						HttpProtocolOptions: &envoy_config_core_v3.Http1ProtocolOptions{},
					},
				},
			}
		}
	})
}
