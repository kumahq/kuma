package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
)

type HttpDownstreamConfigurer struct{}

var _ ClusterConfigurer = &HttpDownstreamConfigurer{}

func (p *HttpDownstreamConfigurer) Configure(c *envoy_cluster.Cluster) error {
	return UpdateCommonHttpProtocolOptions(c, func(options *envoy_upstream_http.HttpProtocolOptions) {
		if options.UpstreamProtocolOptions == nil {
			options.UpstreamProtocolOptions = &envoy_upstream_http.HttpProtocolOptions_UseDownstreamProtocolConfig{
				UseDownstreamProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_UseDownstreamHttpConfig{},
			}
		}
	})
}
