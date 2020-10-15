package topology

import (
	"context"

	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func GetTrafficTrace(ctx context.Context, dataplane *mesh_core.DataplaneResource, manager core_manager.ReadOnlyResourceManager) (*mesh_core.TrafficTraceResource, error) {
	traces := mesh_core.TrafficTraceResourceList{}
	if err := manager.List(ctx, &traces, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}
	return SelectTrafficTrace(dataplane, traces.Items), nil
}

func SelectTrafficTrace(dataplane *mesh_core.DataplaneResource, traces []*mesh_core.TrafficTraceResource) *mesh_core.TrafficTraceResource {
	policies := make([]core_policy.DataplanePolicy, len(traces))
	for i, trace := range traces {
		policies[i] = trace
	}
	if policy := core_policy.SelectDataplanePolicy(dataplane, policies); policy != nil {
		return policy.(*mesh_core.TrafficTraceResource)
	}
	return nil
}
