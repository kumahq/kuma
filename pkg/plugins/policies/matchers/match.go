package matchers

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(model.ResourceType, *core_mesh.DataplaneResource, xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	// TODO @lobkovilya Implement standard matching strategy
	return core_xds.TypedMatchingPolicies{}, nil
}
