package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
)

// ClusterConfigurer is responsible for configuring a single aspect of the entire Envoy cluster,
// such as filter chain, transparent proxying, etc.
type ClusterConfigurer interface {
	// Configure configures a single aspect on a given Envoy cluster.
	Configure(cluster *envoy_cluster.Cluster) error
}
