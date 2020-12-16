package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v2"
)

type StaticClusterConfigurer struct {
	Name    string
	Address string
	Port    uint32
}

var _ ClusterConfigurer = &StaticClusterConfigurer{}

func (e *StaticClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_STATIC}
	c.LoadAssignment = envoy_endpoints.CreateStaticEndpoint(e.Name, e.Address, e.Port)
	return nil
}
