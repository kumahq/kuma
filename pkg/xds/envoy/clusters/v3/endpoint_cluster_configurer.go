package clusters

import (
	"net"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/xds"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
)

type ProvidedEndpointClusterConfigurer struct {
	Name      string
	Endpoints []xds.Endpoint
	HasIPv6   bool
}

var _ ClusterConfigurer = &ProvidedEndpointClusterConfigurer{}

func (e *ProvidedEndpointClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	if len(e.Endpoints) == 0 {
		return errors.New("cluster must have at least 1 endpoint")
	}
	c.Name = e.Name
	if len(e.Endpoints) > 1 {
		c.LbPolicy = envoy_cluster.Cluster_ROUND_ROBIN
	}
	var nonIpEndpoints []xds.Endpoint
	var ipEndpoints []xds.Endpoint
	for _, endpoint := range e.Endpoints {
		if net.ParseIP(endpoint.Target) != nil || endpoint.UnixDomainPath != "" {
			ipEndpoints = append(ipEndpoints, endpoint)
		} else {
			nonIpEndpoints = append(nonIpEndpoints, endpoint)
		}
	}
	if len(nonIpEndpoints) > 0 && len(ipEndpoints) > 0 {
		return errors.New("cluster is a mix of ips and hostnames, can't generate envoy config.")
	}
	if len(nonIpEndpoints) > 0 {
		c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STRICT_DNS}
		if e.HasIPv6 {
			c.DnsLookupFamily = envoy_cluster.Cluster_AUTO
		} else {
			c.DnsLookupFamily = envoy_cluster.Cluster_V4_ONLY
		}
	} else {
		c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STATIC}
	}
	c.LoadAssignment = envoy_endpoints.CreateClusterLoadAssignment(e.Name, e.Endpoints)
	return nil
}
