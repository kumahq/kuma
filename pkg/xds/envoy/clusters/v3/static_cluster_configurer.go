package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
)

type StaticClusterConfigurer struct {
	Name    string
	Address string
	Port    uint32
}

var _ ClusterConfigurer = &StaticClusterConfigurer{}

func (e *StaticClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STATIC}
	c.LoadAssignment = envoy_endpoints.CreateStaticEndpoint(e.Name, e.Address, e.Port)
	return nil
}
