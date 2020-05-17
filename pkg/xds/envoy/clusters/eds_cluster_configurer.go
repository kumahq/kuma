package clusters

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func EdsCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&EdsClusterConfigurer{
			Name: name,
		})
		config.Add(&AltStatNameConfigurer{})
		config.Add(&TimeoutConfigurer{})
	})
}

type EdsClusterConfigurer struct {
	Name string
}

func (e *EdsClusterConfigurer) Configure(c *v2.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &v2.Cluster_Type{Type: v2.Cluster_EDS}
	c.EdsClusterConfig = &v2.Cluster_EdsClusterConfig{
		EdsConfig: &envoy_core.ConfigSource{
			ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
				Ads: &envoy_core.AggregatedConfigSource{},
			},
		},
	}
	return nil
}
