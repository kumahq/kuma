package dns

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type voMatch struct {
	rank         mesh_proto.TagSelectorRank
	bestSelector mesh_proto.TagSelector
	vob          *core_mesh.VirtualOutboundResource
}

func Match(vos []*core_mesh.VirtualOutboundResource, tags map[string]string) []*core_mesh.VirtualOutboundResource {
	var matchingVirtualOutbounds []voMatch

	for _, vo := range vos {
		bestSelector := mesh_proto.TagSelector{}
		bestRank := mesh_proto.TagSelectorRank{}
		for _, selector := range vo.Selectors() {
			tagSelector := mesh_proto.TagSelector(selector.Match)
			if tagSelector.Matches(tags) {
				r := tagSelector.Rank()
				if bestRank.CompareTo(r) < 0 {
					bestRank = r
					bestSelector = tagSelector
				}
			}
		}
		// If we don't have a bestSelector it means we don't match
		if len(bestSelector) > 0 {
			matchingVirtualOutbounds = append(matchingVirtualOutbounds, voMatch{bestSelector: bestSelector, rank: bestRank, vob: vo})
		}
	}

	sort.Slice(matchingVirtualOutbounds, func(i, j int) bool {
		c := matchingVirtualOutbounds[i].rank.CompareTo(matchingVirtualOutbounds[j].rank)
		if c == 0 {
			// return the youngest policy first
			return matchingVirtualOutbounds[i].vob.Meta.GetCreationTime().After(matchingVirtualOutbounds[j].vob.Meta.GetCreationTime())
		}
		return c > 0
	})

	out := make([]*core_mesh.VirtualOutboundResource, len(matchingVirtualOutbounds))
	for i := range matchingVirtualOutbounds {
		out[i] = matchingVirtualOutbounds[i].vob
	}

	return out
}
