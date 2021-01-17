package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type Http2Configurer struct {
}

var _ ClusterConfigurer = &Http2Configurer{}

func (p *Http2Configurer) Configure(c *envoy_cluster.Cluster) error {
	// nolint:staticcheck ignore this is deprecated in the newest Envoy proto in master below code does not work with previous envoy
	c.Http2ProtocolOptions = &envoy_core.Http2ProtocolOptions{}

	// todo(jakubdyszkiewicz) switch to above code with Envoy 1.17+
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
