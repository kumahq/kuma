package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	envoy_api_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type CircuitBreakerConfigurer struct {
	CircuitBreaker *mesh_core.CircuitBreakerResource
}

var _ ClusterConfigurer = &CircuitBreakerConfigurer{}

func (c *CircuitBreakerConfigurer) Configure(cluster *envoy_api.Cluster) error {
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
				Priority:           envoy_api_core.RoutingPriority_DEFAULT,
				MaxConnections:     thresholds.GetMaxConnections(),
				MaxPendingRequests: thresholds.GetMaxPendingRequests(),
				MaxRetries:         thresholds.GetMaxRetries(),
				MaxRequests:        thresholds.GetMaxRequests(),
			},
		},
	}
	return nil
}
