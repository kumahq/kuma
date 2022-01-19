package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

// MergeSelectors merges the given tags in order.
func MergeSelectors(tags ...mesh_proto.TagSelector) mesh_proto.TagSelector {
	merged := mesh_proto.TagSelector{}

	for _, t := range tags {
		for k, v := range t {
			merged[k] = v
		}
	}

	return merged
}
