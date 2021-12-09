package topology

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// GetHealthChecks resolves all HealthChecks applicable to a given Dataplane.
func GetHealthChecks(ctx context.Context, dataplane *core_mesh.DataplaneResource, destinations core_xds.DestinationMap, manager core_manager.ReadOnlyResourceManager) (core_xds.HealthCheckMap, error) {
	if len(destinations) == 0 {
		return nil, nil
	}
	healthChecks := &core_mesh.HealthCheckResourceList{}
	if err := manager.List(ctx, healthChecks, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, err
	}
	return BuildHealthCheckMap(dataplane, destinations, healthChecks.Items), nil
}

// BuildHealthCheckMap creates a map with health-checking configuration per reachable service.
func BuildHealthCheckMap(dataplane *core_mesh.DataplaneResource, destinations core_xds.DestinationMap, healthChecks []*core_mesh.HealthCheckResource) core_xds.HealthCheckMap {
	if len(destinations) == 0 || len(healthChecks) == 0 {
		return nil
	}
	policies := make([]policy.ConnectionPolicy, len(healthChecks))
	for i, healthCheck := range healthChecks {
		policies[i] = healthCheck
	}

	policyMap := policy.SelectConnectionPolicies(dataplane, policy.ToServicesOf(destinations), policies)

	healthCheckMap := core_xds.HealthCheckMap{}
	for service, policy := range policyMap {
		healthCheckMap[service] = policy.(*core_mesh.HealthCheckResource)
	}
	return healthCheckMap
}
