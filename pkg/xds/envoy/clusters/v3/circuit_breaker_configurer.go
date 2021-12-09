package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type CircuitBreakerConfigurer struct {
	CircuitBreaker *core_mesh.CircuitBreakerResource
}

var _ ClusterConfigurer = &CircuitBreakerConfigurer{}

func (c *CircuitBreakerConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if c.CircuitBreaker == nil {
		return nil
	}
	if !c.CircuitBreaker.HasThresholds() {
		return nil
	}
	thresholds := c.CircuitBreaker.Spec.Conf.GetThresholds()
	cluster.CircuitBreakers = &envoy_cluster.CircuitBreakers{
		Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
			{
				Priority:           envoy_config_core_v3.RoutingPriority_DEFAULT,
				MaxConnections:     thresholds.GetMaxConnections(),
				MaxPendingRequests: thresholds.GetMaxPendingRequests(),
				MaxRetries:         thresholds.GetMaxRetries(),
				MaxRequests:        thresholds.GetMaxRequests(),
			},
		},
	}
	return nil
}
