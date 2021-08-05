package match

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

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
func Routes(m manager.ReadOnlyResourceManager, routeType model.ResourceType, listener mesh_proto.TagSelector) ([]model.Resource, error) {
	list, err := registry.Global().NewList(routeType)
	if err != nil {
		return nil, err
	}

	if err := m.List(context.Background(), list); err != nil {
		return nil, err
	}

	var matched []model.Resource

	for _, i := range list.GetItems() {
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
