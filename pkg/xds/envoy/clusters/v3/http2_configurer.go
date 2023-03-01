package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// Window size defaults.
const (
	DefaultInitialStreamWindowSize     = 64 * 1024
	DefaultInitialConnectionWindowSize = 1024 * 1024
)

type Http2Configurer struct {
	EdgeProxyWindowSizes bool
}

var _ ClusterConfigurer = &Http2Configurer{}

func (p *Http2Configurer) Configure(c *envoy_cluster.Cluster) error {
	return UpdateCommonHttpProtocolOptions(c, func(options *envoy_upstream_http.HttpProtocolOptions) {
		opts := &envoy_core.Http2ProtocolOptions{}

		// These are from Envoy's best practices for edge proxy configuration:
		// https://www.envoyproxy.io/docs/envoy/latest/configuration/best_practices/edge
		if p.EdgeProxyWindowSizes {
			opts.InitialStreamWindowSize = util_proto.UInt32(DefaultInitialStreamWindowSize)
			opts.InitialConnectionWindowSize = util_proto.UInt32(DefaultInitialConnectionWindowSize)
		}

		if options.UpstreamProtocolOptions == nil {
			options.UpstreamProtocolOptions = &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
						Http2ProtocolOptions: opts,
					},
				},
			}
		}
	})
}
