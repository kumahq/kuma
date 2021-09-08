package match

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type gatewayPolicyAdaptor struct {
	*core_mesh.GatewayResource
}

var _ policy.DataplanePolicy = gatewayPolicyAdaptor{}

func (g gatewayPolicyAdaptor) Selectors() []*mesh_proto.Selector {
	return g.Sources()
}

// Gateway selects the matching GatewayResource (if any) for the given DataplaneResorce.
func Gateway(m manager.ReadOnlyResourceManager, dp *core_mesh.DataplaneResource) *core_mesh.GatewayResource {
	gatewayList := &core_mesh.GatewayResourceList{}

	if err := m.List(context.Background(), gatewayList, store.ListByMesh(dp.Meta.GetMesh())); err != nil {
		return nil
	}

	candidates := make([]policy.DataplanePolicy, len(gatewayList.Items))
	for i, gw := range gatewayList.Items {
		candidates[i] = gatewayPolicyAdaptor{gw}
	}

	if p := policy.SelectDataplanePolicy(dp, candidates); p != nil {
		return p.(gatewayPolicyAdaptor).GatewayResource
	}

	return nil
}
