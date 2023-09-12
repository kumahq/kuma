package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_cluster_aggregate "github.com/envoyproxy/go-control-plane/envoy/extensions/clusters/aggregate/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type AggregateClusterConfigurer struct {
	Clusters []string
}

var _ ClusterConfigurer = &AggregateClusterConfigurer{}

func (e *AggregateClusterConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.LbPolicy = envoy_cluster.Cluster_CLUSTER_PROVIDED

	cfg, err := proto.MarshalAnyDeterministic(&envoy_cluster_aggregate.ClusterConfig{
		Clusters: e.Clusters,
	})
	if err != nil {
		return err
	}

	c.ClusterDiscoveryType = &envoy_cluster.Cluster_ClusterType{
		ClusterType: &envoy_cluster.Cluster_CustomClusterType{
			Name:        "envoy.clusters.aggregate",
			TypedConfig: cfg,
		},
	}

	return nil
}
