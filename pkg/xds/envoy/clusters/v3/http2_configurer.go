package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type Http2Configurer struct {
}

var _ ClusterConfigurer = &Http2Configurer{}

func (p *Http2Configurer) Configure(c *envoy_cluster.Cluster) error {
	// nolint:staticcheck // keep deprecated options to be compatible with Envoy 1.16.x in Kuma 1.0.x
	c.Http2ProtocolOptions = &envoy_core.Http2ProtocolOptions{}

	// options := &envoy_upstream_http.HttpProtocolOptions{
	// 	UpstreamProtocolOptions: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
	// 		ExplicitHttpConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
	// 			ProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
	// 				Http2ProtocolOptions: &envoy_core.Http2ProtocolOptions{},
	// 			},
	// 		},
	// 	},
	// }
	//
	// pbst, err := proto.MarshalAnyDeterministic(options)
	// if err != nil {
	// 	return err
	// }
	// c.TypedExtensionProtocolOptions = map[string]*any.Any{
	// 	"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": pbst,
	// }
	return nil
}
