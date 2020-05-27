package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func PassThroughCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&passThroughClusterConfigurer{
			name: name,
		})
		config.Add(&altStatNameConfigurer{})
		config.Add(&timeoutConfigurer{})
	})
}

type passThroughClusterConfigurer struct {
	name string
}

func (p *passThroughClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = p.name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_ORIGINAL_DST}
	c.LbPolicy = envoy_api.Cluster_CLUSTER_PROVIDED
	return nil
}
