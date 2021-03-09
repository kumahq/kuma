package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

type Http2Configurer struct {
}

var _ ClusterConfigurer = &Http2Configurer{}

func (p *Http2Configurer) Configure(c *envoy_api.Cluster) error {
	c.Http2ProtocolOptions = &envoy_core.Http2ProtocolOptions{}
	return nil
}
