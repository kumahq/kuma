package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// XXX(jpeach) When we move to GatewayRoute, all this should be replaced
// by the equivalent of `policy.SelectDataplanePolicy` because GatewayRoute
// resources match Dataplanes using the `selectors` field.

// MatchableBySource is an interface that detects all mesh resources
// that have a `sources` selector.
type MatchableBySource interface {
	Sources() []*mesh_proto.Selector
}

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
func Routes(routes model.ResourceList, listener mesh_proto.TagSelector) ([]model.Resource, error) {
	var matched []model.Resource

	for _, i := range routes.GetItems() {
		matchable, ok := i.(MatchableBySource)
		if !ok {
			return nil, nil
		}

		if matches(listener, matchable.Sources()) {
			matched = append(matched, i)
		}
	}

	return matched, nil
}
