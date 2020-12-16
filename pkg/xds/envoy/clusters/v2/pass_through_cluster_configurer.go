package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

type PassThroughClusterConfigurer struct {
	Name string
}

var _ ClusterConfigurer = &PassThroughClusterConfigurer{}

func (p *PassThroughClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = p.Name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_ORIGINAL_DST}
	c.LbPolicy = envoy_api.Cluster_CLUSTER_PROVIDED
	return nil
}
