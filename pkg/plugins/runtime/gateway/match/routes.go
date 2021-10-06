package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// Routes finds all the route resources of the given type that
// have a `Sources` selector that matches the given listener tags.
func Routes(routes model.ResourceList, listener mesh_proto.TagSelector) []model.Resource {
	var matched []model.Resource

	for _, i := range routes.GetItems() {
		if c, ok := i.(policy.ConnectionPolicy); ok {
			if _, ok := policy.MatchSelector(listener, c.Sources()); ok {
				matched = append(matched, i)
			}
			continue
		}
		if c, ok := i.(policy.DataplanePolicy); ok {
			if _, ok := policy.MatchSelector(listener, c.Selectors()); ok {
				matched = append(matched, i)
			}
			continue
		}
	}

	return matched
}
