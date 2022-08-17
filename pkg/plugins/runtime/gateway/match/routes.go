package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// Routes finds all the route resources of the given type that
// have a `Sources` selector that matches the given listener tags.
func Routes(routes []*core_mesh.MeshGatewayRouteResource, listener mesh_proto.TagSelector) []*core_mesh.MeshGatewayRouteResource {
	var matched []*core_mesh.MeshGatewayRouteResource

	for _, i := range routes {
		if _, ok := policy.MatchSelector(listener, i.Selectors()); ok {
			matched = append(matched, i)
		}
		continue
	}

	return matched
}
