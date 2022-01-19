package match

import (
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// Gateway selects the matching GatewayResource (if any) for the given
// TagMatcher.
func Gateway(gatewayList *core_mesh.GatewayResourceList, tagMatcher policy.TagMatcher) *core_mesh.GatewayResource {
	candidates := make([]policy.DataplanePolicy, len(gatewayList.Items))
	for i, gw := range gatewayList.Items {
		candidates[i] = gw
	}

	if p := policy.SelectDataplanePolicyWithMatcher(tagMatcher, candidates); p != nil {
		return p.(*core_mesh.GatewayResource)
	}

	return nil
}
