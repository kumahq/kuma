package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type BlackHoleClusterConfigurer struct{}

var _ ClusterConfigurer = &BlackHoleClusterConfigurer{}

func (p *BlackHoleClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.ClusterDiscoveryType = &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_STATIC}
	c.ConnectTimeout = util_proto.Duration(policies_defaults.DefaultConnectTimeout)
	return nil
}
