package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	"github.com/kumahq/kuma/pkg/core/xds"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
)

type StrictDNSClusterConfigurer struct {
	Name      string
	Endpoints []xds.Endpoint
}

var _ ClusterConfigurer = &StrictDNSClusterConfigurer{}

func (e *StrictDNSClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STRICT_DNS}
	c.DnsLookupFamily = envoy_cluster.Cluster_V4_ONLY // TODO: make configurable for IPv6
	c.LbPolicy = envoy_cluster.Cluster_ROUND_ROBIN
	c.LoadAssignment = envoy_endpoints.CreateClusterLoadAssignment(e.Name, e.Endpoints)
	return nil
}
