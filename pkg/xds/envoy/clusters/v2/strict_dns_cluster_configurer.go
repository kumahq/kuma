package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/kumahq/kuma/pkg/core/xds"

	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v2"
)

type StrictDNSClusterConfigurer struct {
	Name      string
	Endpoints []xds.Endpoint
}

var _ ClusterConfigurer = &StrictDNSClusterConfigurer{}

func (e *StrictDNSClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_STRICT_DNS}
	c.DnsLookupFamily = envoy_api.Cluster_V4_ONLY // TODO: make configurable for IPv6
	c.LbPolicy = envoy_api.Cluster_ROUND_ROBIN
	c.LoadAssignment = envoy_endpoints.CreateClusterLoadAssignment(e.Name, e.Endpoints)
	return nil
}
