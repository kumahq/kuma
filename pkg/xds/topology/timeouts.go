package topology

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// GetTimeouts picks a single the most specific timeout for each outbound interface of a given Dataplane.
func GetTimeouts(
	ctx context.Context,
	dataplane *core_mesh.DataplaneResource,
	manager core_manager.ReadOnlyResourceManager,
) (core_xds.TimeoutMap, error) {
	if len(dataplane.Spec.Networking.GetOutbound()) == 0 {
		return nil, nil
	}
	timeouts := &core_mesh.TimeoutResourceList{}
	if err := manager.List(ctx, timeouts, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, err
	}
	return BuildTimeoutMap(dataplane, timeouts.Items), nil
}

// BuildTimeoutMap picks a single the most specific timeout for each outbound interface of a given Dataplane.
func BuildTimeoutMap(dataplane *core_mesh.DataplaneResource, timeouts []*core_mesh.TimeoutResource) core_xds.TimeoutMap {
	policies := make([]core_policy.ConnectionPolicy, len(timeouts))
	for i, timeout := range timeouts {
		policies[i] = timeout
	}
	policyMap := core_policy.SelectOutboundConnectionPolicies(dataplane, policies)

	timeoutMap := core_xds.TimeoutMap{}
	for _, oface := range dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter) {
		serviceName := oface.GetService()
		if policy, exists := policyMap[serviceName]; exists {
			outbound := dataplane.Spec.Networking.ToOutboundInterface(oface)
			timeoutMap[outbound] = policy.(*core_mesh.TimeoutResource)
		}
	}
	return timeoutMap
}
