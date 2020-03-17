package topology

import (
	"context"

	core_policy "github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func GetTrafficTrace(ctx context.Context, dataplane *mesh_core.DataplaneResource, manager core_manager.ReadOnlyResourceManager) (*mesh_core.TrafficTraceResource, error) {
	list := mesh_core.TrafficTraceResourceList{}
	if err := manager.List(ctx, &list, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}

	policies := make([]core_policy.DataplanePolicy, len(list.Items))
	for i, trace := range list.Items {
		policies[i] = trace
	}

	if policy := core_policy.SelectDataplanePolicy(dataplane, policies); policy != nil {
		return policy.(*mesh_core.TrafficTraceResource), nil
	}
	return nil, nil
}
