package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
)

type DnsClusterConfigurer struct {
	Name    string
	Address string
	Port    uint32
	IsHttps bool
}

var _ ClusterConfigurer = &DnsClusterConfigurer{}

func (e *DnsClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = e.Name
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STRICT_DNS}
	c.LbPolicy = envoy_cluster.Cluster_ROUND_ROBIN
	c.LoadAssignment = endpoints.CreateStaticEndpoint(e.Name, e.Address, e.Port)
	if e.IsHttps {
		c.TransportSocket = endpoints.UpgradeClusterTlsUpstream(e.Address)
	}
	return nil
}
