package match

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
)

// Routes finds all gateway routes with source selectors that match the
// listener tags.
func Routes(routes []*core_mesh.MeshGatewayRouteResource, listener mesh_proto.TagSelector) []*core_mesh.MeshGatewayRouteResource {
	var matched []*core_mesh.MeshGatewayRouteResource

	for _, route := range routes {
		if _, ok := policy.MatchSelector(listener, route.Selectors()); ok {
			matched = append(matched, route)
		}
	}

	return matched
}
