package clusters

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
)

func StaticCluster(name string, address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&staticClusterConfigurer{
			name:    name,
			address: address,
			port:    port,
		})
		config.Add(&altStatNameConfigurer{})
		config.Add(&timeoutConfigurer{})
	})
}

type staticClusterConfigurer struct {
	name    string
	address string
	port    uint32
}

func (e *staticClusterConfigurer) Configure(c *v2.Cluster) error {
	c.Name = e.name
	c.ClusterDiscoveryType = &v2.Cluster_Type{Type: v2.Cluster_STATIC}
	c.LoadAssignment = envoy_endpoints.CreateStaticEndpoint(e.name, e.address, e.port)
	return nil
}
