package match

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
)

// Gateway selects the matching GatewayResource (if any) for the given
// TagMatcher.
func Gateway(m manager.ReadOnlyResourceManager, tagMatcher policy.TagMatcher) *core_mesh.GatewayResource {
	gatewayList := &core_mesh.GatewayResourceList{}

	if err := m.List(context.Background(), gatewayList); err != nil {
		return nil
	}

	candidates := make([]policy.DataplanePolicy, len(gatewayList.Items))
	for i, gw := range gatewayList.Items {
		candidates[i] = gw
	}

	if p := policy.SelectDataplanePolicyWithMatcher(tagMatcher, candidates); p != nil {
		return p.(*core_mesh.GatewayResource)
	}

	return nil
}
