package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// matches succeeds if any of the given selectors matches the listener tags.
func matches(listener mesh_proto.TagSelector, selectors []*mesh_proto.Selector) bool {
	for _, selector := range selectors {
		sourceSelector := mesh_proto.TagSelector(selector.GetMatch())
		if sourceSelector.Matches(listener) {
			return true
		}
	}

	return false
}

// Routes finds all the route resources of the given type that
// have a `Sources` selector that matches the given listener tags.
func Routes(routes model.ResourceList, listener mesh_proto.TagSelector) []model.Resource {
	var matched []model.Resource

	for _, i := range routes.GetItems() {
		if c, ok := i.(policy.ConnectionPolicy); ok {
			if matches(listener, c.Sources()) {
				matched = append(matched, i)
			}
			continue
		}
		if c, ok := i.(policy.DataplanePolicy); ok {
			if matches(listener, c.Selectors()) {
				matched = append(matched, i)
			}
			continue
		}
	}

	return matched
}
