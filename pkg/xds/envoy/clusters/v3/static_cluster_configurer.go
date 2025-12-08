package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
	envoy_endpoints "github.com/kumahq/kuma/v2/pkg/xds/envoy/endpoints/v3"
)

type StaticClusterConfigurer struct {
	Name      string
	Endpoints []xds.Endpoint
}

var _ ClusterConfigurer = &StaticClusterConfigurer{}

func (s *StaticClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = s.Name
	c.AltStatName = s.Name
	if len(s.Endpoints) > 0 {
		c.LoadAssignment = envoy_endpoints.CreateClusterLoadAssignment(s.Name, s.Endpoints)
	}
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STATIC}
	return nil
}
