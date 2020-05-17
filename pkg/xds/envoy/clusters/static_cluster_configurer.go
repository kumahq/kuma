package clusters

import (
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func StaticCluster(name string, address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&StaticClusterConfigurer{
			Name:    name,
			address: address,
			port:    port,
		})
		config.Add(&AltStatNameConfigurer{})
		config.Add(&TimeoutConfigurer{})
	})
}

type StaticClusterConfigurer struct {
	Name string
	address string
	port uint32
}

func (e *StaticClusterConfigurer) Configure(c *v2.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &v2.Cluster_Type{Type: v2.Cluster_STATIC}
	c.LoadAssignment = envoy_endpoints.CreateStaticEndpoint(e.Name, e.address, e.port)
	return nil
}
