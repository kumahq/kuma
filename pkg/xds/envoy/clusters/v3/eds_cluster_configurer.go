package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type EdsClusterConfigurer struct{}

var _ ClusterConfigurer = &EdsClusterConfigurer{}

func (e *EdsClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_EDS}
	c.EdsClusterConfig = &envoy_cluster.Cluster_EdsClusterConfig{
		EdsConfig: &envoy_core.ConfigSource{
			ResourceApiVersion: envoy_core.ApiVersion_V3,
			ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
				Ads: &envoy_core.AggregatedConfigSource{},
			},
		},
	}
	return nil
}
