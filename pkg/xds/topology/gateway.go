package topology

import (
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func SelectGateway(gateways []*core_mesh.MeshGatewayResource, tagMatcher core_policy.TagMatcher) *core_mesh.MeshGatewayResource {
	candidates := make([]core_policy.DataplanePolicy, len(gateways))
	for i, gw := range gateways {
		candidates[i] = gw
	}

	if p := core_policy.SelectDataplanePolicyWithMatcher(tagMatcher, candidates); p != nil {
		return p.(*core_mesh.MeshGatewayResource)
	}

	return nil
}
