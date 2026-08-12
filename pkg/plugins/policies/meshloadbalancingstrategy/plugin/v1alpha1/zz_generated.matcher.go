package v1alpha1

import (
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshLoadBalancingStrategyType, dataplane, resources, opts...)
}
