package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v2"
)

type DnsClusterConfigurer struct {
	Name    string
	Address string
	Port    uint32
}

var _ ClusterConfigurer = &DnsClusterConfigurer{}

func (e *DnsClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_STRICT_DNS}
	c.LbPolicy = envoy_api.Cluster_ROUND_ROBIN
	c.LoadAssignment = envoy_endpoints.CreateStaticEndpoint(e.Name, e.Address, e.Port)
	return nil
}
