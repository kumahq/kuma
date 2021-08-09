package topology

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// GetCircuitBreakers resolves all CircuitBreakers applicable to a given Dataplane.
func GetCircuitBreakers(ctx context.Context, dataplane *core_mesh.DataplaneResource, destinations core_xds.DestinationMap, manager core_manager.ReadOnlyResourceManager) (core_xds.CircuitBreakerMap, error) {
	if len(destinations) == 0 {
		return nil, nil
	}
	circuitBreakers := &core_mesh.CircuitBreakerResourceList{}
	if err := manager.List(ctx, circuitBreakers, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, err
	}
	return BuildCircuitBreakerMap(dataplane, destinations, circuitBreakers.Items), nil
}

// BuildCircuitBreakerMap creates a map with circuit-breaking configuration per reachable service.
func BuildCircuitBreakerMap(dataplane *core_mesh.DataplaneResource, destinations core_xds.DestinationMap, circuitBreakers []*core_mesh.CircuitBreakerResource) core_xds.CircuitBreakerMap {
	if len(destinations) == 0 || len(circuitBreakers) == 0 {
		return nil
	}
	policies := make([]policy.ConnectionPolicy, len(circuitBreakers))
	for i, circuitBreaker := range circuitBreakers {
		policies[i] = circuitBreaker
	}

	policyMap := policy.SelectConnectionPolicies(dataplane, policy.ToServicesOf(destinations), policies)

	circuitBreakerMap := core_xds.CircuitBreakerMap{}
	for service, policy := range policyMap {
		circuitBreakerMap[service] = policy.(*core_mesh.CircuitBreakerResource)
	}
	return circuitBreakerMap
}
