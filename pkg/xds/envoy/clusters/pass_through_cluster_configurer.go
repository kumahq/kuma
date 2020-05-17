package clusters

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func PassThroughCluster(name string) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&PassThroughClusterConfigurer{
			Name: name,
		})
		config.Add(&AltStatNameConfigurer{})
		config.Add(&TimeoutConfigurer{})
	})
}

type PassThroughClusterConfigurer struct {
	Name string
}

func (p *PassThroughClusterConfigurer) Configure(c *v2.Cluster) error {
	c.Name = p.Name
	c.ClusterDiscoveryType = &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST}
	c.LbPolicy = v2.Cluster_CLUSTER_PROVIDED
	return nil
}
