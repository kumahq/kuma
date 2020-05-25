package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

func EdsCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&edsClusterConfigurer{
			name: name,
		})
		config.Add(&altStatNameConfigurer{})
		config.Add(&timeoutConfigurer{})
	})
}

type edsClusterConfigurer struct {
	name string
}

func (e *edsClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = e.name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_EDS}
	c.EdsClusterConfig = &envoy_api.Cluster_EdsClusterConfig{
		EdsConfig: &envoy_core.ConfigSource{
			ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
				Ads: &envoy_core.AggregatedConfigSource{},
			},
		},
	}
	return nil
}
