package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/kumahq/kuma/pkg/core/xds"

	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
)

func StrictDNSCluster(name string, endpoints []xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&StricDNSClusterConfigurer{
			name:      name,
			endpoints: endpoints,
		})
		config.Add(&altStatNameConfigurer{})
		config.Add(&timeoutConfigurer{})
	})
}

type StricDNSClusterConfigurer struct {
	name      string
	endpoints []xds.Endpoint
}

func (e *StricDNSClusterConfigurer) Configure(c *envoy_api.Cluster) error {
	c.Name = e.name
	c.ClusterDiscoveryType = &envoy_api.Cluster_Type{Type: envoy_api.Cluster_STRICT_DNS}
	c.LbPolicy = envoy_api.Cluster_ROUND_ROBIN
	c.LoadAssignment = envoy_endpoints.CreateClusterLoadAssignment(e.name, e.endpoints)
	return nil
}
