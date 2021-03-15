package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type CircuitBreakerConfigurer struct {
	CircuitBreaker *mesh_core.CircuitBreakerResource
}

var _ ClusterConfigurer = &CircuitBreakerConfigurer{}

func (c *CircuitBreakerConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if c.CircuitBreaker == nil {
		return nil
	}
	if !c.CircuitBreaker.HasThresholds() {
		return nil
	}
	cluster.CircuitBreakers = &envoy_cluster.CircuitBreakers{
		Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
			{
				Priority:           envoy_config_core_v3.RoutingPriority_DEFAULT,
				MaxConnections:     c.CircuitBreaker.Spec.Conf.GetThresholds().GetMaxConnections(),
				MaxPendingRequests: c.CircuitBreaker.Spec.Conf.GetThresholds().GetMaxPendingRequests(),
				MaxRetries:         c.CircuitBreaker.Spec.Conf.GetThresholds().GetMaxRetries(),
				MaxRequests:        c.CircuitBreaker.Spec.Conf.GetThresholds().GetMaxRequests(),
			},
		},
	}
	return nil
}
