package topology

import (
	"context"
	core_policy "github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func GetTrafficLog(ctx context.Context, dataplane *mesh_core.DataplaneResource, manager core_manager.ResourceManager) (*mesh_core.TrafficLogResource, error) {
	list := mesh_core.TrafficLogResourceList{}
	if err := manager.List(ctx, &list, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}

	policies := make([]core_policy.DataplanePolicy, len(list.Items))
	for i, log := range list.Items {
		policies[i] = log
	}

	if policy := core_policy.SelectDataplanePolicy(dataplane, policies); policy != nil {
		return policy.(*mesh_core.TrafficLogResource), nil
	}
	return nil, nil
}
