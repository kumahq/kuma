package topology

import (
	"context"

	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func GetTrafficTrace(ctx context.Context, dataplane *core_mesh.DataplaneResource, manager core_manager.ReadOnlyResourceManager) (*core_mesh.TrafficTraceResource, error) {
	traces := core_mesh.TrafficTraceResourceList{}
	if err := manager.List(ctx, &traces, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}
	return SelectTrafficTrace(dataplane, traces.Items), nil
}

func SelectTrafficTrace(dataplane *core_mesh.DataplaneResource, traces []*core_mesh.TrafficTraceResource) *core_mesh.TrafficTraceResource {
	policies := make([]core_policy.DataplanePolicy, len(traces))
	for i, trace := range traces {
		policies[i] = trace
	}
	if policy := core_policy.SelectDataplanePolicy(dataplane, policies); policy != nil {
		return policy.(*core_mesh.TrafficTraceResource)
	}
	return nil
}
