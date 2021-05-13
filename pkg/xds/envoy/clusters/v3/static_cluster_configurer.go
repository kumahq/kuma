package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

type StaticClusterConfigurer struct {
	Name           string
	LoadAssignment *envoy_endpoint.ClusterLoadAssignment
}

var _ ClusterConfigurer = &StaticClusterConfigurer{}

func (e *StaticClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STATIC}
	c.LoadAssignment = e.LoadAssignment
	return nil
}
