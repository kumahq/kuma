package policy

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

// MatchSelector succeeds if any of the given selectors matches the tags. It
// additionally returns the rank of the selector that matched the tag.
func MatchSelector(tags map[string]string, selectors []*mesh_proto.Selector) (mesh_proto.TagSelectorRank, bool) {
	for _, selector := range selectors {
		sourceSelector := mesh_proto.TagSelector(selector.GetMatch())
		if sourceSelector.Matches(tags) {
			return sourceSelector.Rank(), true
		}
	}

	return mesh_proto.TagSelectorRank{}, false
}
