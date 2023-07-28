package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
)

type PassThroughClusterConfigurer struct{}

var _ ClusterConfigurer = &PassThroughClusterConfigurer{}

func (p *PassThroughClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_ORIGINAL_DST}
	c.LbPolicy = envoy_cluster.Cluster_CLUSTER_PROVIDED
	return nil
}
