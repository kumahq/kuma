package topology

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/policy"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func GetRetries(
	ctx context.Context,
	dataplane *mesh_core.DataplaneResource,
	destinations core_xds.DestinationMap,
	manager core_manager.ReadOnlyResourceManager,
) (core_xds.RetryMap, error) {
	if len(destinations) == 0 {
		return nil, nil
	}

	retries := &mesh_core.RetryResourceList{}
	if err := manager.List(
		ctx,
		retries,
		core_store.ListByMesh(dataplane.Meta.GetMesh()),
	); err != nil {
		return nil, err
	}

	return BuildRetryMap(dataplane, retries.Items, destinations)
}

func BuildRetryMap(
	dataplane *mesh_core.DataplaneResource,
	retries []*mesh_core.RetryResource,
	destinations core_xds.DestinationMap,
) (core_xds.RetryMap, error) {
	if len(retries) == 0 || len(destinations) == 0 {
		return nil, nil
	}

	policies := make([]policy.ConnectionPolicy, len(retries))
	for i, retry := range retries {
		policies[i] = retry
	}

	policyMap := policy.SelectConnectionPolicies(
		dataplane,
		policy.ToServicesOf(destinations),
		policies,
	)

	retriesMap := core_xds.RetryMap{}
	for service, singlePolicy := range policyMap {
		retriesMap[service] = singlePolicy.(*mesh_core.RetryResource)
	}

	return retriesMap, nil
}
