package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
)

type LocalityWeightedLbConfigurer struct{}

func (c *LocalityWeightedLbConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
	}
	cluster.CommonLbConfig.LocalityConfigSpecifier = &envoy_cluster.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{}
	return nil
}
