package topology

import (
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func SelectGateway(gatewayList *core_mesh.MeshGatewayResourceList, tagMatcher core_policy.TagMatcher) *core_mesh.MeshGatewayResource {
	candidates := make([]core_policy.DataplanePolicy, len(gatewayList.Items))
	for i, gw := range gatewayList.Items {
		candidates[i] = gw
	}

	if p := core_policy.SelectDataplanePolicyWithMatcher(tagMatcher, candidates); p != nil {
		return p.(*core_mesh.MeshGatewayResource)
	}

	return nil
}
