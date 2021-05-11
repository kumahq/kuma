package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

type StaticClusterConfigurer struct {
	Name           string
	LoadAssignment *envoy_api.ClusterLoadAssignment
}

var _ ClusterConfigurer = &StaticClusterConfigurer{}

func (e *StaticClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_STATIC}
	c.LoadAssignment = e.LoadAssignment
	return nil
}
